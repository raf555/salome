package prommetric

import "github.com/prometheus/client_golang/prometheus"

type counterOptions struct {
	registerer prometheus.Registerer
}

type CounterOption func(*counterOptions)

func WithCounterRegisterer(r prometheus.Registerer) CounterOption {
	return func(o *counterOptions) {
		o.registerer = r
	}
}

type gaugeOptions struct {
	registerer prometheus.Registerer
}

type GaugeOption func(*gaugeOptions)

func WithGaugeRegisterer(r prometheus.Registerer) GaugeOption {
	return func(o *gaugeOptions) {
		o.registerer = r
	}
}

type durationOptions struct {
	registerer prometheus.Registerer
	buckets    []float64
}

type DurationOption func(*durationOptions)

func WithDurationRegisterer(r prometheus.Registerer) DurationOption {
	return func(o *durationOptions) {
		o.registerer = r
	}
}

func WithBuckets(buckets []float64) DurationOption {
	return func(o *durationOptions) {
		o.buckets = buckets
	}
}
