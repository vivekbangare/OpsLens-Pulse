package collector

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"opslense-pulse/agent/containers"
	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

var (
	containerLastLine     = make(map[string]int)
	containerLastLineFile = filepath.Join(BaseDir, "container_last_line.json")
	containerLastLineLock = &sync.Mutex{}
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

// saveLastLines persists positions
func saveLastLines() {
	containerLastLineLock.Lock()
	defer containerLastLineLock.Unlock()

	data, err := json.Marshal(containerLastLine)
	if err != nil {
		log.Println("Error marshaling container log cursor:", err)
		return
	}

	if err := os.WriteFile(containerLastLineFile, data, 0644); err != nil {
		log.Println("Error writing container log cursor:", err)
	}
}

// StartContainerLogsCollector collects logs from Docker containers
func StartContainerLogsCollector(
	ctx context.Context,
	agentID, accountID, hostname, serverURL, apiKey string,
	intervalSeconds int,
) {
	loadLastLines()

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
				// Extract container name safely
				name := ""
				if len(c.Names) > 0 {
					name = strings.TrimPrefix(c.Names[0], "/")
				}

				// Build entries using root shared.ContainerLog
				var entries []shared.ContainerLog
				for _, line := range newLines {
					entries = append(entries, shared.ContainerLog{
						AccountID:     accountID,
						AgentID:       agentID,
						Hostname:      hostname,
						ContainerID:   c.ID,
						ContainerName: name,
						Timestamp:     time.Now().Unix(),
						Message:       line,
						Level:         "info", // default level
					})
				}

				if len(entries) == 0 {
					continue
				}

				batch := shared.ContainerLogBatch{
					AccountID: accountID,
					AgentID:   agentID,
					Hostname:  hostname,
					Logs:      entries,
				}

				if err := retrySend(3, 2*time.Second, func() error {
					return sender.SendContainerLogs(serverURL, apiKey, batch)
				}); err != nil {
					log.Println("Container logs send failed:", err)
					continue
				}

				containerLastLineLock.Lock()
				containerLastLine[c.ID] = len(lines)
				containerLastLineLock.Unlock()

				saveLastLines()
			}
		}
	}
}
