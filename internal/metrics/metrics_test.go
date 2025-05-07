package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Parallel()

	namespace := "test_namespace"
	Init(namespace)

	// Проверяем, что метрики инициализированы
	require.NotNil(t, RequestCounter)
	require.NotNil(t, RequestDuration)
	require.NotNil(t, ErrorCounter)
	require.NotNil(t, ResponseSize)
	require.NotNil(t, AuthServiceCallDuration)
	require.NotNil(t, StaticServiceCallDuration)
	require.NotNil(t, AuthServiceCallCounter)
	require.NotNil(t, StaticServiceCallCounter)
	require.NotNil(t, LayerErrorCounter)

	// Проверяем, что метрики зарегистрированы
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	registeredMetrics := map[string]bool{
		namespace + "_http_requests_total":                  false,
		namespace + "_http_request_duration_seconds":        false,
		namespace + "_http_errors_total":                    false,
		namespace + "_http_response_size_bytes":             false,
		namespace + "_auth_service_call_duration_seconds":   false,
		namespace + "_static_service_call_duration_seconds": false,
		namespace + "_auth_service_call_total":              false,
		namespace + "_static_service_call_total":            false,
		namespace + "_layer_errors_total":                   false,
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

func TestNormalizePath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected string
	}{
		{"/user/123/profile", "/user/:id/profile"},
		{"/order/550e8400-e29b-41d4-a716-446655440000/details", "/order/:uuid/details"},
		{"/search/query!@", "/search/:param"},
		{"/product/42/features", "/product/:id/features"},
		{"/api/v1/resource/xyz", "/api/v1/resource/xyz"}, // Без изменений
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			result := NormalizePath(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
