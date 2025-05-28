package metrics

import (
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestCounter            *prometheus.CounterVec
	RequestDuration           *prometheus.HistogramVec
	ErrorCounter              *prometheus.CounterVec
	ResponseSize              *prometheus.HistogramVec
	AuthServiceCallDuration   *prometheus.HistogramVec
	StaticServiceCallDuration *prometheus.HistogramVec
	AuthServiceCallCounter    *prometheus.CounterVec
	StaticServiceCallCounter  *prometheus.CounterVec
	LayerErrorCounter         *prometheus.CounterVec
)

func Init(namespace string) {
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "Duration of HTTP requests",
			Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	ErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_errors_total",
			Help:      "Total number of HTTP errors",
		},
		[]string{"method", "path", "status"},
	)

	ResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_response_size_bytes",
			Help:      "Size of HTTP responses",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 5),
		},
		[]string{"method", "path"},
	)

	AuthServiceCallDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "auth_service_call_duration_seconds",
			Help:      "Duration of auth service calls",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	StaticServiceCallDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "static_service_call_duration_seconds",
			Help:      "Duration of static service calls",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	AuthServiceCallCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "auth_service_call_total",
			Help:      "Total number of auth service calls",
		},
		[]string{"method", "status"},
	)

	StaticServiceCallCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "static_service_call_total",
			Help:      "Total number of static service calls",
		},
		[]string{"method", "status"},
	)

	LayerErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "layer_errors_total",
			Help:      "Total errors in layer (delivery/usecase/repository)",
		},
		[]string{"layer", "method"},
	)

	prometheus.MustRegister(
		RequestCounter,
		RequestDuration,
		ErrorCounter,
		ResponseSize,
		AuthServiceCallDuration,
		StaticServiceCallDuration,
		AuthServiceCallCounter,
		StaticServiceCallCounter,
		LayerErrorCounter,
	)

}

var (
	uuidRegex        = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	numericRegex     = regexp.MustCompile(`^\d+$`)
	specialCharRegex = regexp.MustCompile(`[~()/|!#;*^.@_\-=+%]`)
)

func NormalizePath(path string) string {
	segments := strings.Split(path, "/")

	for i, segment := range segments {
		switch {
		case uuidRegex.MatchString(segment):
			segments[i] = ":uuid"
		case numericRegex.MatchString(segment):
			segments[i] = ":id"
		case specialCharRegex.MatchString(segment):
			segments[i] = ":param"
		}
	}

	return strings.Join(segments, "/")
}
