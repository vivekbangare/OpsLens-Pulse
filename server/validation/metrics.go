package validation

import (
	"errors"
	"strings"

	"opslense-pulse/shared"
)

func ValidateHostMetrics(m *shared.HostMetrics) error {
	if strings.TrimSpace(m.AgentID) == "" {
		return errors.New("agent_id is required")
	}

	if strings.TrimSpace(m.Hostname) == "" {
		return errors.New("hostname is required")
	}

	if m.CPUPercent < 0 || m.CPUPercent > 100 {
		return errors.New("cpu_percent must be between 0 and 100")
	}

	if m.MemTotalMB <= 0 {
		return errors.New("mem_total_mb must be positive")
	}

	return nil
}
