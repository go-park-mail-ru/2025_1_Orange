package middleware

import (
	"bytes"
	// "context"
	"net/http"
	"net/http/httptest"

	// "strings"
	"testing"
	// "time"

	"ResuMatch/internal/utils"
	"ResuMatch/pkg/logger"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestCustomResponseWriter(t *testing.T) {
	t.Parallel()

	t.Run("WriteHeader sets status code once", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		cw := &customResponseWriter{ResponseWriter: rec}

		cw.WriteHeader(http.StatusOK)
		require.Equal(t, http.StatusOK, cw.statusCode)

		// Повторный вызов не должен менять статус
		cw.WriteHeader(http.StatusInternalServerError)
		require.Equal(t, http.StatusOK, cw.statusCode)
	})

	t.Run("Write sets default status code", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		cw := &customResponseWriter{ResponseWriter: rec}

		_, err := cw.Write([]byte("test"))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, cw.statusCode) // Write должен установить статус 200
	})
}

func TestAccessLogMiddleware(t *testing.T) {
	t.Parallel()

	// Сохраняем оригинальный логгер
	originalLogger := logger.Log
	defer func() { logger.Log = originalLogger }()

	t.Run("Logs request details", func(t *testing.T) {
		t.Parallel()

		// Настраиваем тестовый логгер
		var logOutput bytes.Buffer
		testLogger := logrus.New()
		testLogger.SetOutput(&logOutput)
		testLogger.SetFormatter(&logger.CoolFormatter{
			NoColors:     true,
			TrimMessages: true,
		})
		logger.Log = testLogger

		// Создаем тестовый запрос
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.Header.Set("X-Forwarded-For", "1.2.3.4")

		// Добавляем requestID в контекст через utils
		ctx := utils.SetRequestID(req.Context(), "test-request-id")
		req = req.WithContext(ctx)

		// Создаем тестовый ResponseWriter
		rec := httptest.NewRecorder()

		// Создаем middleware и тестовый обработчик
		handler := AccessLogMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Вызываем middleware
		handler.ServeHTTP(rec, req)

		// Проверяем вывод логов
		logStr := logOutput.String()
		require.Contains(t, logStr, "method=GET")
		require.Contains(t, logStr, "path=/test")
		require.Contains(t, logStr, "status=200")
		require.Contains(t, logStr, "ip=1.2.3.4")
		require.Contains(t, logStr, "ua=test-agent")
		require.Contains(t, logStr, "requestID=test-request-id")
		require.Contains(t, logStr, "latency=")
	})

	t.Run("Handles missing headers", func(t *testing.T) {
		t.Parallel()

		var logOutput bytes.Buffer
		testLogger := logrus.New()
		testLogger.SetOutput(&logOutput)
		testLogger.SetFormatter(&logger.CoolFormatter{
			NoColors:     true,
			TrimMessages: true,
		})
		logger.Log = testLogger

		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		handler := AccessLogMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		handler.ServeHTTP(rec, req)

		logStr := logOutput.String()
		require.Contains(t, logStr, "ua=")
		require.Contains(t, logStr, "status=404")
	})
}

func TestGetClientIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		headers        map[string]string
		remoteAddr     string
		expectedResult string
	}{
		{
			name:           "X-Forwarded-For",
			headers:        map[string]string{"X-Forwarded-For": "1.2.3.4"},
			expectedResult: "1.2.3.4",
		},
		{
			name:           "X-Real-Ip",
			headers:        map[string]string{"X-Real-Ip": "5.6.7.8"},
			expectedResult: "5.6.7.8",
		},
		{
			name:           "RemoteAddr with port",
			remoteAddr:     "9.10.11.12:1234",
			expectedResult: "9.10.11.12",
		},
		{
			name:           "RemoteAddr without port",
			remoteAddr:     "9.10.11.12",
			expectedResult: "9.10.11.12",
		},
		{
			name:           "Prefer X-Forwarded-For over others",
			headers:        map[string]string{"X-Forwarded-For": "1.2.3.4", "X-Real-Ip": "5.6.7.8"},
			remoteAddr:     "9.10.11.12:1234",
			expectedResult: "1.2.3.4",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			result := getClientIP(req)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}
