package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

const (
	_requestIdHeader  = "X-RequestID"
	_requestIdContext = "requestID"
)

func WithRequestID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(_requestIdHeader)
		if len(requestID) == 0 {
			newID, err := uuid.NewV7()
			if err == nil {
				requestID = newID.String()
			}
		}

		newCtx := context.WithValue(r.Context(), _requestIdContext, requestID)

		w.Header().Add(_requestIdHeader, requestID)

		next.ServeHTTP(w, r.WithContext(newCtx))
	}
}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	logger := slog.With("context", "middleware.Logger")

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(_requestIdContext).(string)

		next.ServeHTTP(w, r)

		statusCode, ok := r.Context().Value("status").(int)
		if !ok {
			statusCode = 0
		}

		level := slog.LevelInfo
		if statusCode >= 300 && statusCode < 500 {
			level = slog.LevelWarn
		} else if statusCode >= 500 {
			level = slog.LevelError
		}

		logger.Log(
			context.Background(),
			level,
			r.Pattern,
			slog.Group("request", "id", requestID, "from", r.RemoteAddr),
			slog.Group("response", "status", statusCode),
		)
	}
}
