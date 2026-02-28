package validation

import (
	"errors"
	"strings"

	"opslense-pulse/shared"
)

func ValidateContainerMetrics(m *shared.ContainerMetrics) error {
	if strings.TrimSpace(m.AgentID) == "" {
		return errors.New("agent_id is required")
	}

	if strings.TrimSpace(m.ContainerID) == "" {
		return errors.New("container_id is required")
	}

	if strings.TrimSpace(m.ContainerName) == "" {
		return errors.New("container_name is required")
	}

	return nil
}
