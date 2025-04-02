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

//package main
//
//import (
//	"context"
//	"database/sql"
//	"fmt"
//	"log"
//	"time"
//
//	"github.com/gomodule/redigo/redis" // Redigo
//	_ "github.com/lib/pq"              // Драйвер PostgreSQL
//)
//
//func main() {
//	// Проверка PostgreSQL
//	checkPostgres()
//
//	// Проверка Redis
//	checkRedis()
//
//	fmt.Println("\nПроверка завершена!")
//}
//
//func checkPostgres() {
//	// Жёстко заданная строка подключения
//	dsn := "postgres://postgres:postgres@localhost:8070/resumatch?sslmode=disable"
//
//	db, err := sql.Open("postgres", dsn)
//	if err != nil {
//		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
//	}
//	defer db.Close()
//
//	// Проверяем соединение с таймаутом
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	err = db.PingContext(ctx)
//	if err != nil {
//		log.Fatalf("PostgreSQL недоступен: %v", err)
//	}
//
//	fmt.Println("✅ PostgreSQL: подключение успешно!")
//}
//
//func checkRedis() {
//	// Подключаемся к Redis (порт 8090, как у вас в docker-compose)
//	conn, err := redis.Dial("tcp", "localhost:8090")
//	if err != nil {
//		log.Fatalf("Redis недоступен: %v", err)
//	}
//	defer conn.Close()
//
//	// Проверяем соединение (PING -> PONG)
//	reply, err := conn.Do("PING")
//	if err != nil {
//		log.Fatalf("Ошибка PING: %v", err)
//	}
//
//	fmt.Printf("✅ Redis: подключение успешно! Ответ PING: %v\n", reply)
//}
