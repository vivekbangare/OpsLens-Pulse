package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"opslense-pulse/server/logger"
)

const RequestIDKey contextKey = "request_id"
const LoggerKey contextKey = "logger"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set("X-Request-ID", requestID)

		// Create request-scoped logger
		reqLogger := logger.Log.With(
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, LoggerKey, reqLogger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(RequestIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

func GetLogger(ctx context.Context) *zap.Logger {
	if v := ctx.Value(LoggerKey); v != nil {
		if l, ok := v.(*zap.Logger); ok {
			return l
		}
	}
	return logger.Log
}
