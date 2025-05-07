package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"ResuMatch/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestMetricsMiddleware(t *testing.T) {
	t.Parallel()

	metrics := &metrics.Metrics{
		RequestCounter:  prometheus.NewCounterVec(prometheus.CounterOpts{Name: "requests_total"}, []string{"method", "path", "status"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "request_duration_seconds"}, []string{"method", "path"}),
		ResponseSize:    prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "response_size_bytes"}, []string{"method", "path"}),
		ErrorCounter:    prometheus.NewCounterVec(prometheus.CounterOpts{Name: "errors_total"}, []string{"method", "path", "status"}),
	}

	prometheus.MustRegister(metrics.RequestCounter, metrics.RequestDuration, metrics.ResponseSize, metrics.ErrorCounter)

	t.Run("Successful request", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		handler := MetricsMiddleware(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, world!"))
		}))

		//start := time.Now()
		handler.ServeHTTP(w, r)
		//duration := time.Since(start).Seconds()

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "Hello, world!", w.Body.String())

		require.NotNil(t, metrics.RequestCounter.WithLabelValues("GET", "/test", strconv.Itoa(http.StatusOK)))
		require.NotNil(t, metrics.RequestDuration.WithLabelValues("GET", "/test"))
		require.NotNil(t, metrics.ResponseSize.WithLabelValues("GET", "/test"))
	})

	t.Run("Error response", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/fail", nil)

		handler := MetricsMiddleware(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "error", http.StatusInternalServerError)
		}))

		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.NotNil(t, metrics.ErrorCounter.WithLabelValues("POST", "/fail", strconv.Itoa(http.StatusInternalServerError)))
	})
}

func TestResponseRecorder(t *testing.T) {
	t.Parallel()

	t.Run("WriteHeader sets status code", func(t *testing.T) {
		t.Parallel()
		rec := httptest.NewRecorder()
		writer := &responseRecorder{ResponseWriter: rec}

		writer.WriteHeader(http.StatusForbidden)
		require.Equal(t, http.StatusForbidden, writer.statusCode)
	})

	t.Run("Write updates size", func(t *testing.T) {
		t.Parallel()
		rec := httptest.NewRecorder()
		writer := &responseRecorder{ResponseWriter: rec}

		data := []byte("Test response")
		size, err := writer.Write(data)
		require.NoError(t, err)
		require.Equal(t, len(data), size)
		require.Equal(t, len(data), writer.size)
	})
}
