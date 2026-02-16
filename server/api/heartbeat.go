package api

import (
	"encoding/json"
	"net/http"
	"time"

	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/shared"
)

func HeartbeatHandler(st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)

		if !ok || tenantID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var hb shared.Heartbeat
		if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
			http.Error(w, "invalid payload", 400)
			return
		}

		// 🔒 enforce tenant from auth
		hb.TenantID = tenantID

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
