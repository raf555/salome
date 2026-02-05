package metric

import (
	"context"
)

type ctxKey struct{}

func WithContext(ctx context.Context, defaultMetric Recorder) context.Context {
	return context.WithValue(ctx, ctxKey{}, defaultMetric)
}

func FromContext(ctx context.Context) Recorder {
	if metric, ok := ctx.Value(ctxKey{}).(Recorder); ok {
		return metric
	}
	return NoopRecorder{}
}
