package heartbeat

import (
	"time"

	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

type Heartbeat struct {
	AgentID   string    `json:"agent_id"`
	Hostname  string    `json:"hostname"`
	Timestamp time.Time `json:"timestamp"`
}

// Send sends a heartbeat to server
func Send(serverURL, token, agentID, hostname string) error {
	hb := shared.Heartbeat{
		AgentID:   agentID,
		Hostname:  hostname,
		Timestamp: time.Now().Unix(),
	}

	return sender.SendHeartbeat(serverURL, token, hb)
}
