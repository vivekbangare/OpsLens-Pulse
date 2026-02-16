package api

import (
	"encoding/json"
	"net/http"
	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/shared"
)

func ContainerMetricsHandler(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)

		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var metrics shared.ContainerMetrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// 🔒 enforce tenant from auth
		metrics.TenantID = tenantID

		if metrics.TenantID == "" || metrics.ContainerID == "" {
			http.Error(w, "missing tenant_id or container_id", http.StatusBadRequest)
			return
		}
		if err := store.SaveContainerMetrics(metrics); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func ContainerLogsHandler(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)

		if !ok || tenantID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var batch shared.ContainerLogBatch
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		batch.TenantID = tenantID
		if err := store.InsertContainerLogs(batch); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
