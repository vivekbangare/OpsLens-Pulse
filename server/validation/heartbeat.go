package validation

import (
	"errors"
	"strings"

	"opslense-pulse/shared"
)

func ValidateHeartbeat(h *shared.Heartbeat) error {
	if strings.TrimSpace(h.AgentID) == "" {
		return errors.New("agent_id is required")
	}

	if strings.TrimSpace(h.Hostname) == "" {
		return errors.New("hostname is required")
	}

	return nil
}
