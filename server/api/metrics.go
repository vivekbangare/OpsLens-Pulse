package api

import (
	"net/http"

	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
	"opslense-pulse/server/validation"
	"opslense-pulse/shared"
)

func MetricsHandler(store store.Store) http.HandlerFunc {
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

		var metrics shared.HostMetrics
		if err := utils.DecodeJSONStrict(r, &metrics); err != nil {
			utils.WriteError(
				w,
				http.StatusBadRequest,
				"invalid_json",
				err.Error(),
				reqID,
			)
			return
		}

		if err := validation.ValidateHostMetrics(&metrics); err != nil {
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
		metrics.TenantID = tenantID
		metrics.PublicIP = extractPublicIP(r)

		if err := store.UpsertAgentMetadata(r.Context(), metrics); err != nil {
			utils.WriteError(
				w,
				http.StatusInternalServerError,
				"internal_error",
				err.Error(),
				reqID,
			)
			return
		}

		if err := store.SaveMetrics(r.Context(), metrics); err != nil {
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
