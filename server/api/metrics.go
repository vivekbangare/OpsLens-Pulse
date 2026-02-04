// server/api/metrics.go
package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"opslense-pulse/server/store"
	"opslense-pulse/shared"
)

// MetricsHandler receives host metrics from agent

func MetricsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var m shared.HostMetrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	store.SaveMetrics(m)
	store.UpdateHeartbeat(m.Hostname, time.Now())

	w.WriteHeader(http.StatusOK)
}

// HostsHandler returns all known hosts with metrics
func HostsHandler(w http.ResponseWriter, r *http.Request) {

	filters := map[string]string{}
	for key, values := range r.URL.Query() {
		if strings.HasPrefix(key, "tag.") && len(values) > 0 {
			tagKey := strings.TrimPrefix(key, "tag.")
			filters[tagKey] = values[0]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(store.GetFiltered(filters))
}
