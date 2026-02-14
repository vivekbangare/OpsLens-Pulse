package collector

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"opslense-pulse/agent/config"
	"opslense-pulse/agent/sender"
)

func startFileCollector(
	ctx context.Context,
	accountID, agentID, hostname string,
	src config.LogSource,
	serverURL, apiKey string,
	interval int,
) {

	loadState()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	fmt.Println("📄 File log collector started:", src.Path)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:

			file, err := os.Open(src.Path)
			if err != nil {
				fmt.Println("❌ File open error:", err)
				continue
			}

			stat, err := file.Stat()
			if err != nil {
				file.Close()
				continue
			}

			sysStat := stat.Sys().(*syscall.Stat_t)
			inode := sysStat.Ino

			state := fileStates[src.Path]

			// Rotation detection
			if state.Inode != inode {
				fmt.Println("🔄 Log rotation detected:", src.Path)
				state.Offset = 0
				state.Inode = inode
			}

			// Truncation detection
			if stat.Size() < state.Offset {
				fmt.Println("⚠️ File truncated:", src.Path)
				state.Offset = 0
			}

			_, err = file.Seek(state.Offset, io.SeekStart)
			if err != nil {
				file.Close()
				continue
			}

			reader := bufio.NewReader(file)
			var buffer strings.Builder
			var entries []sender.LogEntry
			newOffset := state.Offset

			flush := func() {
				if buffer.Len() == 0 {
					return
				}

				fullLine := buffer.String()
				ts, msg, _ := parseLogLine(fullLine)

				if strings.TrimSpace(msg) == "" {
					buffer.Reset()
					return
				}

				entries = append(entries, sender.LogEntry{
					AccountID: accountID,
					AgentID:   agentID,
					Hostname:  hostname,
					Timestamp: ts,
					Level:     "info",
					Message:   strings.TrimSpace(msg),
					Tags: map[string]string{
						"source": src.Name,
						"type":   "file",
					},
				})

				buffer.Reset()
			}

			for {
				line, err := reader.ReadString('\n')

				if err == io.EOF {
					break
				}
				if err != nil {
					break
				}

				newOffset += int64(len(line))
				line = strings.TrimRight(line, "\n")

				if isTimestampLine(line) {
					flush()
					buffer.WriteString(line)
				} else {
					if buffer.Len() > 0 {
						buffer.WriteString("\n")
					}
					buffer.WriteString(line)
				}

				if len(entries) >= 1000 {
					break
				}
			}

			// Flush remaining multiline entry
			flush()

			file.Close()

			if len(entries) == 0 {
				continue
			}

			batch := sender.LogBatch{
				AccountID: accountID,
				AgentID:   agentID,
				Hostname:  hostname,
				Logs:      entries,
			}

			err = sender.SendLogs(serverURL, apiKey, batch)
			if err != nil {
				fmt.Println("❌ Send failed, keeping offset unchanged")
				continue
			}

			// Update offset only after successful send
			state.Offset = newOffset
			fileStates[src.Path] = state
			saveState()

			// Backlog catch-up mode
			backlog := stat.Size() - state.Offset
			if backlog > 5*1024*1024 {
				fmt.Println("⚡ Large backlog detected, reprocessing immediately")
				continue
			}
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

	fmt.Println("📁 Directory log collector started:", src.Path)

	ticker := time.NewTicker(10 * time.Second) // re-scan interval
	defer ticker.Stop()

	running := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			files, err := os.ReadDir(src.Path)
			if err != nil {
				fmt.Println("❌ Directory read error:", err)
				continue
			}

			for _, entry := range files {

				if entry.IsDir() {
					continue
				}

				fileName := entry.Name()
				fullPath := filepath.Join(src.Path, fileName)

				if !matchesInclude(fileName, src.Include) {
					continue
				}

				if matchesExclude(fileName, src.Exclude) {
					continue
				}

				if running[fullPath] {
					continue
				}

				fmt.Println("📄 Starting collector for:", fullPath)

				fileSrc := src
				fileSrc.Path = fullPath
				fileSrc.Name = src.Name + ":" + fileName

				running[fullPath] = true

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
	}
}

func matchesInclude(fileName string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}

	for _, p := range patterns {
		matched, _ := filepath.Match(p, fileName)
		if matched {
			return true
		}
	}
	return false
}

func matchesExclude(fileName string, patterns []string) bool {
	for _, p := range patterns {
		matched, _ := filepath.Match(p, fileName)
		if matched {
			return true
		}
	}
	return false
}

func isTimestampLine(line string) bool {
	if len(line) < 19 {
		return false
	}
	_, err := time.Parse("2006-01-02 15:04:05", line[:19])
	return err == nil
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
