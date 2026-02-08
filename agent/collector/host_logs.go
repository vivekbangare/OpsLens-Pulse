package collector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"opslense-pulse/agent/logs"
	"opslense-pulse/agent/sender"
)

var lastSentLineFile = filepath.Join(BaseDir, "agent_last_line.txt")

// StartLogCollector collects host logs and sends them to the server
func StartLogCollector(
	ctx context.Context,
	agentID, logPath, serverURL, apiKey string,
	intervalSeconds int,
	accountID string,
) {
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	lastSentLine, _ := loadLastLinePosition()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Host log collector stopping...")
			return

		case <-ticker.C:
			content, err := logs.ReadLastLines(logPath, 50)
			if err != nil {
				fmt.Println("Error reading logs:", err)
				continue
			}

			allLines := splitLines(content)
			if lastSentLine >= len(allLines) {
				continue
			}

			newLines := allLines[lastSentLine:]
			lastSentLine = len(allLines)
			_ = saveLastLinePosition(lastSentLine)

			var entries []sender.LogEntry
			for _, line := range newLines {
				ts, msg, _ := parseLogLine(line)
				if ts == 0 {
					ts = time.Now().Unix()
				}
				if msg == "" {
					msg = line
				}

				entries = append(entries, sender.LogEntry{
					AgentID:   agentID,
					Timestamp: ts,
					Message:   msg,
				})
			}

			if len(entries) == 0 {
				continue
			}

			batch := sender.LogBatch{
				AgentID: agentID,
				Logs:    entries,
			}

			go func(b sender.LogBatch) {
				if err := retrySend(3, 2*time.Second, func() error {
					return sender.SendLogs(serverURL, apiKey, b)
				}); err != nil {
					fmt.Println("Error sending logs:", err)
				}
			}(batch)
		}
	}
}

// splitLines splits log content into non-empty lines
func splitLines(content string) []string {
	var raw []string
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) != "" {
			raw = append(raw, line)
		}
	}
	return raw
}

// loadLastLinePosition loads last sent line number from file
func loadLastLinePosition() (int, error) {
	data, err := os.ReadFile(lastSentLineFile)
	if err != nil {
		return 0, nil
	}
	num, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, nil
	}
	return num, nil
}

// saveLastLinePosition saves last sent line number to file
func saveLastLinePosition(pos int) error {
	return os.WriteFile(lastSentLineFile, []byte(strconv.Itoa(pos)), 0644)
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
// 	return fmt.Errorf("all retries failed")
// }

// parseLogLine is a placeholder for parsing timestamp from log line
func parseLogLine(line string) (int64, string, error) {
	// TODO: implement proper timestamp extraction if needed
	return 0, line, nil
}
