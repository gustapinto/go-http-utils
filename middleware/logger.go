package middleware

import (
	"context"
	"log/slog"
	"net/http"
)

type Logger struct{}

func (_ *Logger) Handle(next http.HandlerFunc) http.HandlerFunc {
	logger := slog.With("context", "middleware.Logger")

	return func(w http.ResponseWriter, r *http.Request) {
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
			slog.Group("request", "from", r.RemoteAddr),
			slog.Group("response", "status", statusCode),
		)
	}
}
