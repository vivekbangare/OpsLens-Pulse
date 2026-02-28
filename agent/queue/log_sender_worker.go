package queue

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"time"

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

			// 1️⃣ Read with short lock
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

			// 2️⃣ Send WITHOUT lock
			for _, line := range lines {

				var batch sender.LogBatch
				if err := json.Unmarshal(line, &batch); err != nil {
					continue
				}

				err := retry.Do(3, 2*time.Second, func() error {
					return sender.SendLogs(serverURL, apiKey, batch)
				})

				if err != nil {
					unsent = append(unsent, line)
				}
			}

			// 3️⃣ Rewrite only unsent
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
