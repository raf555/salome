package prommetric

import (
	"time"
)

type Recorder[T Label] interface {
	Count(label T) Counter
	Duration(label T) DurationObserver
	Gauge(label T) Gauge
}

type RecorderNoLabel interface {
	Count() Counter
	Duration() DurationObserver
	Gauge() Gauge
}

type Counter interface {
	Inc()
	Add(val float64)
}

type DurationObserver interface {
	Observe(duration time.Duration)
}

type Gauge interface {
	Set(val float64)
	Inc()
	Dec()
	Add(val float64)
	Sub(val float64)
}
