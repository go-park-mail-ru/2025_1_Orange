package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	RequestCounter  *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	ErrorCounter    *prometheus.CounterVec
	ResponseSize    *prometheus.HistogramVec
}

func NewMetrics(namespace string) *Metrics {
	m := &Metrics{
		RequestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "Duration of HTTP requests",
				Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),
		ErrorCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_errors_total",
				Help:      "Total number of HTTP errors",
			},
			[]string{"method", "path", "status"},
		),
		ResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "Size of HTTP responses",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 5),
			},
			[]string{"method", "path"},
		),
	}

	prometheus.MustRegister(
		m.RequestCounter,
		m.RequestDuration,
		m.ErrorCounter,
		m.ResponseSize,
	)

	return m
}
