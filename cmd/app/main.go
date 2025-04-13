package main

import (
	"ResuMatch/internal/app"
	"ResuMatch/internal/config"
	l "ResuMatch/pkg/logger"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// @title ResuMatch API
// @version 1.0.0
// @description API веб-приложения ResuMatch для поиска работы и сотрудников.
// @BasePath  /api/v1
// @securityDefinitions.apikey csrf_token
// @in header
// @name X-CSRF-Token
// @securityDefinitions.apikey session_cookie
// @in cookie
// @name session_id
func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		l.Log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Инициализация приложения
	srv := app.Init(cfg)

	// 3. Настройка graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		l.Log.Info("Shutting down server...")
		if err := srv.Stop(); err != nil {
			l.Log.Fatalf("Failed to stop server: %v", err)
		}
	}()

	// 4. Запуск сервера
	l.Log.Infof("Starting server on %s", cfg.HTTP.Port)
	if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Log.Fatalf("Failed to run server: %v", err)
	}
}
