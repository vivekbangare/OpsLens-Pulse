package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"opslense-pulse/server/middleware"
)

func MyTenantsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID, ok := r.Context().Value(middleware.CtxUserID).(string)
		if !ok {
			http.Error(w, "user id not found in context", 500)
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
			http.Error(w, err.Error(), 500)
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

		json.NewEncoder(w).Encode(tenants)
	}
}
