package api

import (
	"encoding/json"
	"net/http"
	"opslense-pulse/server/store"
	"opslense-pulse/shared"
)

func MetricsHandler(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics shared.HostMetrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if metrics.AccountID == "" {
			http.Error(w, "missing account_id", http.StatusBadRequest)
			return
		}

		if err := store.SaveMetrics(metrics); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
