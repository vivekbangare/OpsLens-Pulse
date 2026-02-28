package collector

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"opslense-pulse/agent/containers"
	"opslense-pulse/agent/limiter"
	"opslense-pulse/agent/queue"
	"opslense-pulse/shared"
)

var (
	containerLastTime     = make(map[string]int64)
	containerLastLineFile = filepath.Join(StateDir, "container_last_time.json")
	containerLastLineLock = &sync.Mutex{}
	logsDirty             = false
	logsFlushEvery        = 30 * time.Second
)

func loadLastLines() {
	data, err := os.ReadFile(containerLastLineFile)
	if err != nil {
		return
	}
	containerLastLineLock.Lock()
	defer containerLastLineLock.Unlock()
	_ = json.Unmarshal(data, &containerLastTime)
}

func saveLastLines() {
	containerLastLineLock.Lock()
	defer containerLastLineLock.Unlock()

	if !logsDirty {
		return
	}

	data, err := json.Marshal(containerLastTime)
	if err != nil {
		shared.Error("container log cursor marshal failed", "error", err.Error())
		return
	}

	tmp := containerLastLineFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		shared.Error("container log cursor temp write failed", "error", err.Error())
		return
	}

	if err := os.Rename(tmp, containerLastLineFile); err != nil {
		shared.Error("container log cursor rename failed", "error", err.Error())
		return
	}

	logsDirty = false
}

func periodicLogsFlush(ctx context.Context) {
	ticker := time.NewTicker(logsFlushEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			saveLastLines()
			return
		case <-ticker.C:
			saveLastLines()
		}
	}
}

func StartContainerLogsCollector(
	ctx context.Context,
	agentID, hostname string,
	logsQueue *queue.FileQueue,
	intervalSeconds int,
) {

	shared.Info("container logs collector started")

	loadLastLines()
	go periodicLogsFlush(ctx)

	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			saveLastLines()
			shared.Info("container logs collector stopped")
			return

		case <-ticker.C:

			list, err := containers.ListRunning(ctx)
			if err != nil {
				shared.Error("list containers failed", "error", err.Error())
				continue
			}

			for _, c := range list {

				containerLastLineLock.Lock()
				last := containerLastTime[c.ID]
				containerLastLineLock.Unlock()

				lines, err := containers.GetContainerLogsSince(ctx, c.ID, last)
				if err != nil || len(lines) == 0 {
					continue
				}

				name := ""
				if len(c.Names) > 0 {
					name = strings.TrimPrefix(c.Names[0], "/")
				}

				var entries []shared.ContainerLog

				for _, line := range lines {

					if !limiter.Allow() {
						continue
					}

					level := ExtractLevel(line)

					entries = append(entries, shared.ContainerLog{
						AgentID:       agentID,
						Hostname:      hostname,
						ContainerID:   c.ID,
						ContainerName: name,
						Timestamp:     time.Now().Unix(),
						Message:       line,
						Level:         level,
					})
				}

				if len(entries) == 0 {
					continue
				}

				batch := shared.ContainerLogBatch{
					AgentID:  agentID,
					Hostname: hostname,
					Logs:     entries,
				}

				// ✅ Enqueue instead of direct send
				if err := logsQueue.Enqueue(batch); err != nil {
					shared.Error("enqueue container logs failed", "error", err.Error())
					continue
				}

				// Update cursor only after successful enqueue
				containerLastLineLock.Lock()
				containerLastTime[c.ID] = time.Now().Unix()
				logsDirty = true
				containerLastLineLock.Unlock()
			}
		}
	}
}
