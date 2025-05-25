package middleware

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
	"time"

	"ResuMatch/internal/metrics"
)

func MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &hijackableResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rec, r)

			duration := time.Since(start).Seconds()
			method := r.Method
			path := metrics.NormalizePath(r.URL.Path)
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
