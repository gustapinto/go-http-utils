package main

import (
	"log/slog"
	"net"
	"net/http"

	"github.com/gustapinto/go-api-rate-limiter/handler"
	"github.com/gustapinto/go-api-rate-limiter/middleware"
)

const (
	_address = ":5000"
)

func runAPI() error {
	logger := slog.With("context", "main.runAPI")

	mux := http.NewServeMux()
	mid := middleware.NewChain(
		middleware.WithRequestID,
		middleware.Logger,
		middleware.RateLimiter,
	)

	mux.HandleFunc("GET /alive", mid.Handle(handler.Alive))

	listener, err := net.Listen("tcp", _address)
	if err != nil {
		return err
	}

	logger.Info("Listening", "address", _address)

	return http.Serve(listener, mux)
}
