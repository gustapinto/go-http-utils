package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type BucketErrorBehaviour int

const (
	AllowRequestsOnBucketError BucketErrorBehaviour = iota + 1
	DenyRequestsOnBucketError
)

type RateLimiterBucket interface {
	Allow(key string) (bool, error)
}

type RateLimiter struct {
	Bucket               RateLimiterBucket
	BucketErrorBehaviour BucketErrorBehaviour
}

func (rl *RateLimiter) Handle(next http.HandlerFunc) http.HandlerFunc {
	if rl.BucketErrorBehaviour == 0 {
		rl.BucketErrorBehaviour = DenyRequestsOnBucketError
	}

	logger := slog.With("context", "middleware.RateLimiter")

	return func(w http.ResponseWriter, r *http.Request) {
		requestKey := r.RemoteAddr
		requestKeyType := "RemoteAddress"
		if token := r.Header.Get("Authorization"); len(token) > 0 {
			requestKey = strings.TrimSpace(strings.ReplaceAll(token, "Bearer", ""))
			requestKeyType = "BearerToken"
		}

		logger.Debug(
			"Limit middleware",
			slog.Group("request", "key", requestKey, "keyType", requestKeyType, "context", "RateLimiter"),
		)

		allow, err := rl.Bucket.Allow(requestKey)
		if err != nil {
			switch rl.BucketErrorBehaviour {
			case AllowRequestsOnBucketError:
				rl.allow(w, r, next)
				return

			case DenyRequestsOnBucketError:
				rl.deny(w, r)
				return
			}
		}

		if !allow {
			rl.deny(w, r)
			return
		}

		rl.allow(w, r, next)
	}
}

func (rl *RateLimiter) allow(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next.ServeHTTP(w, r)
}

func (rl *RateLimiter) deny(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "status", http.StatusTooManyRequests)
	*r = *(r.WithContext(ctx))

	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(`{"error": "Too many requests"}`))
}
