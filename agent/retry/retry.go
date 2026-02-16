package retry

import (
	"fmt"
	"time"
)

func Do(attempts int, baseDelay time.Duration, fn func() error) error {
	delay := baseDelay

	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			time.Sleep(delay)
			delay *= 2
		} else {
			return nil
		}
	}

	return fmt.Errorf("all retries failed")
}
