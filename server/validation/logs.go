package validation

import (
	"errors"
	"strings"

	"opslense-pulse/shared"
)

func ValidateLogBatch(b *shared.LogBatch) error {
	if strings.TrimSpace(b.AgentID) == "" {
		return errors.New("agent_id is required")
	}

	if len(b.Logs) == 0 {
		return errors.New("logs array cannot be empty")
	}

	return nil
}
