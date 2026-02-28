//go:build windows

package collector

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"opslense-pulse/agent/batcher"
	"opslense-pulse/agent/limiter"
	"opslense-pulse/agent/queue"
	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

type eventXML struct {
	System struct {
		Provider struct {
			Name string `xml:"Name,attr"`
		} `xml:"Provider"`
		EventID       int   `xml:"EventID"`
		Level         int   `xml:"Level"`
		EventRecordID int64 `xml:"EventRecordID"`
		TimeCreated   struct {
			SystemTime string `xml:"SystemTime,attr"`
		} `xml:"TimeCreated"`
	} `xml:"System"`
	EventData struct {
		Data []string `xml:"Data"`
	} `xml:"EventData"`
}

var (
	cursorFile     = filepath.Join(StateDir, "windows_event_cursor.json")
	cursorLock     = &sync.Mutex{}
	channelCursors = make(map[string]int64)
)

func loadCursors() {
	data, err := os.ReadFile(cursorFile)
	if err != nil {
		return
	}
	cursorLock.Lock()
	defer cursorLock.Unlock()
	_ = json.Unmarshal(data, &channelCursors)
}

func saveCursors() {
	cursorLock.Lock()
	defer cursorLock.Unlock()

	data, err := json.Marshal(channelCursors)
	if err != nil {
		return
	}

	tmp := cursorFile + ".tmp"
	_ = os.WriteFile(tmp, data, 0600)
	_ = os.Rename(tmp, cursorFile)
}

func StartWindowsEventCollector(
	ctx context.Context,
	agentID, hostname string,
	srcChannel string,
	logsQueue *queue.FileQueue,
) {

	shared.Info("windows event collector started", "channel", srcChannel)

	loadCursors()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	logBatcher := batcher.New(200, 1*time.Second)

	for {
		select {

		case <-ctx.Done():
			saveCursors()
			return

		case <-ticker.C:

			cursorLock.Lock()
			lastID := channelCursors[srcChannel]
			cursorLock.Unlock()

			query := "*[System[(EventRecordID > " + strconv.FormatInt(lastID, 10) + ")]]"
			execCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(
				execCtx,
				"wevtutil",
				"qe",
				srcChannel,
				"/q:"+query,
				"/f:xml",
			)
			out, err := cmd.Output()
			if err != nil || len(out) == 0 {
				continue
			}

			events := splitXML(string(out))

			for _, e := range events {

				if !limiter.Allow() {
					continue
				}

				var parsed eventXML
				if err := xml.Unmarshal([]byte(e), &parsed); err != nil {
					continue
				}

				level := mapLevel(parsed.System.Level)

				message := strings.Join(parsed.EventData.Data, " | ")

				entry := sender.LogEntry{
					AgentID:   agentID,
					Hostname:  hostname,
					Timestamp: time.Now().Unix(),
					Level:     level,
					Message:   message,
					Tags: map[string]string{
						"type":     "windows-event",
						"channel":  srcChannel,
						"provider": parsed.System.Provider.Name,
						"event_id": strconv.Itoa(parsed.System.EventID),
					},
				}

				outBatch := logBatcher.Add(entry)

				if outBatch != nil {
					var entries []sender.LogEntry
					for _, ev := range outBatch {
						entries = append(entries, ev.(sender.LogEntry))
					}

					batch := sender.LogBatch{
						AgentID:  agentID,
						Hostname: hostname,
						Logs:     entries,
					}

					_ = logsQueue.Enqueue(batch)
				}

				cursorLock.Lock()
				channelCursors[srcChannel] = parsed.System.EventRecordID
				cursorLock.Unlock()
			}

			saveCursors()
		}
	}
}

func splitXML(input string) []string {
	parts := strings.Split(input, "<Event ")
	var events []string

	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			continue
		}
		events = append(events, "<Event "+p)
	}

	return events
}

func mapLevel(level int) string {
	switch level {
	case 1:
		return "critical"
	case 2:
		return "error"
	case 3:
		return "warn"
	case 4:
		return "info"
	case 5:
		return "debug"
	default:
		return "info"
	}
}
