package metric

import (
	"context"
	"time"
)

type NoopRecorder struct{}

var _ MetricRecorder = NoopRecorder{}

// Count implements [MetricRecorder].
func (n NoopRecorder) Count(ctx context.Context, name string, value int64, opts ...RecordOption) {
}

// Duration implements [MetricRecorder].
func (n NoopRecorder) Duration(ctx context.Context, name string, duration time.Duration, opts ...RecordOption) {
}

// Gauge implements [MetricRecorder].
func (n NoopRecorder) Gauge(ctx context.Context, name string, value float64, opts ...RecordOption) {
}

// RecordOperation implements [MetricRecorder].
func (n NoopRecorder) RecordOperation(ctx context.Context, name string, duration time.Duration, opts ...RecordOption) {
}
