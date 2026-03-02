package sender

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"opslense-pulse/agent/internal"
	"opslense-pulse/agent/metrics"
	"opslense-pulse/shared"
)

var httpClient *http.Client
var consecutiveFailures int32
var agentVersion string

func InitHTTPClient(skipVerify bool) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipVerify,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	httpClient = &http.Client{
		Timeout:   15 * time.Second,
		Transport: tr,
	}
}

func SetAgentVersion(v string) {
	agentVersion = v
}

func handleFailure() {
	internal.IncFailure()
	n := atomic.AddInt32(&consecutiveFailures, 1)

	if n > 5 {
		time.Sleep(30 * time.Second)
	}
}

func handleSuccess() {
	internal.ResetFailure()
	atomic.StoreInt32(&consecutiveFailures, 0)
}

func SendRaw(serverURL, apiKey string, metricsPayload shared.HostMetrics) error {
	body, err := json.Marshal(metricsPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", serverURL+"/api/metrics", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-Agent-Version", agentVersion)

	resp, err := httpClient.Do(req)
	if err != nil {
		handleFailure()
		metrics.IncMetricsFailure()
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		handleFailure()
		metrics.IncMetricsFailure()
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	handleSuccess()
	return nil
}

func SendHeartbeat(serverURL, apiKey string, hb shared.Heartbeat) error {
	body, err := json.Marshal(hb)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", serverURL+"/api/heartbeat", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-Agent-Version", agentVersion)

	resp, err := httpClient.Do(req)
	if err != nil {
		handleFailure()
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		handleFailure()
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	handleSuccess()
	return nil
}

func SendContainerMetrics(serverURL, apiKey string, m shared.ContainerMetrics) error {

	body, err := json.Marshal(m)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		serverURL+"/api/container/metrics",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-Agent-Version", agentVersion)

	resp, err := httpClient.Do(req)
	if err != nil {
		handleFailure()
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		handleFailure()
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	handleSuccess()
	return nil
}

func SendAgentRegistration(serverURL, apiKey string, a shared.AgentInfo) error {

	body, _ := json.Marshal(a)

	req, _ := http.NewRequest("POST",
		fmt.Sprintf("%s/api/agents/register", serverURL),
		bytes.NewBuffer(body),
	)

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"registration failed: status=%d body=%s",
			resp.StatusCode,
			string(bodyBytes),
		)
	}

	return nil
}
