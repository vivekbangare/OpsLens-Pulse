package handlers

import (
	"encoding/json"
	"net/http"
	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/shared"
)

type Handler struct {
	Store *store.ClickHouseStore
}

func (h *Handler) SearchLogs(w http.ResponseWriter, r *http.Request) {
	var req shared.LogSearchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
	if !ok || tenantID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	logs, err := h.Store.SearchLogs(tenantID, req)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
