package otel

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

type NoopOpenTelemetry struct{}

var _ OpenTelemetry = NoopOpenTelemetry{}

// MeterProvider implements [OpenTelemetry].
func (n NoopOpenTelemetry) MeterProvider() metric.MeterProvider {
	return metricnoop.NewMeterProvider()
}

// TracerProvider implements [OpenTelemetry].
func (n NoopOpenTelemetry) TracerProvider() trace.TracerProvider {
	return tracenoop.NewTracerProvider()
}

// Shutdown implements [OpenTelemetry].
func (n NoopOpenTelemetry) Shutdown(ctx context.Context) error {
	return ctx.Err()
}
