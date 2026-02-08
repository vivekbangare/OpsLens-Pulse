package api

import (
	"encoding/json"
	"net/http"
	"opslense-pulse/server/store"
	"opslense-pulse/shared"
)

func ContainerMetricsHandler(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics shared.ContainerMetrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if metrics.AccountID == "" || metrics.ContainerID == "" {
			http.Error(w, "missing account_id or container_id", http.StatusBadRequest)
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
		var batch shared.ContainerLogBatch
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if batch.AccountID == "" {
			http.Error(w, "missing account_id", http.StatusBadRequest)
			return
		}
		if err := store.InsertContainerLogs(batch); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
