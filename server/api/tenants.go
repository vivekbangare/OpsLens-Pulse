package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"opslense-pulse/server/middleware"
	"opslense-pulse/server/utils"
)

func MyTenantsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		userID, ok := r.Context().Value(middleware.CtxUserID).(string)
		if !ok || userID == "" {
			utils.WriteError(
				w,
				http.StatusInternalServerError,
				"internal_error",
				"user id not found in context",
				reqID,
			)
			return
		}

		rows, err := db.Query(`
			SELECT DISTINCT t.id, t.name, t.slug
			FROM tenants t
			JOIN user_tenants ut ON ut.tenant_id = t.id
			WHERE ut.user_id = $1
			  AND t.is_active = true
		`, userID)
		if err != nil {
			utils.WriteError(
				w,
				http.StatusInternalServerError,
				"internal_error",
				err.Error(),
				reqID,
			)
			return
		}
		defer rows.Close()

		type Tenant struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Slug string `json:"slug"`
		}

		var tenants []Tenant

		for rows.Next() {
			var t Tenant
			if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err == nil {
				tenants = append(tenants, t)
			}
		}
		if err := rows.Err(); err != nil {
			utils.WriteError(
				w,
				http.StatusInternalServerError,
				"internal_error",
				err.Error(),
				reqID,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tenants)
	}
}
