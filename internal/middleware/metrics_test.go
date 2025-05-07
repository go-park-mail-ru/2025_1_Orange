package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"ResuMatch/internal/metrics"

	"github.com/stretchr/testify/require"
)

func TestMetricsMiddleware(t *testing.T) {
	t.Parallel()

	// Инициализация метрик
	metrics.Init("test_namespace")

	t.Run("Successful request", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test/path", nil)

		handler := MetricsMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("Hello, world!"))
			require.NoError(t, err)
		}))

		//start := time.Now()
		handler.ServeHTTP(w, r)
		//duration := time.Since(start).Seconds()

		require.Equal(t, http.StatusOK, w.Code)

		normalizedPath := metrics.NormalizePath("/test/path")
		status := strconv.Itoa(http.StatusOK)

		require.NotNil(t, metrics.RequestCounter.WithLabelValues("GET", normalizedPath, status))
		require.NotNil(t, metrics.RequestDuration.WithLabelValues("GET", normalizedPath))
		require.NotNil(t, metrics.ResponseSize.WithLabelValues("GET", normalizedPath))
	})

	t.Run("Error response", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/error/path", nil)

		handler := MetricsMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "error", http.StatusInternalServerError)
		}))

		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)

		normalizedPath := metrics.NormalizePath("/error/path")
		status := strconv.Itoa(http.StatusInternalServerError)

		require.NotNil(t, metrics.ErrorCounter.WithLabelValues("POST", normalizedPath, status))
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
