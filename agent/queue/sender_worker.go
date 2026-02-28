package queue

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"time"

	"opslense-pulse/agent/metrics"
	"opslense-pulse/agent/retry"
	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

func StartMetricsSender(ctx context.Context, q *FileQueue, serverURL, apiKey string) {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:

			q.mu.Lock()

			file, err := os.Open(q.path)
			if err != nil {
				q.mu.Unlock()
				continue
			}

			var lines [][]byte
			scanner := bufio.NewScanner(file)
			scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

			for scanner.Scan() {
				b := make([]byte, len(scanner.Bytes()))
				copy(b, scanner.Bytes())
				lines = append(lines, b)
			}

			file.Close()
			q.mu.Unlock()

			if len(lines) == 0 {
				continue
			}

			var unsent [][]byte

			for _, line := range lines {

				var m shared.HostMetrics
				if err := json.Unmarshal(line, &m); err != nil {
					continue
				}

				err := retry.Do(3, 2*time.Second, func() error {
					return sender.SendRaw(serverURL, apiKey, m)
				})

				if err != nil {
					metrics.IncMetricsFailure()
					unsent = append(unsent, line)
				}
			}

			q.mu.Lock()

			tmpPath := q.path + ".tmp"
			tmpFile, err := os.Create(tmpPath)
			if err == nil {
				for _, u := range unsent {
					tmpFile.Write(u)
					tmpFile.Write([]byte("\n"))
				}
				tmpFile.Close()
				os.Rename(tmpPath, q.path)
			}

			q.mu.Unlock()
		}
	}
}
