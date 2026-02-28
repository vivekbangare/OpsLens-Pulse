package collector

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"opslense-pulse/agent/batcher"
	"opslense-pulse/agent/config"
	"opslense-pulse/agent/limiter"
	"opslense-pulse/agent/queue"
	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

func StartLogSource(
	ctx context.Context,
	agentID, hostname string,
	src config.LogSource,
	logsQueue *queue.FileQueue,
) {

	shared.Info("starting log source", "name", src.Name, "type", src.Type)

	switch src.Type {

	case "file":
		go startFileCollector(ctx, agentID, hostname, src, logsQueue)

	case "directory":
		go startDirectoryCollector(ctx, agentID, hostname, src, logsQueue)

	case "journal":
		go StartJournalCollector(ctx, agentID, hostname, logsQueue)

	case "dmesg":
		go startDmesgMode(ctx, agentID, hostname, logsQueue)

	case "windows-event":
		if runtime.GOOS == "windows" {
			go StartWindowsEventCollector(
				ctx,
				agentID,
				hostname,
				src.Channel,
				logsQueue,
			)
		}

	default:
		shared.Warn("Unsupported log source type", "type", src.Type)
	}
}

func startDirectoryCollector(
	ctx context.Context,
	agentID, hostname string,
	src config.LogSource,
	logsQueue *queue.FileQueue,
) {

	shared.Info("directory collector started", "path", src.Path)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	running := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			files, err := os.ReadDir(src.Path)
			if err != nil {
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

				fileSrc := src
				fileSrc.Path = fullPath
				fileSrc.Name = src.Name + ":" + fileName

				running[fullPath] = true

				go startFileCollector(ctx, agentID, hostname, fileSrc, logsQueue)
			}
		}
	}
}

/* =========================
   SECURE MODES BELOW
   ========================= */

func startDmesgMode(
	ctx context.Context,
	agentID, hostname string,
	logsQueue *queue.FileQueue,
) {
	if runtime.GOOS != "linux" {
		return
	}

	executeAndSend(ctx, agentID, hostname, logsQueue, "dmesg")
}

func startDockerLogsMode(
	ctx context.Context,
	agentID, hostname string,
	logsQueue *queue.FileQueue,
) {
	executeAndSend(ctx, agentID, hostname, logsQueue, "docker", "logs", "--tail", "100")
}

/* =========================
   SAFE EXECUTION CORE
   ========================= */

func executeAndSend(
	ctx context.Context,
	agentID, hostname string,
	logsQueue *queue.FileQueue,
	name string,
	args ...string,
) {

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	logBatcher := batcher.New(200, 1*time.Second)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:

			// 5 second execution timeout
			execCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

			cmd := exec.CommandContext(execCtx, name, args...)
			out, err := cmd.Output()
			cancel()

			if err != nil {
				continue
			}

			lines := splitLines(string(out))
			if len(lines) == 0 {
				continue
			}

			for _, l := range lines {

				if !limiter.Allow() {
					continue
				}

				entry := sender.LogEntry{
					AgentID:   agentID,
					Hostname:  hostname,
					Timestamp: time.Now().Unix(),
					Level:     ExtractLevel(l),
					Message:   l,
					Tags: map[string]string{
						"type": name,
					},
				}

				outBatch := logBatcher.Add(entry)

				if outBatch != nil {
					var entries []sender.LogEntry
					for _, e := range outBatch {
						entries = append(entries, e.(sender.LogEntry))
					}

					batch := sender.LogBatch{
						AgentID:  agentID,
						Hostname: hostname,
						Logs:     entries,
					}

					_ = logsQueue.Enqueue(batch)
				}
			}
		}
	}
}

/* ========================= */

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
