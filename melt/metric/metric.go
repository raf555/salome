package metric

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

type RecorderProvider struct {
	svcName         string
	mp              *metric.MeterProvider
	defaultMeter    otelmetric.Meter
	defaultRecorder *Recorder
}

func New(serviceName string, provider *metric.MeterProvider) (*RecorderProvider, error) {
	defaultMeter := provider.Meter(serviceName)
	defaultRecorder, err := newRecorder(serviceName, "salome", "default", defaultMeter)
	if err != nil {
		return nil, fmt.Errorf("newRecorder: %w", err)
	}

	rp := &RecorderProvider{
		svcName:         serviceName,
		mp:              provider,
		defaultMeter:    defaultMeter,
		defaultRecorder: defaultRecorder,
	}

	return rp, nil
}

func (r *RecorderProvider) Shutdown(ctx context.Context) error {
	return r.mp.Shutdown(ctx)
}

func (r *RecorderProvider) DefaultRecorder() MetricRecorder {
	return r.defaultRecorder
}

func (r *RecorderProvider) CreateRecorder(prefix, name string) (MetricRecorder, error) {
	return newRecorder(r.svcName, prefix, name, r.defaultMeter)
}

type Recorder struct {
	counter  otelmetric.Int64Counter
	duration otelmetric.Float64Histogram
	gauge    otelmetric.Float64Gauge

	svcName string
}

func newRecorder(svcName, prefix, name string, meter otelmetric.Meter) (*Recorder, error) {
	counter, err := meter.Int64Counter(
		fmt.Sprintf("%s_%s_counter_total", prefix, name),
		otelmetric.WithDescription(name+" counter"),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Int64Counter: %w", err)
	}

	duration, err := meter.Float64Histogram(
		fmt.Sprintf("%s_%s_duration_milliseconds", prefix, name),
		otelmetric.WithDescription(name+" duration (in ms)"),
		otelmetric.WithUnit("ms"),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Histogram: %w", err)
	}

	gauge, err := meter.Float64Gauge(
		fmt.Sprintf("%s_%s_gauge", prefix, name),
		otelmetric.WithDescription(name+" gauge"),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Gauge: %w", err)
	}

	return &Recorder{
		counter:  counter,
		duration: duration,
		gauge:    gauge,
		svcName:  svcName,
	}, nil
}

func (r *Recorder) Count(ctx context.Context, name string, value int64, opts ...RecordOption) {
	opt := buildOptions(opts...)
	labels := r.buildLabels(name, opt.label)

	r.counter.Add(ctx, value, otelmetric.WithAttributeSet(labels))
}

func (r *Recorder) Duration(ctx context.Context, name string, duration time.Duration, opts ...RecordOption) {
	opt := buildOptions(opts...)
	labels := r.buildLabels(name, opt.label)

	r.duration.Record(ctx, float64(duration)/float64(time.Millisecond), otelmetric.WithAttributeSet(labels))
}

func (r *Recorder) RecordOperation(ctx context.Context, name string, duration time.Duration, opts ...RecordOption) {
	opt := buildOptions(opts...)
	labels := r.buildLabels(name, opt.label)

	r.counter.Add(ctx, 1,
		otelmetric.WithAttributeSet(labels))
	r.duration.Record(ctx, float64(duration)/float64(time.Millisecond),
		otelmetric.WithAttributeSet(labels))
}

func (r *Recorder) Gauge(ctx context.Context, name string, value float64, opts ...RecordOption) {
	opt := buildOptions(opts...)
	labels := r.buildLabels(name, opt.label)

	r.gauge.Record(ctx, value, otelmetric.WithAttributeSet(labels))
}

func (r *Recorder) buildLabels(operation string, label Labeler) attribute.Set {
	if label == nil {
		label = NoLabel{}
	}

	labelAttributes := label.Label()

	attributes := make([]attribute.KeyValue, 0, len(labelAttributes)+2)

	attributes = append(attributes, labelAttributes...)
	attributes = append(attributes,
		attribute.String("recorded_name", operation),
		attribute.String("service_name", r.svcName),
	)

	return attribute.NewSet(attributes...)
}
