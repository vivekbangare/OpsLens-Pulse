package middleware

import (
	"context"
	"net/http"
	"opslense-pulse/server/store"
	"strings"
)

func AgentAuth(validator store.APIKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			rawKey := strings.TrimPrefix(header, "Bearer ")

			ok, tenantID, err := validator.ValidateAPIKey(rawKey)
			if err != nil || !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CtxTenantID, tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
