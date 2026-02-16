package collector

import (
	"context"
	"fmt"
	"sync"

	"opslense-pulse/agent/config"
)

var stateFlushOnce sync.Once

func StartLogSource(
	ctx context.Context,
	agentID, hostname string,
	src config.LogSource,
	serverURL, apiKey string,
	interval int,
) {

	stateFlushOnce.Do(func() {
		go periodicStateFlush(ctx)
	})
	fmt.Println("📦 Starting log source:", src.Name, "type:", src.Type)

	switch src.Type {
	case "file":
		startFileCollector(
			ctx,
			agentID,
			hostname,
			src,
			serverURL,
			apiKey,
			interval,
		)

	case "directory":
		startDirectoryCollector(
			ctx,
			agentID,
			hostname,
			src,
			serverURL,
			apiKey,
			interval,
		)

	case "command":
		startCommandCollector(
			ctx,
			agentID,
			hostname,
			src,
			serverURL,
			apiKey,
			interval,
		)

	default:
		fmt.Println("⚠️ Unknown log source type:", src.Type)
	}
}
