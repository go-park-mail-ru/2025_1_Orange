package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewMetrics(t *testing.T) {
	t.Parallel()

	namespace := "test_namespace"
	metrics := NewMetrics(namespace)

	require.NotNil(t, metrics)
	require.NotNil(t, metrics.RequestCounter)
	require.NotNil(t, metrics.RequestDuration)
	require.NotNil(t, metrics.ErrorCounter)
	require.NotNil(t, metrics.ResponseSize)
}

func TestMetrics_Registration(t *testing.T) {
	t.Parallel()

	namespace := "test_namespace"
	//metrics := NewMetrics(namespace)

	// Проверяем, что метрики зарегистрированы
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	registeredMetrics := map[string]bool{
		namespace + "_http_requests_total":           false,
		namespace + "_http_request_duration_seconds": false,
		namespace + "_http_errors_total":             false,
		namespace + "_http_response_size_bytes":      false,
	}

	for _, mf := range metricFamilies {
		if _, exists := registeredMetrics[*mf.Name]; exists {
			registeredMetrics[*mf.Name] = true
		}
	}

	for name, registered := range registeredMetrics {
		require.True(t, registered, "Метрика %s не зарегистрирована", name)
	}
}
