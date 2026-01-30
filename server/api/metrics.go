// server/api/metrics.go
package api

import (
	"encoding/json"
	"net/http"
	"opslense-pulse/server/store"
	"opslense-pulse/shared"
	"strings"
	"time"
)

var expectedToken string

func SetToken(token string) {
	expectedToken = token
}

func validateToken(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token == expectedToken
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if !validateToken(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
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

func HostsHandler(w http.ResponseWriter, r *http.Request) {
	if !validateToken(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(store.GetAll())
}
