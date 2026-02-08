package api

import (
	"encoding/json"
	"net/http"
	"opslense-pulse/server/store"
	"time"
)

type HeartbeatPayload struct {
	AccountID string `json:"account_id"`
	AgentID   string `json:"agent_id"`
	Hostname  string `json:"hostname"` // optional if needed
}

func HeartbeatHandler(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var hb HeartbeatPayload
		if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if hb.AccountID == "" || hb.AgentID == "" {
			http.Error(w, "missing account_id or agent_id", http.StatusBadRequest)
			return
		}

		if err := store.UpdateHeartbeat(hb.AccountID, hb.AgentID, hb.Hostname); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"server_time": time.Now().Unix(),
		})
	}
}
