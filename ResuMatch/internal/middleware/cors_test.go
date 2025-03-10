package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Тестируем, что CORS-заголовки устанавливаются
func TestCORSHeaders(t *testing.T) {
	// Создаем моковый обработчик, который просто возвращает 200 OK
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Оборачиваем его в middleware.CORS
	handler := CORS(mockHandler)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем, что CORS-заголовки установлены
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":      "http://localhost:5173 http://localhost:8001",
		"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers":     "Content-Type, Authorization",
		"Access-Control-Allow-Credentials": "true",
	}

	for key, expected := range expectedHeaders {
		if value := rr.Header().Get(key); value != expected {
			t.Errorf("Ожидался заголовок %s: %s, но получили %s", key, expected, value)
		}
	}
}

// Тестируем обработку preflight-запроса (OPTIONS)
func TestCORSPreflight(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS(mockHandler)

	req, _ := http.NewRequest("OPTIONS", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Должен вернуть 200 OK без передачи запроса дальше
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Preflight-запрос (OPTIONS) должен возвращать 200, но получил %d", status)
	}
}
