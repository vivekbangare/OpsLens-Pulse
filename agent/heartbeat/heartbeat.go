package heartbeat

import (
	"time"

	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

type Heartbeat struct {
	AccountID string    `json:"account_id"`
	AgentID   string    `json:"agent_id"`
	Hostname  string    `json:"hostname"`
	Timestamp time.Time `json:"timestamp"`
}

// Send sends a heartbeat to server
func Send(serverURL, token, accountID, agentID, hostname string) error {
	hb := shared.Heartbeat{
		AccountID: accountID,
		AgentID:   agentID,
		Hostname:  hostname,
		Timestamp: time.Now(),
	}

	return sender.SendHeartbeat(serverURL, token, hb)
}
