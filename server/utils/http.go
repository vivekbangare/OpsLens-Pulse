package utils

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP extracts real client IP considering proxies.
func GetClientIP(r *http.Request) string {

	// X-Forwarded-For may contain multiple IPs: client,proxy1,proxy2
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	// Fallback to remote address
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
