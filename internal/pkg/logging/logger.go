package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// InitLogger initializes the global structured logger.
// It parses the level from the string (DEBUG, INFO, WARN, ERROR).
func InitLogger(levelStr string) *slog.Logger {
	var level slog.Level
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Use JSON handler for production logging
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// WithContext returns a logger decorated with context values if needed (e.g. correlation IDs).
func WithContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
	// Add trace context or trace ID if present (could be integrated with OpenTelemetry)
	return logger
}
