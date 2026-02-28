package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"opslense-pulse/agent/metrics"
	"opslense-pulse/shared"
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
	shared.Info(
		"send logs called",
		"agent", payload.AgentID,
		"host", payload.Hostname,
		"count", len(payload.Logs),
	)
	body, error := json.Marshal(payload)

	if error != nil {
		return error
	}
	req, err := http.NewRequest("POST", serverURL+"/api/logs", bytes.NewBuffer(body))
	if err != nil {
		metrics.IncLogFailure()
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-Agent-Version", agentVersion)

	resp, err := httpClient.Do(req)
	// fmt.Println("📤 POST /api/logs →", serverURL)
	// fmt.Printf("📦 Payload size: %d logs\n", len(payload.Logs))
	if err != nil {
		metrics.IncLogFailure()
		return err
	}
	defer resp.Body.Close()
	//fmt.Println("📥 Response status:", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		metrics.IncLogFailure()
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}
