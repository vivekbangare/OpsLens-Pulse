package api

import (
	"net/http"
	"time"

	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
	"opslense-pulse/server/validation"
	"opslense-pulse/shared"
)

func HeartbeatHandler(st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
		if !ok || tenantID == "" {
			utils.WriteError(
				w,
				http.StatusUnauthorized,
				"unauthorized",
				"tenant missing",
				reqID,
			)
			return
		}

		var hb shared.Heartbeat
		if err := utils.DecodeJSONStrict(r, &hb); err != nil {
			utils.WriteError(
				w,
				http.StatusBadRequest,
				"invalid_json",
				err.Error(),
				reqID,
			)
			return
		}

		if err := validation.ValidateHeartbeat(&hb); err != nil {
			utils.WriteError(
				w,
				http.StatusBadRequest,
				"validation_error",
				err.Error(),
				reqID,
			)
			return
		}

		// 🔒 Enforce tenant from auth
		hb.TenantID = tenantID

		if hb.Timestamp == 0 {
			hb.Timestamp = time.Now().Unix()
		}

		if err := st.UpsertAgentHeartbeat(r.Context(), hb); err != nil {
			utils.WriteError(
				w,
				http.StatusInternalServerError,
				"internal_error",
				err.Error(),
				reqID,
			)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
