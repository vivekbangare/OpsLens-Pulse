package collector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"opslense-pulse/agent/logs"
	"opslense-pulse/agent/sender"
)

// StartLogCollector collects host logs and sends them to the server
func StartLogCollector(
	ctx context.Context,
	agentID, hostname, logPath, serverURL, apiKey string,
	intervalSeconds int,
	accountID string,
) {
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	fmt.Println("📄 Host log collector started")
	fmt.Println("   file:", logPath)
	fmt.Println("   agent:", agentID)
	fmt.Println("   host:", hostname)
	fmt.Println("   account:", accountID)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("🛑 Host log collector stopping...")
			return

		case <-ticker.C:
			content, err := logs.ReadLastLines(logPath, 50)
			if err != nil {
				fmt.Println("❌ Error reading logs:", err)
				continue
			}

			lines := splitLines(content)
			if len(lines) == 0 {
				continue
			}

			var entries []sender.LogEntry

			for _, line := range lines {
				ts, msg, _ := parseLogLine(line)
				if msg == "" {
					continue
				}

				entries = append(entries, sender.LogEntry{
					AccountID: accountID,
					AgentID:   agentID,
					Hostname:  hostname,
					Timestamp: ts,
					Level:     "info",
					Message:   msg,
					Tags: map[string]string{
						"source": "host-log",
					},
				})
			}

			if len(entries) == 0 {
				continue
			}

			fmt.Printf(
				"🚚 Sending host logs: account=%s agent=%s host=%s count=%d\n",
				accountID,
				agentID,
				hostname,
				len(entries),
			)

			batch := sender.LogBatch{
				AccountID: accountID,
				AgentID:   agentID,
				Hostname:  hostname,
				Logs:      entries,
			}

			if err := retrySend(3, 2*time.Second, func() error {
				return sender.SendLogs(serverURL, apiKey, batch)
			}); err != nil {
				fmt.Println("❌ Failed to send logs:", err)
			}
		}
	}
}

// -----------------------
// Helpers
// -----------------------

func splitLines(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func parseLogLine(line string) (int64, string, error) {
	// Expected: "YYYY-MM-DD HH:MM:SS message"
	if len(line) < 20 {
		return time.Now().Unix(), line, nil
	}

	tsPart := line[:19]
	msg := strings.TrimSpace(line[19:])

	t, err := time.Parse("2006-01-02 15:04:05", tsPart)
	if err != nil {
		return time.Now().Unix(), line, nil
	}

	return t.Unix(), msg, nil
}
