package store

import (
	"opslense-pulse/shared"
	"time"
)

type Store interface {
	// Ingest
	SaveMetrics(m shared.HostMetrics) error
	InsertLogs(batch shared.LogBatch) error
	SaveContainerMetrics(m shared.ContainerMetrics) error
	InsertContainerLogs(batch shared.ContainerLogBatch) error
	UpsertAgentHeartbeat(hb shared.Heartbeat) error
	UpsertAgentMetadata(m shared.HostMetrics) error

	// Query
	ListAgents(tenantID string) ([]shared.AgentInfo, error)
	GetLatestHostMetrics(tenantID string) (map[string]shared.HostMetrics, error)
	GetLogs(
		tenantID string,
		hostname string,
		agentID string,
		from, to time.Time,
		level string,
		source string,
		limit int,
	) ([]shared.LogEntry, error)
	GetLogSources(tenantID, agentID string) ([]string, error)
}
