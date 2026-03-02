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
)

func StartLogsSender(ctx context.Context, q *FileQueue, serverURL, apiKey string) {

	ticker := time.NewTicker(3 * time.Second)
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

				var batch sender.LogBatch
				if err := json.Unmarshal(line, &batch); err != nil {
					continue
				}

				if internal.IsCircuitOpen() {
					tmpFile.Write(line)
					tmpFile.Write([]byte("\n"))
					continue
				}

				err := retry.Do(3, 2*time.Second, func() error {
					return sender.SendLogs(serverURL, apiKey, batch)
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
