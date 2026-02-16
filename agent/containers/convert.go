package containers

import (
	"opslense-pulse/shared"
)

func ToSharedMetrics(
	raw RawContainerMetrics,
	agentID, hostname string,
) shared.ContainerMetrics {

	return shared.ContainerMetrics{
		AgentID:       agentID,
		Hostname:      hostname,
		ContainerID:   raw.ID,
		ContainerName: raw.Name,
		Image:         raw.Image,
		Status:        raw.Status,
		CPUPercent:    float32(raw.CPU),
		MemTotalMB:    float32(raw.MemLimitMB),
		MemUsedMB:     float32(raw.MemUsedMB),
		Timestamp:     raw.Timestamp.Unix(),
	}
}
