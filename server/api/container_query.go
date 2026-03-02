package api

import (
	"net/http"

	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
)

func ContainerMetricsQueryHandler(st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
		if !ok || tenantID == "" {
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "tenant missing", reqID)
			return
		}

		agentID := r.URL.Query().Get("agent_id")
		if agentID == "" {
			utils.WriteError(w, http.StatusBadRequest, "invalid_request", "agent_id is required", reqID)
			return
		}

		metrics, err := st.FetchContainerMetrics(r.Context(), tenantID, agentID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), reqID)
			return
		}

		utils.WriteJSON(w, http.StatusOK, metrics)
	}
}
