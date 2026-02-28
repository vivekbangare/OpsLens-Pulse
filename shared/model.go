package shared

import "time"

// -------------------------------
// HostMetrics: metrics sent by an agent
// -------------------------------
type HostMetrics struct {
	TenantID string `json:"tenant_id"` // multi-tenant
	AgentID  string `json:"agent_id"`  // unique agent id
	Hostname string `json:"hostname"`  // host name
	OS       string `json:"os"`        // add this
	IP       string `json:"ip"`
	PublicIP string `json:"public_ip"` // add this
	Version  string `json:"version"`

	Cores      int     `json:"cores"`
	CPUPercent float32 `json:"cpu_percent"` // CPU usage %

	MemUsedMB  float32 `json:"mem_used_mb"`  // Memory used in MB
	MemTotalMB float32 `json:"mem_total_mb"` // Total memory in MB

	DiskUsedMB  float32 `json:"disk_used_mb"`  // Disk used in MB
	DiskTotalMB float32 `json:"disk_total_mb"` // Disk total in MB

	NetInBytesPerSec  float32 `json:"net_in_bytes_per_sec"`
	NetOutBytesPerSec float32 `json:"net_out_bytes_per_sec"`

	DiskReadBytesPerSec  float32 `json:"disk_read_bytes_per_sec"`
	DiskWriteBytesPerSec float32 `json:"disk_write_bytes_per_sec"`

	Load1  float32 `json:"load_1"`  // 1-minute load average
	Load5  float32 `json:"load_5"`  // 5-minute load average
	Load15 float32 `json:"load_15"` // 15-minute load average

	ProcTotal   int `json:"proc_total"`   // Total number of processes
	ProcRunning int `json:"proc_running"` // Number of running processes

	UptimeSec  uint64            `json:"uptime_sec"` // Uptime in seconds
	Tags       map[string]string `json:"tags"`       // key-value tags
	TTLDays    uint16            `json:"ttl_days"`   // TTL in days (optional)
	Timestamp  int64             `json:"timestamp"`  // optional: epoch seconds, default now
	Interfaces []InterfaceMetric `json:"interfaces,omitempty"`
	Agent      AgentHealth       `json:"agent"` // optional agent health metrics
	Services   []ServiceStatus   `json:"services,omitempty"`
	Disks      []DiskMetric      `json:"disks,omitempty"`
	// CPU intelligence
	CPUCritical bool `json:"cpu_critical,omitempty"`
	CPUSpike    bool `json:"cpu_spike,omitempty"`

	// Memory intelligence
	MemCritical bool               `json:"mem_critical,omitempty"`
	MemPressure bool               `json:"mem_pressure,omitempty"`
	Network     []NetworkInterface `json:"network"`
}

type ServiceStatus struct {
	Name          string `json:"name"`
	Running       bool   `json:"running"`
	Enabled       bool   `json:"enabled"`
	Status        string `json:"status"` // healthy | down | disabled
	CrashDetected bool   `json:"crash_detected,omitempty"`
}

type DiskMetric struct {
	MountPoint string  `json:"mount_point"`
	FSType     string  `json:"fs_type"`
	TotalMB    float32 `json:"total_mb"`
	UsedMB     float32 `json:"used_mb"`
	UsedPct    float32 `json:"used_percent"`
	// Lightweight anomaly intelligence
	SpikeDetected bool `json:"spike_detected"`
	Critical      bool `json:"critical"`
}

type InterfaceMetric struct {
	Name           string  `json:"name"`
	InBytesPerSec  float32 `json:"in_bytes_per_sec"`
	OutBytesPerSec float32 `json:"out_bytes_per_sec"`
}

type NetworkInterface struct {
	Name   string  `json:"name"`
	InBps  float32 `json:"in_bps"`
	OutBps float32 `json:"out_bps"`
}

type AgentHealth struct {
	CPUPercent      float32 `json:"agent_cpu_percent"`
	MemoryMB        float32 `json:"agent_mem_mb"`
	Goroutines      int     `json:"agent_goroutines"`
	UptimeSec       uint64  `json:"agent_uptime_sec"`
	MetricsFailures uint64  `json:"metrics_send_failures"`
	LogFailures     uint64  `json:"logs_send_failures"`
}

type Event struct {
	TenantID    string `json:"tenant_id"`
	AgentID     string `json:"agent_id"`
	Hostname    string `json:"hostname"`
	EventType   string `json:"event_type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Timestamp   int64  `json:"timestamp"`
}

// -------------------------------
// LogEntry: individual log from agent
// -------------------------------
type LogEntry struct {
	TenantID   string            `json:"tenant_id"` // multi-tenant
	AgentID    string            `json:"agent_id"`
	Hostname   string            `json:"hostname"`
	Timestamp  int64             `json:"timestamp"` // epoch seconds
	Level      string            `json:"level"`     // info, warn, error
	Message    string            `json:"message"`
	Tags       map[string]string `json:"tags"`                 // optional tags
	TTLDays    uint16            `json:"ttl_days"`             // optional TTL
	HumanTime  string            `json:"human_time,omitempty"` // formatted string for UI
	SourceName string            `json:"source_name"`          // optional source name (e.g. app name)
	SourceType string            `json:"source_type"`          // optional source type (e.g. host, container)
}

// -------------------------------
// LogBatch: batch of logs sent by agent
// -------------------------------
type LogBatch struct {
	TenantID string     `json:"tenant_id"`
	AgentID  string     `json:"agent_id"`
	Hostname string     `json:"hostname"`
	Logs     []LogEntry `json:"logs"`
}

// -------------------------------
// APIKey: API key info
// -------------------------------
type APIKey struct {
	TenantID    string `json:"tenant_id"`
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
	TenantID    string            `json:"tenant_id"`
	AgentID     string            `json:"agent_id"`
	Hostname    string            `json:"hostname"`
	IP          string            `json:"ip"`
	PublicIP    string            `json:"public_ip"`
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
	TenantID  string    `json:"tenant_id"`
	AgentID   string    `json:"agent_id"`
	Hostname  string    `json:"hostname"`
	Timestamp time.Time `json:"timestamp"` // epoch seconds
}

type ContainerMetrics struct {
	TenantID      string            `json:"tenant_id"`
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
	TenantID      string            `json:"tenant_id"`
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
	TenantID string         `json:"tenant_id"`
	AgentID  string         `json:"agent_id"`
	Hostname string         `json:"hostname"`
	Logs     []ContainerLog `json:"logs"`
}

type ContainerLogEntry struct {
	TenantID    string `json:"tenant_id"`
	AgentID     string `json:"agent_id"`
	Hostname    string `json:"hostname"`
	ContainerID string `json:"container_id"`
	Timestamp   int64  `json:"timestamp"`
	Message     string `json:"message"`
}

type LogSearchRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Level  string `json:"level"`
	Source string `json:"source"` // host | container
	Query  string `json:"query"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}
