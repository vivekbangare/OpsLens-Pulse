package middleware

import (
	"context"
	"net/http"
	"strings"

	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
)

func AgentAuth(validator store.APIKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			reqID := GetRequestID(r.Context())

			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing api key", reqID)
				return
			}

			rawKey := strings.TrimPrefix(header, "Bearer ")

			ok, tenantID, err := validator.ValidateAPIKey(rawKey)
			if err != nil || !ok {
				utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid api key", reqID)
				return
			}

			ctx := context.WithValue(r.Context(), CtxTenantID, tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
