package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

const (
	_logLevel = slog.LevelInfo
)

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      _logLevel,
			TimeFormat: time.Kitchen,
		}),
	))

	logger := slog.With("context", "main.Main")

	if err := runAPI(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
