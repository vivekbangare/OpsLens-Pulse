package api

import (
	"net/http"
	"time"

	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
)

func AuditLogsHandler(pgStore *store.PostgresStore) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
		if !ok || tenantID == "" {
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "tenant missing", reqID)
			return
		}

		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")

		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid from timestamp", reqID)
			return
		}

		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid to timestamp", reqID)
			return
		}

		userID := r.URL.Query().Get("user_id")
		action := r.URL.Query().Get("action")
		status := r.URL.Query().Get("status")

		logs, err := pgStore.GetAuditLogs(
			r.Context(),
			tenantID,
			from,
			to,
			userID,
			action,
			status,
			100,
		)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", "query failed", reqID)
			return
		}

		utils.WriteSuccess(w, http.StatusOK, logs, "", reqID)
	}
}
