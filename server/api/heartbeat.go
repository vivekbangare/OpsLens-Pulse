package api

import (
	"encoding/json"
	"net/http"
	"time"

	"opslense-pulse/server/store"
	"opslense-pulse/shared"
)

func HeartbeatHandler(w http.ResponseWriter, r *http.Request) {
	var hb shared.Heartbeat
	if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
		http.Error(w, "invalid heartbeat", 400)
		return
	}

	store.UpdateHeartbeat(hb.Hostname, time.Now())
	w.WriteHeader(http.StatusOK)
}
