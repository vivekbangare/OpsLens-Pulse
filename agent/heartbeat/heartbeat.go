package heartbeat

import (
	"time"

	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

// Send sends a heartbeat to server
func Send(serverURL, token, hostname string) error {
	hb := shared.Heartbeat{
		Hostname:  hostname,
		Timestamp: time.Now(),
	}

	return sender.SendHeartbeat(serverURL, token, hb)
}
