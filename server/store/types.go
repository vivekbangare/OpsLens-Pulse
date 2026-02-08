package store

import "opslense-pulse/shared"

type HostState struct {
	IP       string
	Metrics  shared.HostMetrics
	Logs     []shared.LogEntry
	LastSeen int64
}
