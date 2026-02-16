package collector

import (
	"context"
	"encoding/json"
	"log"
	"opslense-pulse/agent/containers"
	"opslense-pulse/agent/retry"
	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	containerLastLine     = make(map[string]int)
	containerLastLineFile = filepath.Join(BaseDir, "container_last_line.json")
	containerLastLineLock = &sync.Mutex{}

	logsDirty      = false
	logsFlushEvery = 30 * time.Second
)

// loadLastLines loads last read positions
func loadLastLines() {
	data, err := os.ReadFile(containerLastLineFile)
	if err != nil {
		return
	}
	containerLastLineLock.Lock()
	defer containerLastLineLock.Unlock()
	_ = json.Unmarshal(data, &containerLastLine)
}

// saveLastLines persists positions (only if dirty)
func saveLastLines() {
	containerLastLineLock.Lock()
	defer containerLastLineLock.Unlock()

	if !logsDirty {
		return
	}

	data, err := json.Marshal(containerLastLine)
	if err != nil {
		log.Println("Error marshaling container log cursor:", err)
		return
	}

	tmp := containerLastLineFile + ".tmp"

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		log.Println("Error writing temp container log cursor:", err)
		return
	}

	if err := os.Rename(tmp, containerLastLineFile); err != nil {
		log.Println("Error renaming container log cursor:", err)
		return
	}

	logsDirty = false
}

// periodicLogsFlush background flusher
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

// StartContainerLogsCollector collects logs from Docker containers
func StartContainerLogsCollector(
	ctx context.Context,
	agentID, hostname, serverURL, apiKey string,
	intervalSeconds int,
) {
	loadLastLines()
	go periodicLogsFlush(ctx)

	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("📦 Container logs collector stopped")
			saveLastLines()
			return

		case <-ticker.C:
			list, err := containers.ListRunning()
			if err != nil {
				log.Println("Error listing containers:", err)
				continue
			}

			for _, c := range list {
				lines, err := containers.GetContainerLogs(c.ID, 100)
				if err != nil || len(lines) == 0 {
					continue
				}

				containerLastLineLock.Lock()
				last := containerLastLine[c.ID]
				containerLastLineLock.Unlock()

				if last >= len(lines) {
					continue
				}

				newLines := lines[last:]

				name := ""
				if len(c.Names) > 0 {
					name = strings.TrimPrefix(c.Names[0], "/")
				}

				var entries []shared.ContainerLog
				for _, line := range newLines {
					entries = append(entries, shared.ContainerLog{
						AgentID:       agentID,
						Hostname:      hostname,
						ContainerID:   c.ID,
						ContainerName: name,
						Timestamp:     time.Now().Unix(),
						Message:       line,
						Level:         "info",
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

				if err := retry.Do(3, 2*time.Second, func() error {
					return sender.SendContainerLogs(serverURL, apiKey, batch)
				}); err != nil {
					log.Println("Container logs send failed:", err)
					continue
				}

				containerLastLineLock.Lock()
				containerLastLine[c.ID] = len(lines)
				logsDirty = true
				containerLastLineLock.Unlock()
			}
		}
	}
}
