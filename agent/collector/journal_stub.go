//go:build !linux

package collector

import (
	"context"
	"opslense-pulse/agent/queue"
)

func StartJournalCollector(
	ctx context.Context,
	agentID, hostname string,
	logsQueue *queue.FileQueue,
) {
}
