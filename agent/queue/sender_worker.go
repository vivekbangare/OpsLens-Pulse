package queue

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"time"

	"opslense-pulse/agent/internal"
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

			tmpPath := q.path + ".tmp"
			tmpFile, err := os.Create(tmpPath)
			if err != nil {
				file.Close()
				q.mu.Unlock()
				continue
			}

			scanner := bufio.NewScanner(file)
			scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

			for scanner.Scan() {

				line := scanner.Bytes()

				var m shared.HostMetrics
				if err := json.Unmarshal(line, &m); err != nil {
					continue
				}

				// Circuit open → just rewrite
				if internal.IsCircuitOpen() {
					tmpFile.Write(line)
					tmpFile.Write([]byte("\n"))
					continue
				}

				err := retry.Do(3, 2*time.Second, func() error {
					return sender.SendRaw(serverURL, apiKey, m)
				})

				if err != nil {

					internal.IncFailure()

					if internal.GetFailures() > 10 {
						internal.OpenCircuit(30 * time.Second)
					}

					tmpFile.Write(line)
					tmpFile.Write([]byte("\n"))

				} else {
					internal.ResetFailure()
				}
			}

			file.Close()
			tmpFile.Close()

			os.Rename(tmpPath, q.path)

			q.mu.Unlock()
		}
	}
}
