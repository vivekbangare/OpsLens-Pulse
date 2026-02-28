package api

import (
	"net/http"

	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
	"opslense-pulse/server/validation"
	"opslense-pulse/shared"
)

func ContainerMetricsHandler(store store.Store) http.HandlerFunc {
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

		var metrics shared.ContainerMetrics
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

		if err := validation.ValidateContainerMetrics(&metrics); err != nil {
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

		if metrics.ContainerID == "" {
			utils.WriteError(
				w,
				http.StatusBadRequest,
				"validation_error",
				"container_id is required",
				reqID,
			)
			return
		}

		if err := store.SaveContainerMetrics(r.Context(), metrics); err != nil {
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

func ContainerLogsHandler(store store.Store) http.HandlerFunc {
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

		var batch shared.ContainerLogBatch
		if err := utils.DecodeJSONStrict(r, &batch); err != nil {
			utils.WriteError(
				w,
				http.StatusBadRequest,
				"invalid_json",
				err.Error(),
				reqID,
			)
			return
		}

		// 🔒 Enforce tenant from auth
		batch.TenantID = tenantID

		if err := store.InsertContainerLogs(r.Context(), batch); err != nil {
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
