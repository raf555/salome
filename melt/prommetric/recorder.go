package prommetric

import "time"

type CounterRecorder[T Label] interface {
	Count(label T) Counter
}

type GaugeRecorder[T Label] interface {
	Gauge(label T) Gauge
}

type DurationRecorder[T Label] interface {
	Duration(label T) DurationObserver
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
