package log

import (
	"log/slog"
)

func New(h slog.Handler) *slog.Logger {
	log := slog.New(h)

	slog.SetDefault(log)
	return log
}
