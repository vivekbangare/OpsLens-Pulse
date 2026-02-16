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
	"opslense-pulse/agent/retry"
	"opslense-pulse/agent/sender"
)

var (
	lastMetrics     = make(map[string]time.Time) // containerID -> last sent time
	lastMetricsFile = filepath.Join(BaseDir, "container_metrics.json")
	lastMetricsLock = &sync.Mutex{}
	metricsDirty    = false
	flushInterval   = 30 * time.Second
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

	if !metricsDirty {
		return
	}

	data, err := json.Marshal(lastMetrics)
	if err != nil {
		log.Println("Error marshaling container metrics data:", err)
		return
	}

	tmp := lastMetricsFile + ".tmp"

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		log.Println("Error writing temp container metrics file:", err)
		return
	}

	if err := os.Rename(tmp, lastMetricsFile); err != nil {
		log.Println("Error renaming container metrics file:", err)
		return
	}

	metricsDirty = false
}

// periodicMetricsFlush runs background flush
func periodicMetricsFlush(ctx context.Context) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			saveLastMetrics() // final flush
			return
		case <-ticker.C:
			saveLastMetrics()
		}
	}
}

// StartContainerMetricsCollector collects metrics from running Docker containers
func StartContainerMetricsCollector(
	ctx context.Context,
	agentID, hostname, serverURL, apiKey string,
	intervalSeconds int,
) {
	loadLastMetrics()

	// Start background flusher
	go periodicMetricsFlush(ctx)

	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("📦 Container metrics collector stopped")
			saveLastMetrics() // final safety flush
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

				if !lastSent.IsZero() && !raw.Timestamp.After(lastSent) {
					continue
				}

				metric := containers.ToSharedMetrics(
					raw,
					agentID,
					hostname,
				)

				if err := retry.Do(3, 2*time.Second, func() error {
					return sender.SendContainerMetrics(serverURL, apiKey, metric)
				}); err != nil {
					log.Println("Container metrics send failed:", err)
					continue
				}

				// Update in-memory only
				lastMetricsLock.Lock()
				lastMetrics[c.ID] = raw.Timestamp
				metricsDirty = true
				lastMetricsLock.Unlock()
			}
		}
	}
}
