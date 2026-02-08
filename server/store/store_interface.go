package store

import (
	"opslense-pulse/shared"
	"time"
)

type Store interface {
	SaveMetrics(m shared.HostMetrics) error
	InsertLogs(batch shared.LogBatch) error
	UpdateHeartbeat(accountID, agentID, hostname string) error
	GetLogs(accountID, agentID string, from, to time.Time, level string, limit int) ([]shared.LogEntry, error)
	GetFiltered(accountID string, filters map[string]string) ([]map[string]interface{}, error)
	SaveContainerMetrics(m shared.ContainerMetrics) error
	InsertContainerLogs(batch shared.ContainerLogBatch) error
	InsertAPIKey(k shared.APIKey) error
	ValidateAPIKey(rawKey string) (bool, error)
	CountAPIKeys() (int, error)
}
