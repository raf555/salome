package otel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	promclient "github.com/prometheus/client_golang/prometheus"
	prombridge "go.opentelemetry.io/contrib/bridges/prometheus"
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
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Otel struct {
	meterProvider   *metric.MeterProvider
	tracerProvider  *trace.TracerProvider
	shutdownTimeout time.Duration
}

// NewOrNoop bootstraps initialization of OpenTelemetry tracer and metric
// with [go.opentelemetry.io/otel/sdk/metric.MeterProvider] and
// [go.opentelemetry.io/otel/sdk/trace.TracerProvider].
//
// It also sets the global tracer and meter providers and registers Go
// runtime metrics.
//
// NewOrNoop checks for the OTEL_EXPORTER_OTLP_ENDPOINT environment variable.
// If unset, it returns NoopOpenTelemetry without doing anything else.
func NewOrNoop(ctx context.Context, serviceName string, opts ...Option) (OpenTelemetry, error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return NoopOpenTelemetry{}, nil
	}

	cfg := newOptions(opts...)

	if cfg.errorHandler != nil {
		otel.SetErrorHandler(cfg.errorHandler)
	}

	res, err := resource.New(
		ctx,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(cfg.resourceAttributes(serviceName)...),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("resource.New: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithCompressor("gzip"),
		otlptracegrpc.WithTimeout(cfg.traceExportTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("otlptracegrpc.New: %w", err)
	}

	sampler := trace.ParentBased(trace.TraceIDRatioBased(cfg.traceSampleRatio))

	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithSampler(sampler),
		trace.WithBatcher(
			traceExporter,
			trace.WithMaxQueueSize(cfg.traceMaxQueueSize),
			trace.WithMaxExportBatchSize(cfg.traceMaxExportBatchSize),
			trace.WithBatchTimeout(cfg.traceBatchTimeout),
			trace.WithExportTimeout(cfg.traceExportTimeout),
		),
	)

	metricExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithCompressor("gzip"),
		otlpmetricgrpc.WithTimeout(cfg.metricExportTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("otlpmetricgrpc.New: %w", err)
	}

	readerOpts := []metric.PeriodicReaderOption{
		metric.WithInterval(cfg.metricExportInterval),
		metric.WithTimeout(cfg.metricExportTimeout),
		metric.WithProducer(otelruntime.NewProducer()),
	}
	if cfg.promBridgeEnabled {
		gatherer := cfg.promBridgeGatherer
		if gatherer == nil {
			gatherer = promclient.DefaultGatherer
		}
		readerOpts = append(readerOpts,
			metric.WithProducer(prombridge.NewMetricProducer(prombridge.WithGatherer(gatherer))),
		)
	}

	periodicReader := metric.NewPeriodicReader(metricExporter, readerOpts...)

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(periodicReader),
	)

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagator)

	if err := otelhost.Start(); err != nil {
		_ = tracerProvider.Shutdown(ctx)
		_ = meterProvider.Shutdown(ctx)
		return nil, fmt.Errorf("otelhost.Start: %w", err)
	}

	if err := otelruntime.Start(); err != nil {
		_ = tracerProvider.Shutdown(ctx)
		_ = meterProvider.Shutdown(ctx)
		return nil, fmt.Errorf("otelruntime.Start: %w", err)
	}

	return &Otel{
		meterProvider:   meterProvider,
		tracerProvider:  tracerProvider,
		shutdownTimeout: cfg.shutdownTimeout,
	}, nil
}

// Shutdown flushes and stops the trace and metric pipelines. It applies the
// configured shutdown timeout (default 30s) on top of the caller's context;
// whichever deadline fires first wins. Errors from both pipelines are joined.
func (o *Otel) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, o.shutdownTimeout)
	defer cancel()

	var errs []error
	if err := o.meterProvider.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("meterProvider.Shutdown: %w", err))
	}
	if err := o.tracerProvider.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("tracerProvider.Shutdown: %w", err))
	}
	return errors.Join(errs...)
}

func (o *Otel) TracerProvider() oteltrace.TracerProvider { return o.tracerProvider }
func (o *Otel) MeterProvider() otelmetric.MeterProvider  { return o.meterProvider }
