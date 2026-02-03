package otel

import (
	"context"

	otelmetric "go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type OpenTelemetry interface {
	MeterProvider() otelmetric.MeterProvider
	TracerProvider() oteltrace.TracerProvider
	Shutdown(ctx context.Context) error
}
