package log

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

// SetContext attaches [slog.Logger] to the ctx.
func SetContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext retrieves [slog.Logger] from the ctx.
// If it's not available, it will return logger from [slog.Default].
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
