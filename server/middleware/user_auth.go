package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"opslense-pulse/server/auth"
	"opslense-pulse/server/store"
)

func UserAuth(jwtManager *auth.JWTManager, db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")

			claims, err := jwtManager.Validate(tokenStr)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			tenantID := claims.TenantID
			if tenantID == "" {
				http.Error(w, "tenant missing in token", http.StatusUnauthorized)
				return
			}

			err = db.QueryRow(`
				SELECT t.id
				FROM tenants t
				JOIN user_tenants ut ON ut.tenant_id = t.id
				WHERE ut.user_id = $1
				AND t.id = $2
				AND t.is_active = true
			`, claims.UserID, tenantID).Scan(&tenantID)

			if err == sql.ErrNoRows {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			permissions, err := store.LoadUserPermissions(db, claims.UserID, tenantID)
			if err != nil {
				http.Error(w, "permission load failed", 500)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUserID, claims.UserID)
			ctx = context.WithValue(ctx, CtxTenantID, tenantID)
			ctx = context.WithValue(ctx, CtxPermissions, permissions)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
