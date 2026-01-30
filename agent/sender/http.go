package sender

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"opslense-pulse/shared"
)

// Send sends host metrics to server
func Send(serverURL, token string, metrics shared.HostMetrics) error {
	body, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		serverURL+"/api/metrics",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// SendHeartbeat sends heartbeat info
func SendHeartbeat(serverURL, token string, hb shared.Heartbeat) error {
	body, _ := json.Marshal(hb)

	req, _ := http.NewRequest(
		"POST",
		serverURL+"/api/heartbeat",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	_, err := client.Do(req)
	return err
}
