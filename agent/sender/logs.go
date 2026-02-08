package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LogEntry represents a log entry
type LogEntry struct {
	AgentID   string `json:"agent_id"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

type LogBatch struct {
	AgentID string     `json:"agent_id"`
	Logs    []LogEntry `json:"logs"`
}

func SendLogs(serverURL, apiKey string, payload LogBatch) error {
	body, error := json.Marshal(payload)

	if error != nil {
		return error
	}

	req, err := http.NewRequest("POST", serverURL+"/api/logs", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}
	return nil
}
