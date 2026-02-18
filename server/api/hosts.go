package api

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
)

func HostsHandler(st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)

		if !ok || tenantID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		agents, err := st.ListAgents(tenantID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(agents)
	}
}

func HostSummaryHandler(st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)

		if !ok || tenantID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		data, err := st.GetLatestHostMetrics(tenantID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(data)
	}
}

func extractPublicIP(r *http.Request) string {
	// Check X-Forwarded-For (can contain multiple IPs)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	// Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
