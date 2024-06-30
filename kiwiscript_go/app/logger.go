package app

import (
	"log/slog"
	"os"
)

func DefaultLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func GetLogger(env string, debug bool) *slog.Logger {
	logLevel := slog.LevelInfo

	if debug {
		logLevel = slog.LevelDebug
	}
	if env == "production" {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}
