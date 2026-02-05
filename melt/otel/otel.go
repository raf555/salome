package otel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/contrib/detectors/autodetect"
	otelhost "go.opentelemetry.io/contrib/instrumentation/host"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.38.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Otel struct {
	meterProvider  *metric.MeterProvider
	tracerProvider *trace.TracerProvider
}

// New bootstraps initialization of tracer and metric
// with [go.opentelemetry.io/otel/sdk/metric.MeterProvider] and [go.opentelemetry.io/otel/sdk/trace.TracerProvider].
//
// It also sets default tracer and metric.
//
// New also initiates a couple of metrics, such as runtime metrics.
//
// New detects environment variable of `OTEL_EXPORTER_OTLP_ENDPOINT`. If it's not present, New returns [NoopOpenTelemetry].
func New(ctx context.Context, serviceName string) (OpenTelemetry, error) {
	// TODO: leverage options

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return NoopOpenTelemetry{}, nil
	}

	detectors, err := autodetect.Detector(autodetect.Registered()...)
	if err != nil {
		return nil, fmt.Errorf("autodetect.Detector: %w", err)
	}

	res, err := resource.New(
		ctx,
		resource.WithDetectors(detectors),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("resource.New: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithCompressor("gzip"))
	if err != nil {
		return nil, fmt.Errorf("otlptracegrpc.New: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithBatcher(traceExporter),
	)

	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithCompressor("gzip"))
	if err != nil {
		return nil, fmt.Errorf("otlpmetricgrpc.New: %w", err)
	}

	var periodicReaderOptions []metric.PeriodicReaderOption
	if os.Getenv("OTEL_METRIC_EXPORT_INTERVAL") == "" { // use 10s for default to avoid large batch size on each export
		periodicReaderOptions = append(periodicReaderOptions, metric.WithInterval(10*time.Second))
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter, periodicReaderOptions...)),
	)

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagator)

	// runtime metrics

	err = otelhost.Start()
	if err != nil {
		return nil, fmt.Errorf("otelhost.Start: %w", err)
	}

	err = otelruntime.Start()
	if err != nil {
		return nil, fmt.Errorf("otelruntime.Start: %w", err)
	}

	return Otel{
		meterProvider:  meterProvider,
		tracerProvider: tracerProvider,
	}, nil
}

func (o Otel) Shutdown(ctx context.Context) error {
	var errs []error

	if err := o.meterProvider.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	if err := o.tracerProvider.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (o Otel) TracerProvider() oteltrace.TracerProvider {
	return o.tracerProvider
}

func (o Otel) MeterProvider() otelmetric.MeterProvider {
	return o.meterProvider
}
