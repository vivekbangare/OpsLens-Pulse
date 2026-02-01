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

// validateToken validates Authorization header using shared auth token
func validateToken(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token == GetAuthToken()
}

// MetricsHandler receives host metrics from agent
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if !validateToken(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
	if !validateToken(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(store.GetAll())
}
