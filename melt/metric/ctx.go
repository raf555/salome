package metric

import (
	"context"
)

type ctxKey struct{}

func WithContext(ctx context.Context, defaultMetric MetricRecorder) context.Context {
	return context.WithValue(ctx, ctxKey{}, defaultMetric)
}

func FromContext(ctx context.Context) MetricRecorder {
	if metric, ok := ctx.Value(ctxKey{}).(MetricRecorder); ok {
		return metric
	}
	return NoopRecorder{}
}
