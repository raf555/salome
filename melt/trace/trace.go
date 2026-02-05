package trace

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type TracerProvider struct {
	svcName string
	tp      oteltrace.TracerProvider
}

func New(serviceName string, tracer oteltrace.TracerProvider) *TracerProvider {
	return &TracerProvider{
		svcName: serviceName,
		tp:      tracer,
	}
}

func (t *TracerProvider) Tracer(opts ...oteltrace.TracerOption) Tracer {
	traceStarter := t.tp.Tracer(t.svcName, opts...)
	return &tracer{Tracer: traceStarter}
}

type tracer struct {
	Tracer
}

var _ oteltrace.Tracer = (*tracer)(nil)

func (t *tracer) Start(ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption) (context.Context, Span) {
	opts2 := make([]oteltrace.SpanStartOption, 0, len(opts)+1)
	opts2 = append(opts2, oteltrace.WithTimestamp(time.Now()))
	opts2 = append(opts2, opts...)

	ctx, otelspan := t.Tracer.Start(ctx, spanName, opts2...)

	osp := &span{
		Span: otelspan,
	}
	ctx = oteltrace.ContextWithSpan(ctx, osp)
	return ctx, osp
}

type span struct {
	Span
}

var _ oteltrace.Span = (*span)(nil)

func (s *span) RecordError(err error, opts ...oteltrace.EventOption) {
	if err == nil {
		return
	}

	opts2 := make([]oteltrace.EventOption, 0, len(opts)+2)
	opts2 = append(opts2,
		oteltrace.WithTimestamp(time.Now()),
		oteltrace.WithStackTrace(true),
	)
	opts2 = append(opts2, opts...)

	s.SetStatus(codes.Error, err.Error())
	s.Span.RecordError(err, opts2...)
}

func (s *span) End(opts ...oteltrace.SpanEndOption) {
	opts2 := make([]oteltrace.SpanEndOption, 0, len(opts)+1)
	opts2 = append(opts2, oteltrace.WithTimestamp(time.Now()))
	opts2 = append(opts2, opts...)
	s.Span.End(opts2...)
}
