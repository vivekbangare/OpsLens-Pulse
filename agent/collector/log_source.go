package collector

import (
	"context"
	"fmt"

	"opslense-pulse/agent/config"
)

func StartLogSource(
	ctx context.Context,
	accountID, agentID, hostname string,
	src config.LogSource,
	serverURL, apiKey string,
	interval int,
) {
	fmt.Println("📦 Starting log source:", src.Name, "type:", src.Type)

	switch src.Type {
	case "file":
		startFileCollector(
			ctx,
			accountID,
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
			accountID,
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
			accountID,
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
