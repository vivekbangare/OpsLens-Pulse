package collector

import "time"

// retrySend retries the provided function with exponential backoff
func retrySend(attempts int, baseDelay time.Duration, fn func() error) error {
	delay := baseDelay
	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			time.Sleep(delay)
			delay *= 2
		} else {
			return nil
		}
	}
	return nil
}
