package trace

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type ctxKey struct{}

func WithContext(ctx context.Context, tracer Tracer) context.Context {
	return context.WithValue(ctx, ctxKey{}, tracer)
}

func FromContext(ctx context.Context) Tracer {
	if tracer, ok := ctx.Value(ctxKey{}).(Tracer); ok {
		return tracer
	}
	return noop.Tracer{}
}

func SpanFromContext(ctx context.Context) Span {
	otelspan := oteltrace.SpanFromContext(ctx)

	if s, ok := otelspan.(*span); ok {
		return s
	}

	return &span{
		Span: oteltrace.SpanFromContext(ctx),
	}
}
