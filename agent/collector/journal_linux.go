//go:build linux

package collector

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"opslense-pulse/agent/batcher"
	"opslense-pulse/agent/limiter"
	"opslense-pulse/agent/queue"
	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

var (
	journalCursorFile = filepath.Join(StateDir, "journal_cursor")
	journalCursor     string
	journalCursorLock = &sync.Mutex{}
)

func loadJournalCursor() {
	data, err := os.ReadFile(journalCursorFile)
	if err != nil {
		return
	}
	journalCursorLock.Lock()
	journalCursor = string(data)
	journalCursorLock.Unlock()
}

func saveJournalCursor(cursor string) {
	journalCursorLock.Lock()
	defer journalCursorLock.Unlock()

	journalCursor = cursor

	tmp := journalCursorFile + ".tmp"
	_ = os.WriteFile(tmp, []byte(cursor), 0600)
	_ = os.Rename(tmp, journalCursorFile)
}

func StartJournalCollector(
	ctx context.Context,
	agentID, hostname string,
	logsQueue *queue.FileQueue,
) {

	shared.Info("journalctl collector started")

	loadJournalCursor()

	logBatcher := batcher.New(200, 1*time.Second)

	args := []string{
		"-f",
		"-o", "json",
	}

	if journalCursor != "" {
		args = append(args, "--after-cursor", journalCursor)
	} else {
		args = append(args, "-n", "0")
	}

	cmd := exec.CommandContext(ctx, "journalctl", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		shared.Error("journal stdout pipe failed", "error", err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		shared.Error("journalctl start failed", "error", err.Error())
		return
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {

		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			return
		default:
		}

		if !limiter.Allow() {
			continue
		}

		line := scanner.Bytes()

		var obj map[string]interface{}
		if err := json.Unmarshal(line, &obj); err != nil {
			continue
		}

		msg, _ := obj["MESSAGE"].(string)
		cursor, _ := obj["__CURSOR"].(string)

		if msg == "" {
			continue
		}

		logEntry := sender.LogEntry{
			AgentID:   agentID,
			Hostname:  hostname,
			Timestamp: time.Now().Unix(),
			Level:     ExtractLevel(msg),
			Message:   msg,
			Tags: map[string]string{
				"type": "journal",
			},
		}

		out := logBatcher.Add(logEntry)

		if out != nil {
			var entries []sender.LogEntry
			for _, e := range out {
				entries = append(entries, e.(sender.LogEntry))
			}

			batch := sender.LogBatch{
				AgentID:  agentID,
				Hostname: hostname,
				Logs:     entries,
			}

			_ = logsQueue.Enqueue(batch)
		}

		if cursor != "" {
			saveJournalCursor(cursor)
		}
	}

	_ = cmd.Wait()
}
