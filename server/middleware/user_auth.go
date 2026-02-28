package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"opslense-pulse/server/auth"
	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
)

func UserAuth(jwtManager *auth.JWTManager, db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			reqID := GetRequestID(r.Context())

			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token", reqID)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")

			claims, err := jwtManager.Validate(tokenStr)
			if err != nil {
				utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid token", reqID)
				return
			}

			tenantID := claims.TenantID
			if tenantID == "" {
				utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "tenant missing in token", reqID)
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
				utils.WriteError(w, http.StatusForbidden, "forbidden", "access denied", reqID)
				return
			}
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, "internal_error", "database error", reqID)
				return
			}

			permissions, err := store.LoadUserPermissions(db, claims.UserID, tenantID)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, "internal_error", "permission load failed", reqID)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUserID, claims.UserID)
			ctx = context.WithValue(ctx, CtxTenantID, tenantID)
			ctx = context.WithValue(ctx, CtxPermissions, permissions)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
