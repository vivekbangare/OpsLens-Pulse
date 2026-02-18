package api

import (
	"encoding/json"
	"net/http"
	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/shared"
)

func MetricsHandler(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)

		if !ok || tenantID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var metrics shared.HostMetrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// 🔒 enforce tenant from auth
		metrics.TenantID = tenantID
		metrics.PublicIP = extractPublicIP(r)

		if err := store.UpsertAgentMetadata(metrics); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if err := store.SaveMetrics(metrics); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
