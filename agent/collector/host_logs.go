package collector

import (
	"context"
	"strings"
	"time"

	"opslense-pulse/agent/batcher"
	"opslense-pulse/agent/config"
	"opslense-pulse/agent/limiter"
	"opslense-pulse/agent/queue"
	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"

	"github.com/nxadm/tail"
)

func startFileCollector(
	ctx context.Context,
	agentID, hostname string,
	src config.LogSource,
	logsQueue *queue.FileQueue,
) {

	shared.Info("Tail log collector started", "path", src.Path)

	t, err := tail.TailFile(src.Path, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Poll:      true,
	})
	if err != nil {
		shared.Error("Tail error", "error", err.Error())
		return
	}

	// ✅ Batcher (200 logs OR 1 second flush)
	logBatcher := batcher.New(200, 1*time.Second)

	for {
		select {

		case <-ctx.Done():
			t.Cleanup()
			return

		case line := <-t.Lines:
			if line == nil {
				continue
			}
			if !limiter.Allow() {
				continue
			}

			level := ExtractLevel(line.Text)

			entry := sender.LogEntry{
				AgentID:   agentID,
				Hostname:  hostname,
				Timestamp: time.Now().Unix(),
				Level:     level,
				Message:   strings.TrimSpace(line.Text),
				Tags: map[string]string{
					"source": src.Name,
					"type":   "file",
				},
			}

			out := logBatcher.Add(entry)

			if out != nil {
				var entries []sender.LogEntry
				for _, e := range out {
					entries = append(entries, e.(sender.LogEntry))
				}

				batch := sender.LogBatch{
					AgentID:  agentID,
					Hostname: hostname,
					Logs:     entries,
				}

				_ = logsQueue.Enqueue(batch)
			}

		}
	}
}
