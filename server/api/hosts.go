package api

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
)

func HostsHandler(st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
		if !ok || tenantID == "" {
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "tenant missing", reqID)
			return
		}

		agents, err := st.ListAgents(r.Context(), tenantID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", "operation failed", reqID)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(agents)
	}
}

func HostSummaryHandler(st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
		if !ok || tenantID == "" {
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "tenant missing", reqID)
			return
		}

		data, err := st.GetLatestHostMetrics(r.Context(), tenantID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", "operation failed", reqID)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(data)
	}
}

func extractPublicIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
