// Package logger builds the application's structured logger.
package logger

import (
	"log/slog"
	"os"
)

// New returns a JSON logger in production and a human-readable one otherwise.
func New(env string) *slog.Logger {
	var h slog.Handler
	if env == "production" {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	return slog.New(h)
}
