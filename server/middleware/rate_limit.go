package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"opslense-pulse/server/utils"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

var visitors = make(map[string]*visitor)
var mu sync.Mutex

func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			reqID := GetRequestID(r.Context())

			ip := r.Header.Get("X-Forwarded-For")
			if ip != "" {
				parts := strings.Split(ip, ",")
				ip = strings.TrimSpace(parts[0])
			} else {
				ip = r.RemoteAddr
			}

			mu.Lock()
			v, exists := visitors[ip]
			if !exists {
				v = &visitor{lastSeen: time.Now()}
				visitors[ip] = v
			}

			if time.Since(v.lastSeen) > window {
				v.count = 0
				v.lastSeen = time.Now()
			}

			v.count++
			if v.count > maxRequests {
				mu.Unlock()
				utils.WriteError(w, http.StatusTooManyRequests, "rate_limited", "rate limit exceeded", reqID)
				return
			}
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
