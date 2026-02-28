package store

import (
	"context"
	"log"
	"time"
)

const maxBatchSize = 2000
const maxQueryLimit = 1000

////////////////////////////////////////////////////////////
// Helpers
////////////////////////////////////////////////////////////

func boolToUInt8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func resolveTimestamp(epoch int64) time.Time {
	if epoch == 0 {
		return time.Now()
	}
	if epoch > 1e12 {
		return time.UnixMilli(epoch)
	}
	return time.Unix(epoch, 0)
}

func (c *ClickHouseStore) execWithRetry(
	parentCtx context.Context,
	query string,
	args ...interface{},
) error {

	const maxRetries = 5
	baseDelay := 500 * time.Millisecond
	timeoutPerAttempt := 5 * time.Second

	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {

		if parentCtx.Err() != nil {
			return parentCtx.Err()
		}

		ctx, cancel := context.WithTimeout(parentCtx, timeoutPerAttempt)

		_, err = c.db.ExecContext(ctx, query, args...)
		cancel()

		if err == nil {
			return nil
		}

		log.Printf("ClickHouse write failed (attempt %d/%d): %v",
			attempt, maxRetries, err)

		time.Sleep(baseDelay * time.Duration(attempt))
	}

	return err
}

func chunkLogs[T any](items []T, size int) [][]T {
	if len(items) == 0 {
		return nil
	}
	var chunks [][]T

	for size < len(items) {
		items, chunks = items[size:], append(chunks, items[0:size:size])
	}

	chunks = append(chunks, items)
	return chunks
}

func (c *ClickHouseStore) StartHealthMonitor() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err := c.db.PingContext(ctx)
			cancel()

			if err != nil {
				c.health.Store(false)
				log.Println("ClickHouse unhealthy:", err)
			} else {
				c.health.Store(true)
			}
		}
	}()
}
