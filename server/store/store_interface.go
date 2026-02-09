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
    ListAgents(accountID string) ([]shared.AgentInfo, error)

	// Query
	GetLogs(
		accountID string,
		hostname string,
		agentID string,
		from, to time.Time,
		level string,
		limit int,
	) ([]shared.LogEntry, error)

	// API keys
	InsertAPIKey(k shared.APIKey) error
	ValidateAPIKey(rawKey string) (bool, error)
	CountAPIKeys() (int, error)
}
