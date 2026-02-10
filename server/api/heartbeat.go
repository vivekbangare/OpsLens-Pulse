package api

import (
	"encoding/json"
	"net/http"
	"time"

	"opslense-pulse/server/store"
	"opslense-pulse/shared"
)

func HeartbeatHandler(st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var hb shared.Heartbeat
		if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
			http.Error(w, "invalid payload", 400)
			return
		}

		if hb.Timestamp.IsZero() {
			hb.Timestamp = time.Now()
		}

		if err := st.UpsertAgentHeartbeat(hb); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
