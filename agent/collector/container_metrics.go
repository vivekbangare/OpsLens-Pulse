package collector

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"opslense-pulse/agent/containers"
	"opslense-pulse/agent/sender"
)

var (
	lastMetrics     = make(map[string]time.Time) // containerID -> last sent time
	lastMetricsFile = filepath.Join(BaseDir, "container_metrics.json")
	lastMetricsLock = &sync.Mutex{}
)

// loadLastMetrics loads last sent timestamps from disk
func loadLastMetrics() {
	data, err := os.ReadFile(lastMetricsFile)
	if err != nil {
		return
	}
	lastMetricsLock.Lock()
	defer lastMetricsLock.Unlock()
	_ = json.Unmarshal(data, &lastMetrics)
}

// saveLastMetrics persists timestamps to disk
func saveLastMetrics() {
	lastMetricsLock.Lock()
	defer lastMetricsLock.Unlock()

	data, err := json.Marshal(lastMetrics)
	if err != nil {
		log.Println("Error marshaling container metrics data:", err)
		return
	}

	if err := os.WriteFile(lastMetricsFile, data, 0644); err != nil {
		log.Println("Error writing container metrics file:", err)
	}
}

// StartContainerMetricsCollector collects metrics from running Docker containers
func StartContainerMetricsCollector(
	ctx context.Context,
	agentID, accountID, hostname, serverURL, apiKey string,
	intervalSeconds int,
) {
	loadLastMetrics()

	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("📦 Container metrics collector stopped")
			saveLastMetrics()
			return

		case <-ticker.C:
			list, err := containers.ListRunning()
			if err != nil {
				log.Println("Error listing containers:", err)
				continue
			}

			for _, c := range list {
				raw, err := containers.GetContainerMetrics(c)
				if err != nil {
					continue
				}

				lastMetricsLock.Lock()
				lastSent := lastMetrics[c.ID]
				lastMetricsLock.Unlock()

				// Skip if nothing new (defensive, metrics usually always advance)
				if !lastSent.IsZero() && !raw.Timestamp.After(lastSent) {
					continue
				}

				metric := containers.ToSharedMetrics(
					raw,
					accountID,
					agentID,
					hostname,
				)

				if err := retrySend(3, 2*time.Second, func() error {
					return sender.SendContainerMetrics(serverURL, apiKey, metric)
				}); err != nil {
					log.Println("Container metrics send failed:", err)
					continue
				}

				lastMetricsLock.Lock()
				lastMetrics[c.ID] = raw.Timestamp
				lastMetricsLock.Unlock()

				saveLastMetrics()
			}
		}
	}
}

// retrySend retries the provided function with exponential backoff
// func retrySend(attempts int, baseDelay time.Duration, fn func() error) error {
// 	delay := baseDelay
// 	for i := 0; i < attempts; i++ {
// 		if err := fn(); err != nil {
// 			time.Sleep(delay)
// 			delay *= 2
// 		} else {
// 			return nil
// 		}
// 	}
// 	return nil
// }
