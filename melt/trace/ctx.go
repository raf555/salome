package trace

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type ctxKey struct{}

func WithContext(ctx context.Context, tracer oteltrace.Tracer) context.Context {
	return context.WithValue(ctx, ctxKey{}, tracer)
}

func FromContext(ctx context.Context) oteltrace.Tracer {
	if tracer, ok := ctx.Value(ctxKey{}).(oteltrace.Tracer); ok {
		return tracer
	}
	return noop.Tracer{}
}

func SpanFromContext(ctx context.Context) oteltrace.Span {
	return &Span{
		Span: oteltrace.SpanFromContext(ctx),
	}
}
