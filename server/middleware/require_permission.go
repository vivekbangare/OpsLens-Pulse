package middleware

import (
	"net/http"

	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
)

func RequirePermission(
	permission string,
	pgStore *store.PostgresStore,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			reqID := GetRequestID(r.Context())

			perms, ok := r.Context().Value(CtxPermissions).(map[string]bool)
			if !ok {

				utils.WriteError(
					w,
					http.StatusForbidden,
					"forbidden",
					"permissions not found",
					reqID,
				)
				return
			}

			if !perms[permission] {

				// Extract context values
				userID, _ := r.Context().Value(CtxUserID).(string)
				tenantID, _ := r.Context().Value(CtxTenantID).(string)

				go pgStore.InsertAuditLog(r.Context(), store.AuditLog{
					TenantID:  tenantID,
					UserID:    &userID,
					Action:    "permission_denied",
					Status:    "failure",
					IPAddress: utils.GetClientIP(r),
					UserAgent: r.UserAgent(),
					Metadata: map[string]interface{}{
						"permission": permission,
						"path":       r.URL.Path,
						"method":     r.Method,
					},
				})

				utils.WriteError(
					w,
					http.StatusForbidden,
					"forbidden",
					"insufficient permissions",
					reqID,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
