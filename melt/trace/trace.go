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

func (t *TracerProvider) Tracer(opts ...oteltrace.TracerOption) oteltrace.Tracer {
	tracer := t.tp.Tracer(t.svcName, opts...)
	return &Tracer{Tracer: tracer}
}

type Tracer struct {
	oteltrace.Tracer
}

var _ oteltrace.Tracer = (*Tracer)(nil)

func (t *Tracer) Start(ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	opts2 := make([]oteltrace.SpanStartOption, 0, len(opts)+1)
	opts2 = append(opts2, oteltrace.WithTimestamp(time.Now()))
	opts2 = append(opts2, opts...)

	ctx, span := t.Tracer.Start(ctx, spanName, opts2...)
	return ctx, &Span{
		Span: span,
	}
}

type Span struct {
	oteltrace.Span
}

func (s *Span) RecordError(err error, opts ...oteltrace.EventOption) {
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

func (s *Span) End(opts ...oteltrace.SpanEndOption) {
	opts2 := make([]oteltrace.SpanEndOption, 0, len(opts)+1)
	opts2 = append(opts2, oteltrace.WithTimestamp(time.Now()))
	opts2 = append(opts2, opts...)
	s.Span.End(opts2...)
}
