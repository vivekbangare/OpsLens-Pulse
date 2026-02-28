//go:build !windows

package collector

import (
	"context"
	"opslense-pulse/agent/queue"
)

func StartWindowsEventCollector(
	ctx context.Context,
	agentID, hostname string,
	channel string,
	logsQueue *queue.FileQueue,
) {
	// No-op on non-windows platforms
}
