package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gustapinto/go-api-rate-limiter/bucket"
	"github.com/gustapinto/go-api-rate-limiter/handler"
	"github.com/gustapinto/go-api-rate-limiter/middleware"
	"github.com/lmittmann/tint"
	"github.com/redis/go-redis/v9"
)

const (
	_address      = "0.0.0.0:5000"
	_redisAddress = "127.0.0.1:6379"
)

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.DateTime,
		}),
	))

	logger := slog.With("context", "main.Main")

	if err := run(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	logger := slog.With("context", "main.RunAPI")

	redisBucket := bucket.NewRedis(100, 1, redis.NewClient(&redis.Options{
		Addr: _redisAddress,
	}))
	rateLimiter := middleware.RateLimiter{
		Bucket:               redisBucket,
		BucketErrorBehaviour: middleware.AllowRequestsOnBucketError,
	}

	mid := middleware.NewChain(
		&middleware.RequestID{},
		&middleware.Logger{},
		&rateLimiter,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /alive", mid.Handle(handler.Alive))

	listener, err := net.Listen("tcp", _address)
	if err != nil {
		return err
	}

	logger.Info("Listening", "address", _address)

	return http.Serve(listener, mux)
}
