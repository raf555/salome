package prommetric

import "github.com/prometheus/client_golang/prometheus"

type options struct {
	buckets    []float64
	registerer prometheus.Registerer
}

type Option func(*options)

func WithBuckets(buckets []float64) Option {
	return func(o *options) {
		o.buckets = buckets
	}
}

func WithRegisterer(r prometheus.Registerer) Option {
	return func(o *options) {
		o.registerer = r
	}
}
