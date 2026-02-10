package collector

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"opslense-pulse/agent/config"
	"opslense-pulse/agent/logs"
	"opslense-pulse/agent/sender"
)

func startFileCollector(
	ctx context.Context,
	accountID, agentID, hostname string,
	src config.LogSource,
	serverURL, apiKey string,
	interval int,
) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	fmt.Println("📄 File log collector started:", src.Path)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			content, err := logs.ReadLastLines(src.Path, 50)
			if err != nil {
				fmt.Println("❌ Read error:", err)
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
						"source": src.Name,
						"type":   "file",
					},
				})
			}

			if len(entries) == 0 {
				continue
			}

			batch := sender.LogBatch{
				AccountID: accountID,
				AgentID:   agentID,
				Hostname:  hostname,
				Logs:      entries,
			}

			_ = retrySend(3, 2*time.Second, func() error {
				return sender.SendLogs(serverURL, apiKey, batch)
			})
		}
	}
}

func startDirectoryCollector(
	ctx context.Context,
	accountID, agentID, hostname string,
	src config.LogSource,
	serverURL, apiKey string,
	interval int,
) {
	files, _ := filepath.Glob(filepath.Join(src.Path, "*"))

	for _, f := range files {
		fileSrc := src
		fileSrc.Path = f
		fileSrc.Name = src.Name + ":" + filepath.Base(f)

		go startFileCollector(
			ctx,
			accountID,
			agentID,
			hostname,
			fileSrc,
			serverURL,
			apiKey,
			interval,
		)
	}
}

func startCommandCollector(
	ctx context.Context,
	accountID, agentID, hostname string,
	src config.LogSource,
	serverURL, apiKey string,
	interval int,
) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			out, err := exec.Command("sh", "-c", src.Command).Output()
			if err != nil {
				continue
			}

			lines := splitLines(string(out))
			if len(lines) == 0 {
				continue
			}

			var entries []sender.LogEntry
			for _, l := range lines {
				entries = append(entries, sender.LogEntry{
					AccountID: accountID,
					AgentID:   agentID,
					Hostname:  hostname,
					Timestamp: time.Now().Unix(),
					Level:     "info",
					Message:   l,
					Tags: map[string]string{
						"source": src.Name,
						"type":   "command",
					},
				})
			}

			batch := sender.LogBatch{
				AccountID: accountID,
				AgentID:   agentID,
				Hostname:  hostname,
				Logs:      entries,
			}

			_ = sender.SendLogs(serverURL, apiKey, batch)
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
