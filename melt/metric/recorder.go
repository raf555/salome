package metric

import (
	"context"
	"time"
)

type Recorder interface {
	Count(ctx context.Context, name string, value int64, opts ...RecordOption)
	Duration(ctx context.Context, name string, duration time.Duration, opts ...RecordOption)
	RecordOperation(ctx context.Context, name string, duration time.Duration, opts ...RecordOption)
	Gauge(ctx context.Context, name string, value float64, opts ...RecordOption)
}
