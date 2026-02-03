package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.38.0"

	otelhost "go.opentelemetry.io/contrib/instrumentation/host"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
)

type Otel struct {
	MeterProvider  *metric.MeterProvider
	TracerProvider *trace.TracerProvider
}

// New bootstraps initialization of tracer and metric
// with [go.opentelemetry.io/otel/sdk/metric.MeterProvider] and [go.opentelemetry.io/otel/sdk/trace.TracerProvider].
//
// It also sets default tracer and metric.
//
// New also initiates a couple of metrics, such as runtime metrics.
func New(ctx context.Context, serviceName string) (Otel, error) {
	// TODO: leverage options

	res, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return Otel{}, fmt.Errorf("resource.New: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithCompressor("gzip"))
	if err != nil {
		return Otel{}, fmt.Errorf("otlptracegrpc.New: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithBatcher(traceExporter),
	)
	otel.SetTracerProvider(tracerProvider)

	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithCompressor("gzip"))
	if err != nil {
		return Otel{}, fmt.Errorf("otlpmetricgrpc.New: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)
	otel.SetMeterProvider(meterProvider)

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)

	// runtime metrics

	err = otelhost.Start()
	if err != nil {
		return Otel{}, fmt.Errorf("otelhost.Start: %w", err)
	}

	err = otelruntime.Start()
	if err != nil {
		return Otel{}, fmt.Errorf("otelruntime.Start: %w", err)
	}

	return Otel{
		MeterProvider:  meterProvider,
		TracerProvider: tracerProvider,
	}, nil
}
