package api

import (
	"encoding/json"
	"net/http"
	"opslense-pulse/server/store"
	"opslense-pulse/shared"
	"strconv"
	"time"
)

func LogsHandler(store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var batch shared.LogBatch
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if batch.AccountID == "" {
			http.Error(w, "missing account_id", http.StatusBadRequest)
			return
		}

		if err := store.InsertLogs(batch); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// FetchLogsHandler returns logs filtered by hostname, time range, and optional limit
func FetchLogsHandler(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		hostname := r.URL.Query().Get("hostname")
		accountID := r.URL.Query().Get("account_id")
		level := r.URL.Query().Get("level")

		startTsStr := r.URL.Query().Get("start")
		endTsStr := r.URL.Query().Get("end")
		limitStr := r.URL.Query().Get("limit")

		var start, end time.Time
		var limit int
		var err error

		if startTsStr != "" {
			ts, err := strconv.ParseInt(startTsStr, 10, 64)
			if err == nil {
				start = time.Unix(ts, 0)
			}
		}
		if endTsStr != "" {
			ts, err := strconv.ParseInt(endTsStr, 10, 64)
			if err == nil {
				end = time.Unix(ts, 0)
			}
		}
		if limitStr != "" {
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				limit = 100
			}
		} else {
			limit = 100
		}

		logs, err := s.GetLogs(hostname, accountID, start, end, level, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	}
}
