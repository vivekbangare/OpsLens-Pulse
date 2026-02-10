package api

import (
	"encoding/json"
	"log"
	"net/http"
	"opslense-pulse/server/store"
	"opslense-pulse/shared"
	"strconv"
	"time"
)

/*
POST /api/logs
Agent → Server (INSERT)
*/
func LogsHandler(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var batch shared.LogBatch
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		log.Printf(
			"📥 Inserting logs: account=%s agent=%s host=%s count=%d",
			batch.AccountID,
			batch.AgentID,
			batch.Hostname,
			len(batch.Logs),
		)

		if batch.AccountID == "" {
			http.Error(w, "missing account_id", http.StatusBadRequest)
			return
		}
		if batch.AgentID == "" {
			http.Error(w, "missing agent_id", http.StatusBadRequest)
			return
		}
		if batch.Hostname == "" {
			http.Error(w, "missing hostname", http.StatusBadRequest)
			return
		}

		if err := store.InsertLogs(batch); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

/*
GET /api/logs/fetch
UI → Server (QUERY)
*/
func FetchLogsHandler(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// ✅ Query params
		accountID := r.URL.Query().Get("account_id")
		agentID := r.URL.Query().Get("agent_id")
		hostname := r.URL.Query().Get("hostname")
		level := r.URL.Query().Get("level")

		if accountID == "" {
			http.Error(w, "missing account_id", http.StatusBadRequest)
			return
		}

		// Time range
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
			accountID,
			hostname,
			agentID,
			start,
			end,
			level,
			limit,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(logs)
	}
}
