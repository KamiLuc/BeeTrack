// Package logging configures the process-wide structured logger and carries
// per-request loggers through context.Context.
package logging

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

type ctxKey struct{}

// Init sets the global slog default logger to emit structured JSON to w at
// the given level ("debug", "info", "warn", or "error"; an empty or
// unrecognized value falls back to "info"). Callers pass os.Stdout in
// production; tests pass a buffer.
func Init(w io.Writer, level string) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: parseLevel(level),
	})))
}

// WithContext returns a copy of ctx carrying logger for later retrieval via FromContext.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext returns the logger attached to ctx by WithContext, or
// slog.Default() if ctx carries none.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
