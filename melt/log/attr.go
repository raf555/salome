package log

import "log/slog"

// Error is a syntactic sugar for `slog.String("error", err.Error())`
func Error(err error) slog.Attr {
	return slog.String("error", err.Error())
}
