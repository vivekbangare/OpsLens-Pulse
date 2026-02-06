package shared

import "time"

type HostMetrics struct {
	AccountID  string            `json:"account_id"`
	AgentID    string            `json:"agent_id"`
	Hostname   string            `json:"hostname"`
	OS         string            `json:"os"`
	Timestamp  time.Time         `json:"timestamp"`
	Cores      int               `json:"cores"`
	MemTotalMB uint64            `json:"mem_total_mb"`
	UpTimeSec  uint64            `json:"uptime_sec"`
	MemUsedMB  uint64            `json:"mem_used_mb"`
	CPUPercent float64           `json:"cpu_percent"`
	Tags       map[string]string `json:"tags"`
	IP         string            `json:"ip"`
}

type ContainerInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type Heartbeat struct {
	AccountID string    `json:"account_id"`
	AgentID   string    `json:"agent_id"`
	Hostname  string    `json:"hostname"`
	Timestamp time.Time `json:"timestamp"`
}
