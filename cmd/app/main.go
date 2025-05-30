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
	// 1. создание vault client
	// 2. Загрузка конфигурации
	cfg, err := config.LoadAppConfig()
	if err != nil {
		l.Log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}

	// 3. Инициализация приложения
	srv := app.Init(cfg)

	// 4. Настройка graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-quit
		l.Log.Infof("Завершение работы сервера приложения... : %v", sig)
		if err := srv.Stop(); err != nil {
			l.Log.Fatalf("Не удалось остановить сервер приложения: %v", err)
		}
	}()

	// 5. Запуск сервера
	l.Log.Infof("Запуск сервера на порте %s", cfg.HTTP.Port)
	if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
