package containers

import (
	"opslense-pulse/shared"
)

func ToSharedMetrics(
	raw RawContainerMetrics,
	accountID, agentID, hostname string,
) shared.ContainerMetrics {

	return shared.ContainerMetrics{
		AccountID:   accountID,
		AgentID:     agentID,
		Hostname:    hostname,
		ContainerID: raw.ID,
		ContainerName: raw.Name,
		Image:       raw.Image,
		Status:      raw.Status,
		CPUPercent:  float32(raw.CPU),
		MemTotalMB:  float32(raw.MemLimitMB),
		MemUsedMB:   float32(raw.MemUsedMB),
		Timestamp:   raw.Timestamp.Unix(),
	}
}
