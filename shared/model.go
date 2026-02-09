package shared

import "time"

// -------------------------------
// HostMetrics: metrics sent by an agent
// -------------------------------
type HostMetrics struct {
	AccountID    string            `json:"account_id"` // multi-tenant
	AgentID      string            `json:"agent_id"`   // unique agent id
	Hostname     string            `json:"hostname"`   // host name
	OS           string            `json:"os"`         // add this
	IP           string            `json:"ip"`
	Cores        int               `json:"cores"`
	CPUPercent   float32           `json:"cpu_percent"`    // CPU usage %
	MemUsedMB    float32           `json:"mem_used_mb"`    // Memory used in MB
	MemTotalMB   float32           `json:"mem_total_mb"`   // Total memory in MB
	DiskUsedMB   float32           `json:"disk_used_mb"`   // Disk used in MB
	DiskTotalMB  float32           `json:"disk_total_mb"`  // Disk total in MB
	NetworkInMB  float32           `json:"network_in_mb"`  // Network received in MB
	NetworkOutMB float32           `json:"network_out_mb"` // Network sent in MB
	UptimeSec    uint64            `json:"uptime_sec"`     // Uptime in seconds
	Tags         map[string]string `json:"tags"`           // key-value tags
	TTLDays      uint16            `json:"ttl_days"`       // TTL in days (optional)
	Timestamp    int64             `json:"timestamp"`      // optional: epoch seconds, default now
}

// -------------------------------
// LogEntry: individual log from agent
// -------------------------------
type LogEntry struct {
	AccountID string            `json:"account_id"` // multi-tenant
	AgentID   string            `json:"agent_id"`
	Hostname  string            `json:"hostname"`
	Timestamp int64             `json:"timestamp"` // epoch seconds
	Level     string            `json:"level"`     // info, warn, error
	Message   string            `json:"message"`
	Tags      map[string]string `json:"tags"`                 // optional tags
	TTLDays   uint16            `json:"ttl_days"`             // optional TTL
	HumanTime string            `json:"human_time,omitempty"` // formatted string for UI
}

// -------------------------------
// LogBatch: batch of logs sent by agent
// -------------------------------
type LogBatch struct {
	AccountID string     `json:"account_id"`
	AgentID   string     `json:"agent_id"`
	Hostname  string     `json:"hostname"`
	Logs      []LogEntry `json:"logs"`
}

// -------------------------------
// APIKey: API key info
// -------------------------------
type APIKey struct {
	AccountID   string `json:"account_id"`
	KeyID       string `json:"key_id"`
	KeyHash     string `json:"key_hash"`
	Name        string `json:"name"`
	IsActive    uint8  `json:"is_active"`    // 1 = active, 0 = inactive
	IsBootstrap uint8  `json:"is_bootstrap"` // 1 = bootstrap key
	CreatedAt   time.Time
}

// -------------------------------
// AgentInfo: for agent metadata & heartbeat
// -------------------------------
type AgentInfo struct {
	AccountID   string            `json:"account_id"`
	AgentID     string            `json:"agent_id"`
	Hostname    string            `json:"hostname"`
	IP          string            `json:"ip"`
	OS          string            `json:"os"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Tags        map[string]string `json:"tags"`
	FirstSeen   time.Time         `json:"first_seen"`
	LastSeen    time.Time         `json:"last_seen"`
	Alive       bool              `json:"alive"`
}

// -------------------------------
// Heartbeat struct (optional for API)
// -------------------------------
type Heartbeat struct {
	AccountID string    `json:"account_id"`
	AgentID   string    `json:"agent_id"`
	Hostname  string    `json:"hostname"`
	Timestamp time.Time `json:"timestamp"` // epoch seconds
}

type ContainerMetrics struct {
	AccountID     string            `json:"account_id"`
	AgentID       string            `json:"agent_id"`
	Hostname      string            `json:"hostname"`
	Image         string            `json:"image"`
	Status        string            `json:"status"`
	ContainerID   string            `json:"container_id"`
	ContainerName string            `json:"container_name"`
	CPUPercent    float32           `json:"cpu_percent"`
	MemUsedMB     float32           `json:"mem_used_mb"`
	MemTotalMB    float32           `json:"mem_total_mb"`
	DiskUsedMB    float32           `json:"disk_used_mb"`
	DiskTotalMB   float32           `json:"disk_total_mb"`
	NetworkInMB   float32           `json:"network_in_mb"`
	NetworkOutMB  float32           `json:"network_out_mb"`
	Tags          map[string]string `json:"tags"`
	Timestamp     int64             `json:"timestamp,omitempty"`
	TTLDays       uint16            `json:"ttl_days,omitempty"`
}

type ContainerLog struct {
	AccountID     string            `json:"account_id"`
	AgentID       string            `json:"agent_id"`
	Hostname      string            `json:"hostname"`
	ContainerID   string            `json:"container_id"`
	ContainerName string            `json:"container_name"`
	Level         string            `json:"level"`
	Message       string            `json:"message"`
	Tags          map[string]string `json:"tags"`
	Timestamp     int64             `json:"timestamp,omitempty"`
	TTLDays       uint16            `json:"ttl_days,omitempty"`
}

type ContainerLogBatch struct {
	AccountID string         `json:"account_id"`
	AgentID   string         `json:"agent_id"`
	Hostname  string         `json:"hostname"`
	Logs      []ContainerLog `json:"logs"`
}

type ContainerLogEntry struct {
	AccountID   string `json:"account_id"`
	AgentID     string `json:"agent_id"`
	Hostname    string `json:"hostname"`
	ContainerID string `json:"container_id"`
	Timestamp   int64  `json:"timestamp"`
	Message     string `json:"message"`
}
