package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/gustapinto/go-api-rate-limiter/bucket"
	"github.com/redis/go-redis/v9"
)

const (
	_redisAddress = "localhost:6379"
)

var (
	_limitBucket     *bucket.Redis
	_limitBucketOnce sync.Once
)

func RateLimiter(next http.HandlerFunc) http.HandlerFunc {
	logger := slog.With("context", "middleware.RateLimiter")

	_limitBucketOnce.Do(func() {
		_limitBucket = bucket.NewRedis(5, 1, redis.NewClient(&redis.Options{
			Addr: _redisAddress,
		}))
	})

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(_requestIdContext).(string)

		requestKey := r.RemoteAddr
		requestKeyType := "RemoteAddress"
		if token := r.Header.Get("Authorization"); len(token) > 0 {
			requestKey = strings.TrimSpace(strings.ReplaceAll(token, "Bearer", ""))
			requestKeyType = "BearerToken"
		}

		logger.Debug(
			"Limit middleware",
			slog.Group("request", "id", requestID, "key", requestKey, "keyType", requestKeyType, "context", "RateLimiter"),
		)

		if !_limitBucket.Allow(requestKey) {
			ctx := context.WithValue(r.Context(), "status", http.StatusTooManyRequests)
			*r = *(r.WithContext(ctx))

			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Too many requests"}`))
			return
		}

		next.ServeHTTP(w, r)
	}
}
