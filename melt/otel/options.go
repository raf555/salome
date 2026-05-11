package otel

import (
	"time"

	promclient "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.38.0"
)

type options struct {
	serviceVersion string

	// Sampling. Range [0, 1]. Default 1.0.
	traceSampleRatio float64

	// Metric export.
	metricExportInterval time.Duration // default 10s
	metricExportTimeout  time.Duration // default 30s

	// Trace batching.
	traceMaxQueueSize       int           // default 2048
	traceMaxExportBatchSize int           // default 512
	traceExportTimeout      time.Duration // default 30s
	traceBatchTimeout       time.Duration // default 5s

	// Lifecycle.
	shutdownTimeout time.Duration // default 30s

	// Error handling. If nil, the OTel global error handler is left untouched.
	errorHandler otel.ErrorHandlerFunc

	// Prometheus bridge. When enabled, a prometheus-bridge metric.Producer is
	// registered on the periodic reader using promBridgeGatherer (nil means
	// the prometheus DefaultGatherer).
	promBridgeEnabled  bool
	promBridgeGatherer promclient.Gatherer
}

// Option configures NewOrNoop behavior.
type Option func(*options)

// WithServiceVersion sets the service.version resource attribute.
func WithServiceVersion(version string) Option {
	return func(o *options) {
		o.serviceVersion = version
	}
}

// WithTraceSampleRatio sets the head-based trace sampling ratio.
// Values outside [0, 1] are clamped. Defaults to 1.0 (sample everything).
//
// Wrapped by ParentBased so child spans honor the upstream sampling decision.
func WithTraceSampleRatio(ratio float64) Option {
	return func(o *options) {
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
		o.traceSampleRatio = ratio
	}
}

// WithMetricExportInterval sets how often the metric reader exports.
// Defaults to 10s.
func WithMetricExportInterval(d time.Duration) Option {
	return func(o *options) {
		o.metricExportInterval = d
	}
}

// WithMetricExportTimeout sets the per-export timeout for the metric reader.
// Defaults to 30s. Must be <= the export interval.
func WithMetricExportTimeout(d time.Duration) Option {
	return func(o *options) {
		o.metricExportTimeout = d
	}
}

// WithTraceQueueSize sets the maximum number of spans buffered before drops.
// Defaults to 2048.
func WithTraceQueueSize(n int) Option {
	return func(o *options) {
		o.traceMaxQueueSize = n
	}
}

// WithTraceBatchSize sets the maximum spans per export batch.
// Defaults to 512. Should be <= the queue size.
func WithTraceBatchSize(n int) Option {
	return func(o *options) {
		o.traceMaxExportBatchSize = n
	}
}

// WithTraceExportTimeout sets the per-batch export timeout.
// Defaults to 30s.
func WithTraceExportTimeout(d time.Duration) Option {
	return func(o *options) {
		o.traceExportTimeout = d
	}
}

// WithTraceBatchTimeout sets the maximum delay before an incomplete batch is
// flushed. Defaults to 5s.
func WithTraceBatchTimeout(d time.Duration) Option {
	return func(o *options) {
		o.traceBatchTimeout = d
	}
}

// WithShutdownTimeout caps how long Shutdown will wait for the providers to
// flush and disconnect. Defaults to 30s. The caller's context deadline still
// applies — whichever fires first wins.
func WithShutdownTimeout(d time.Duration) Option {
	return func(o *options) {
		o.shutdownTimeout = d
	}
}

// WithErrorHandler routes SDK-internal errors (failed exports, queue drops,
// etc.) to the given handler instead of the default stderr logger.
func WithErrorHandler(h otel.ErrorHandlerFunc) Option {
	return func(o *options) {
		o.errorHandler = h
	}
}

// WithPrometheusMetricsBridge enables the OTel Prometheus bridge, which
// gathers metrics from a [github.com/prometheus/client_golang/prometheus.Gatherer]
// on every collection cycle and emits them through the OTLP pipeline alongside
// the native OTel metrics.
//
// Pass nil to use [github.com/prometheus/client_golang/prometheus.DefaultGatherer]
// (the global registry that
// [github.com/prometheus/client_golang/prometheus.MustRegister] writes to).
// Pass a specific Gatherer to bridge only a curated subset of metrics.
//
// Typical use case: surface the rich Go runtime metrics emitted by
// [github.com/prometheus/client_golang/prometheus/collectors.NewGoCollector]
// (with GoRuntimeMetricsCollection enabled) via OTLP, removing the need for
// a separate Prometheus scrape pipeline.
//
// IMPORTANT — avoid double-counting:
//
// If your application also exposes a Prometheus scrape endpoint (commonly
// /metrics on the main HTTP server, or a dedicated promhttp listener) AND
// that endpoint is being scraped (by Prometheus, vmagent, the OTel
// Collector's prometheus receiver, etc.) into the same backend that receives
// your OTLP metrics, every metric in the bridged gatherer will arrive twice
// — once via OTLP, once via scrape — with the same name. Pick one path:
//
//   - Bridge (this option) and remove the scrape configuration; or
//   - Keep scraping and do NOT enable this option.
//
// If /metrics must stay up for other consumers (a local debugging probe, a
// liveness check, etc.), use a non-default registry for the bridged metrics
// and pass it explicitly here, keeping the bridged set disjoint from the
// scraped set.
//
// Performance:
//
// The bridge calls Gatherer.Gather() on every metric export cycle (default
// 10s). For a registry containing the full Go runtime collector this is
// cheap in absolute terms but is not free — expect a small, steady increase
// in CPU and memory footprint compared to running without the bridge. Per
// the upstream docs, the bridge typically reduces overall footprint compared
// to running a parallel Prometheus scrape pipeline to the same Collector,
// since the scrape's HTTP serving and serialization cost is eliminated.
//
// See https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/prometheus.
func WithPrometheusMetricsBridge(gatherer promclient.Gatherer) Option {
	return func(o *options) {
		o.promBridgeEnabled = true
		o.promBridgeGatherer = gatherer
	}
}

// newOptions builds an options with defaults already applied, then applies
// the user-supplied options on top.
func newOptions(opts ...Option) *options {
	cfg := &options{
		traceSampleRatio:        1.0,
		metricExportInterval:    10 * time.Second,
		metricExportTimeout:     30 * time.Second,
		traceMaxQueueSize:       2048,
		traceMaxExportBatchSize: 512,
		traceExportTimeout:      30 * time.Second,
		traceBatchTimeout:       5 * time.Second,
		shutdownTimeout:         30 * time.Second,
	}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

func (cfg *options) resourceAttributes(serviceName string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
	}
	if cfg.serviceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersion(cfg.serviceVersion))
	}
	return attrs
}
