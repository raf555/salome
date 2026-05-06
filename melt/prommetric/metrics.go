package prommetric

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics[T Label] struct {
	counter  *prometheus.CounterVec
	gauge    *prometheus.GaugeVec
	duration *prometheus.HistogramVec
}

func New[T Label](prefix, name string, opts ...Option) Recorder[T] {
	o := &options{
		buckets:    prometheus.DefBuckets,
		registerer: prometheus.DefaultRegisterer,
	}
	for _, opt := range opts {
		opt(o)
	}

	var zeroLabel T
	factory := promauto.With(o.registerer)

	counter := factory.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: prefix,
			Name:      name + "_total",
			Help:      name + " counter",
		},
		zeroLabel.Labels(),
	)

	gauge := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: prefix,
			Name:      name,
			Help:      name + " gauge",
		},
		zeroLabel.Labels(),
	)

	duration := factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: prefix,
			Name:      name + "_duration_seconds",
			Help:      name + " duration in seconds",
			Buckets:   o.buckets,
		},
		zeroLabel.Labels(),
	)

	return &Metrics[T]{
		duration: duration,
		gauge:    gauge,
		counter:  counter,
	}
}

// Count implements [Recorder].
func (m *Metrics[T]) Count(label T) Counter {
	c := m.counter.WithLabelValues(label.Values()...)
	return &counter{c}
}

type counter struct {
	c prometheus.Counter
}

// Add implements [Counter].
func (c *counter) Add(val float64) {
	c.c.Add(val)
}

// Inc implements [Counter].
func (c *counter) Inc() {
	c.c.Inc()
}

// Duration implements [Recorder].
func (m *Metrics[T]) Duration(label T) DurationObserver {
	h := m.duration.WithLabelValues(label.Values()...)
	return &durationObserver{h}
}

type durationObserver struct {
	h prometheus.Observer
}

// Observe implements [DurationObserver].
func (d *durationObserver) Observe(dur time.Duration) {
	d.h.Observe(dur.Seconds())
}

// Gauge implements [Recorder].
func (m *Metrics[T]) Gauge(label T) Gauge {
	g := m.gauge.WithLabelValues(label.Values()...)
	return &gauge{g}
}

type gauge struct {
	g prometheus.Gauge
}

// Set implements [Gauge].
func (g *gauge) Set(val float64) { g.g.Set(val) }

// Inc implements [Gauge].
func (g *gauge) Inc() { g.g.Inc() }

// Dec implements [Gauge].
func (g *gauge) Dec() { g.g.Dec() }

// Add implements [Gauge].
func (g *gauge) Add(val float64) { g.g.Add(val) }

// Sub implements [Gauge].
func (g *gauge) Sub(val float64) { g.g.Sub(val) }
