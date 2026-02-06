package heartbeat

import (
	"time"

	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

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
