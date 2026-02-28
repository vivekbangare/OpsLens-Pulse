package middleware

import (
	"net/http"

	"opslense-pulse/server/utils"

	"go.uber.org/zap"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := GetRequestID(r.Context())
		defer func() {
			if rec := recover(); rec != nil {
				logger := GetLogger(r.Context())
				if logger == nil {
					logger = zap.L()
				}
				logger.Error("panic recovered",
					zap.Any("error", rec),
				)
				utils.WriteError(
					w,
					http.StatusInternalServerError,
					"internal_error",
					"unexpected server error",
					reqID,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
