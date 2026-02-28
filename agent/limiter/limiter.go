package limiter

import "golang.org/x/time/rate"

var L = rate.NewLimiter(200, 500) // 200 events/sec

func Allow() bool {
	return L.Allow()
}
