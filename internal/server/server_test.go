package server

import (
	"ResuMatch/internal/config"
	// "context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewServer(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок конфига
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Host:           "localhost",
			Port:           "8080",
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}

	server := NewServer(cfg)

	require.NotNil(t, server)
	require.Equal(t, "localhost:8080", server.httpServer.Addr)
	require.Equal(t, 10*time.Second, server.httpServer.ReadTimeout)
	require.Equal(t, 10*time.Second, server.httpServer.WriteTimeout)
	require.Equal(t, 1<<20, server.httpServer.MaxHeaderBytes)
}

func TestSetupRoutes(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок конфига
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			CORSAllowedOrigins: []string{"*"},
		},
		CSRF: config.CSRFConfig{
			CookieName: "csrf_token",
			Secret:     "secret",
		},
	}
	server := NewServer(cfg)

	// Мокируем функцию routeConfig
	called := false
	mockRouteConfig := func(mux *http.ServeMux) {
		called = true
		mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {})
	}

	server.SetupRoutes(mockRouteConfig)

	require.True(t, called, "routeConfig should be called")
	require.NotNil(t, server.httpServer.Handler)
}

func TestRunAndStop(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем тестовый сервер
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Host:           "localhost",
			Port:           "0", // 0 означает автоматический выбор порта
			ReadTimeout:    1 * time.Second,
			WriteTimeout:   1 * time.Second,
			MaxHeaderBytes: 1 << 10,
		},
	}
	server := NewServer(cfg)
	server.SetupRoutes(func(mux *http.ServeMux) {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	// Запускаем сервер в отдельной горутине
	go func() {
		err := server.Run()
		require.Error(t, err, http.ErrServerClosed)
	}()

	// Даем серверу время запуститься
	time.Sleep(100 * time.Millisecond)

	// Останавливаем сервер
	err := server.Stop()
	require.NoError(t, err)
}
