package log

import "log/slog"

// Error is a syntactic sugar for `slog.String("error", err.Error())`.
func Error(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "<nil>")
	}
	return slog.String("error", err.Error())
}
