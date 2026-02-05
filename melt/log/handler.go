package log

import (
	"log/slog"

	slogotel "github.com/remychantenay/slog-otel"
	slogformatter "github.com/samber/slog-formatter"
)

func WithOtelHandler(h slog.Handler, opts ...slogotel.OtelHandlerOpt) slog.Handler {
	return slogotel.New(h, opts...)
}

func WithFormatter(h slog.Handler, formatters ...slogformatter.Formatter) slog.Handler {
	handler := slogformatter.NewFormatterHandler(formatters...)
	return handler(h)
}

func DurationToStringAttrFormatter() slogformatter.Formatter {
	return slogformatter.FormatByKind(slog.KindDuration, func(v slog.Value) slog.Value {
		return slog.StringValue(v.Duration().String())
	})
}
