package handlers

import (
	"net/http"

	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
	"opslense-pulse/shared"
)

type Handler struct {
	Store *store.ClickHouseStore
}

func (h *Handler) SearchLogs(w http.ResponseWriter, r *http.Request) {

	reqID := middleware.GetRequestID(r.Context())

	var req shared.LogSearchRequest
	if err := utils.DecodeJSONStrict(r, &req); err != nil {
		utils.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_json",
			"invalid payload",
			reqID,
		)
		return
	}

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

	logs, err := h.Store.SearchLogs(r.Context(), tenantID, req)
	if err != nil {
		utils.WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"query failed",
			reqID,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.WriteJSON(w, http.StatusOK, logs)
}
