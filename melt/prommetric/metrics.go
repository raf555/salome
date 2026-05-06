package prommetric

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// CounterMetrics is a labeled Prometheus counter vec.
type CounterMetrics[T Label] struct {
	vec *prometheus.CounterVec
}

func NewCounterWithLabel[T Label](prefix, name string, opts ...CounterOption) CounterRecorder[T] {
	o := &counterOptions{registerer: prometheus.DefaultRegisterer}
	for _, opt := range opts {
		opt(o)
	}
	var zero T
	vec := promauto.With(o.registerer).NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: prefix,
			Name:      name + "_total",
			Help:      name + " counter",
		},
		zero.Labels(),
	)
	return &CounterMetrics[T]{vec: vec}
}

func (m *CounterMetrics[T]) Count(label T) Counter {
	return &counter{m.vec.WithLabelValues(label.Values()...)}
}

// NewCounter returns a Counter with no labels.
func NewCounter(prefix, name string, opts ...CounterOption) Counter {
	return NewCounterWithLabel[*NoLabel](prefix, name, opts...).Count(nil)
}

// GaugeMetrics is a labeled Prometheus gauge vec.
type GaugeMetrics[T Label] struct {
	vec *prometheus.GaugeVec
}

func NewGaugeWithLabel[T Label](prefix, name string, opts ...GaugeOption) GaugeRecorder[T] {
	o := &gaugeOptions{registerer: prometheus.DefaultRegisterer}
	for _, opt := range opts {
		opt(o)
	}
	var zero T
	vec := promauto.With(o.registerer).NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: prefix,
			Name:      name,
			Help:      name + " gauge",
		},
		zero.Labels(),
	)
	return &GaugeMetrics[T]{vec: vec}
}

func (m *GaugeMetrics[T]) Gauge(label T) Gauge {
	return &gauge{m.vec.WithLabelValues(label.Values()...)}
}

// NewGauge returns a Gauge with no labels.
func NewGauge(prefix, name string, opts ...GaugeOption) Gauge {
	return NewGaugeWithLabel[*NoLabel](prefix, name, opts...).Gauge(nil)
}

// DurationMetrics is a labeled Prometheus histogram vec for durations.
type DurationMetrics[T Label] struct {
	vec *prometheus.HistogramVec
}

func NewDurationWithLabel[T Label](prefix, name string, opts ...DurationOption) DurationRecorder[T] {
	o := &durationOptions{
		registerer: prometheus.DefaultRegisterer,
		buckets:    prometheus.DefBuckets,
	}
	for _, opt := range opts {
		opt(o)
	}
	var zero T
	vec := promauto.With(o.registerer).NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: prefix,
			Name:      name + "_duration_seconds",
			Help:      name + " duration in seconds",
			Buckets:   o.buckets,
		},
		zero.Labels(),
	)
	return &DurationMetrics[T]{vec: vec}
}

func (m *DurationMetrics[T]) Duration(label T) DurationObserver {
	return &durationObserver{m.vec.WithLabelValues(label.Values()...)}
}

// NewDuration returns a DurationObserver with no labels.
func NewDuration(prefix, name string, opts ...DurationOption) DurationObserver {
	return NewDurationWithLabel[*NoLabel](prefix, name, opts...).Duration(nil)
}

type counter struct{ c prometheus.Counter }

func (c *counter) Add(val float64) { c.c.Add(val) }
func (c *counter) Inc()            { c.c.Inc() }

type durationObserver struct{ h prometheus.Observer }

func (d *durationObserver) Observe(dur time.Duration) { d.h.Observe(dur.Seconds()) }

type gauge struct{ g prometheus.Gauge }

func (g *gauge) Set(val float64) { g.g.Set(val) }
func (g *gauge) Inc()            { g.g.Inc() }
func (g *gauge) Dec()            { g.g.Dec() }
func (g *gauge) Add(val float64) { g.g.Add(val) }
func (g *gauge) Sub(val float64) { g.g.Sub(val) }
