package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"opslense-pulse/server/logger"
	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
	"opslense-pulse/server/validation"
	"opslense-pulse/shared"
)

/*
POST /api/logs
*/
func LogsHandler(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
		if !ok || tenantID == "" {
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "tenant missing", reqID)
			return
		}

		var batch shared.LogBatch
		if err := utils.DecodeJSONStrict(r, &batch); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), reqID)
			return
		}

		if err := validation.ValidateLogBatch(&batch); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "validation_error", err.Error(), reqID)
			return
		}

		batch.TenantID = tenantID

		log := middleware.GetLogger(r.Context())
		log.Info("inserting logs",
			zap.String("tenant_id", batch.TenantID),
			zap.String("agent_id", batch.AgentID),
			zap.String("hostname", batch.Hostname),
			zap.Int("count", len(batch.Logs)),
		)

		if err := store.InsertLogs(r.Context(), batch); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", "operation failed", reqID)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

/*
GET /api/logs/fetch
*/
func FetchLogsHandler(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
		if !ok || tenantID == "" {
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "tenant missing", reqID)
			return
		}

		agentID := r.URL.Query().Get("agent_id")
		hostname := r.URL.Query().Get("hostname")
		level := r.URL.Query().Get("level")
		source := r.URL.Query().Get("source")
		sourceType := r.URL.Query().Get("source_type")

		var start, end time.Time

		if v := r.URL.Query().Get("start"); v != "" {
			if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
				start = time.Unix(ts, 0)
			}
		}

		if v := r.URL.Query().Get("end"); v != "" {
			if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
				end = time.Unix(ts, 0)
			}
		}

		limit := 100
		if v := r.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				limit = n
			}
		}

		logs, err := s.GetLogs(
			r.Context(),
			tenantID,
			hostname,
			agentID,
			start,
			end,
			level,
			source,
			sourceType,
			limit,
		)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", "operation failed", reqID)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(logs)
	}
}

func LogSourcesHandler(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
		if !ok || tenantID == "" {
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "tenant missing", reqID)
			return
		}

		agentID := r.URL.Query().Get("agent_id")

		out, err := s.GetLogSources(r.Context(), tenantID, agentID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", "operation failed", reqID)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func SearchLogs(chStore *store.ClickHouseStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST required", reqID)
			return
		}

		var req shared.LogSearchRequest
		if err := utils.DecodeJSONStrict(r, &req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), reqID)
			return
		}

		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
		if !ok || tenantID == "" {
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "tenant missing", reqID)
			return
		}

		results, err := chStore.SearchLogs(r.Context(), tenantID, req)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", "query failed", reqID)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(results)
	})
}

func LogsTimelineHandler(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
		if !ok || tenantID == "" {
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "tenant missing", reqID)
			return
		}

		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")
		agentID := r.URL.Query().Get("agent_id")
		sourceType := r.URL.Query().Get("source_type")

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

		result, err := s.GetLogsTimeline(
			r.Context(),
			tenantID,
			agentID,
			sourceType,
			from,
			to,
		)

		if err != nil {
			logger.Log.Error("timeline query failed",
				zap.String("tenant_id", tenantID),
				zap.Error(err),
			)
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", "operation failed", reqID)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}
