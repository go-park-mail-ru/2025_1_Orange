package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCORS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		allowedOrigins      []string
		requestOrigin       string
		expectedAllowOrigin string
		method              string
		expectedStatus      int
	}{
		{
			name:                "Allowed origin",
			allowedOrigins:      []string{"http://example.com"},
			requestOrigin:       "http://example.com",
			expectedAllowOrigin: "http://example.com",
			method:              http.MethodGet,
			expectedStatus:      http.StatusOK,
		},
		{
			name:                "Wildcard origin",
			allowedOrigins:      []string{"*"},
			requestOrigin:       "http://any-origin.com",
			expectedAllowOrigin: "http://any-origin.com",
			method:              http.MethodGet,
			expectedStatus:      http.StatusOK,
		},
		{
			name:                "Not allowed origin",
			allowedOrigins:      []string{"http://example.com"},
			requestOrigin:       "http://other.com",
			expectedAllowOrigin: "",
			method:              http.MethodGet,
			expectedStatus:      http.StatusOK,
		},
		{
			name:                "OPTIONS request",
			allowedOrigins:      []string{"http://example.com"},
			requestOrigin:       "http://example.com",
			expectedAllowOrigin: "http://example.com",
			method:              http.MethodOptions,
			expectedStatus:      http.StatusNoContent,
		},
		{
			name:                "No origin header",
			allowedOrigins:      []string{"http://example.com"},
			requestOrigin:       "",
			expectedAllowOrigin: "",
			method:              http.MethodGet,
			expectedStatus:      http.StatusOK,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Создаем тестовый обработчик
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Создаем middleware с тестовыми параметрами
			handler := CORS(tc.allowedOrigins)(nextHandler)

			// Создаем тестовый запрос
			req := httptest.NewRequest(tc.method, "/", nil)
			if tc.requestOrigin != "" {
				req.Header.Set("Origin", tc.requestOrigin)
			}

			// Создаем recorder для записи ответа
			w := httptest.NewRecorder()

			// Вызываем middleware
			handler.ServeHTTP(w, req)

			// Проверяем результаты
			resp := w.Result()
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
			require.Equal(t, tc.expectedAllowOrigin, allowOrigin)

			// Проверяем стандартные CORS заголовки
			if tc.expectedAllowOrigin != "" {
				require.Equal(t, "GET,POST,PUT,DELETE,OPTIONS",
					resp.Header.Get("Access-Control-Allow-Methods"))
				require.Equal(t, "Content-Type,Authorization,X-CSRF-Token",
					resp.Header.Get("Access-Control-Allow-Headers"))
				require.Equal(t, "X-CSRF-Token",
					resp.Header.Get("Access-Control-Expose-Headers"))
				require.Equal(t, "true",
					resp.Header.Get("Access-Control-Allow-Credentials"))
			}
		})
	}
}
