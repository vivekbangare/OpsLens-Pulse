package store

import (
	"context"
	"opslense-pulse/shared"
	"time"
)

type TimelinePoint struct {
	Bucket     time.Time `json:"bucket"`
	Hostname   string    `json:"hostname"`
	SourceType string    `json:"source_type"`
	Total      uint64    `json:"total"`
	Errors     uint64    `json:"errors"`
	Warnings   uint64    `json:"warnings"`
	ErrorRate  float64   `json:"error_rate"`
}

type TimelineResult struct {
	Points  []TimelinePoint `json:"points"`
	Anomaly bool            `json:"anomaly"`
}

type Store interface {
	// Ingest
	SaveMetrics(ctx context.Context, m shared.HostMetrics) error
	InsertLogs(ctx context.Context, batch shared.LogBatch) error
	SaveContainerMetrics(ctx context.Context, m shared.ContainerMetrics) error
	InsertContainerLogs(ctx context.Context, batch shared.ContainerLogBatch) error
	UpsertAgentHeartbeat(ctx context.Context, hb shared.Heartbeat) error
	UpsertAgentMetadata(ctx context.Context, m shared.HostMetrics) error

	// Query
	ListAgents(ctx context.Context, tenantID string) ([]shared.AgentInfo, error)
	GetLatestHostMetrics(ctx context.Context, tenantID string) (map[string]shared.HostMetrics, error)
	GetLogs(
		ctx context.Context,
		tenantID string,
		hostname string,
		agentID string,
		from, to time.Time,
		level string,
		source string,
		source_type string,
		limit int,
	) ([]shared.LogEntry, error)
	GetLogSources(ctx context.Context, tenantID, agentID string) ([]string, error)

	GetLogsTimeline(
		ctx context.Context,
		tenantID string,
		agentID string,
		sourceType string,
		from, to time.Time,
	) (TimelineResult, error)
}
