package collector

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"opslense-pulse/agent/containers"
	"opslense-pulse/agent/retry"
	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

var (
	lastMetrics     = make(map[string]time.Time)
	lastMetricsFile = filepath.Join(StateDir, "container_metrics.json")
	lastMetricsLock = &sync.Mutex{}
	metricsDirty    = false
	flushInterval   = 30 * time.Second
)

func loadLastMetrics() {
	data, err := os.ReadFile(lastMetricsFile)
	if err != nil {
		return
	}
	lastMetricsLock.Lock()
	defer lastMetricsLock.Unlock()
	_ = json.Unmarshal(data, &lastMetrics)
}

func saveLastMetrics() {
	lastMetricsLock.Lock()
	defer lastMetricsLock.Unlock()

	if !metricsDirty {
		return
	}

	data, err := json.Marshal(lastMetrics)
	if err != nil {
		shared.Error("container metrics marshal failed", "error", err.Error())
		return
	}

	tmp := lastMetricsFile + ".tmp"

	if err := os.WriteFile(tmp, data, 0600); err != nil {
		shared.Error("container metrics temp write failed", "error", err.Error())
		return
	}

	if err := os.Rename(tmp, lastMetricsFile); err != nil {
		shared.Error("container metrics rename failed", "error", err.Error())
		return
	}

	metricsDirty = false
}

func periodicMetricsFlush(ctx context.Context) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			saveLastMetrics()
			return
		case <-ticker.C:
			saveLastMetrics()
		}
	}
}

func StartContainerMetricsCollector(
	ctx context.Context,
	agentID, hostname, serverURL, apiKey string,
	intervalSeconds int,
) {

	shared.Info("container metrics collector started")

	loadLastMetrics()
	go periodicMetricsFlush(ctx)

	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			saveLastMetrics()
			shared.Info("container metrics collector stopped")
			return

		case <-ticker.C:

			list, err := containers.ListRunning(ctx)
			if err != nil {
				continue
			}

			for _, c := range list {

				raw, err := containers.GetContainerMetrics(ctx, c)
				if err != nil {
					continue
				}

				lastMetricsLock.Lock()
				lastSent := lastMetrics[c.ID]
				lastMetricsLock.Unlock()

				if !lastSent.IsZero() && !raw.Timestamp.After(lastSent) {
					continue
				}

				metric := containers.ToSharedMetrics(raw, agentID, hostname)

				if err := retry.Do(3, 2*time.Second, func() error {
					return sender.SendContainerMetrics(serverURL, apiKey, metric)
				}); err != nil {
					shared.Error("container metrics send failed", "error", err.Error())
					continue
				}

				lastMetricsLock.Lock()
				lastMetrics[c.ID] = raw.Timestamp
				metricsDirty = true
				lastMetricsLock.Unlock()
			}
		}
	}
}
