package main

import (
	"ResuMatch/internal/app"
	"ResuMatch/internal/config"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Инициализация приложения
	srv := app.Init(cfg)

	// 3. Настройка graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
		if err := srv.Stop(); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
	}()

	// 4. Запуск сервера
	log.Printf("Starting server on %s", cfg.HTTP.Port)
	if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed: %v", err)
	}
}
