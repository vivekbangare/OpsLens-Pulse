package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// LogEntry represents a log entry
type LogEntry struct {
	AgentID   string            `json:"agent_id"`
	Hostname  string            `json:"hostname"`
	Timestamp int64             `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Tags      map[string]string `json:"tags"`
}

type LogBatch struct {
	AgentID  string     `json:"agent_id"`
	Hostname string     `json:"hostname"`
	Logs     []LogEntry `json:"logs"`
}

func SendLogs(serverURL, apiKey string, payload LogBatch) error {
	log.Printf(
		"🚨 SEND LOGS CALLED: agent=%s host=%s count=%d server=%s",
		payload.AgentID,
		payload.Hostname,
		len(payload.Logs),
		serverURL,
	)
	body, error := json.Marshal(payload)

	if error != nil {
		return error
	}
	//log.Printf("📦 LOG PAYLOAD: %s", string(body))
	req, err := http.NewRequest("POST", serverURL+"/api/logs", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	fmt.Println("📤 POST /api/logs →", serverURL)
	fmt.Printf("📦 Payload size: %d logs\n", len(payload.Logs))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Println("📥 Response status:", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}
