package middleware

import (
	"net/http"
	"strconv"
	"time"

	"ResuMatch/internal/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsMiddleware(metrics *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rec, r)

			duration := time.Since(start).Seconds()
			method := r.Method
			path := r.URL.Path
			status := strconv.Itoa(rec.statusCode)

			metrics.RequestCounter.WithLabelValues(method, path, status).Inc()
			metrics.RequestDuration.WithLabelValues(method, path).Observe(duration)
			metrics.ResponseSize.WithLabelValues(method, path).Observe(float64(rec.size))

			if rec.statusCode >= 400 {
				metrics.ErrorCounter.WithLabelValues(method, path, status).Inc()
			}
		})
	}
}
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.statusCode = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}
