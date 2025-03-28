package app

import (
	"ResuMatch/internal/repository/mysql"
	ses "ResuMatch/internal/repository/redis"

	"ResuMatch/internal/usecase"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

// InitializeDatabase - подключает MySQL и возвращает объект базы данных.
func InitializeDatabase() (*sql.DB, error) {
	dsn := configs.MySQLConfig.GetConnectionString()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// InitializeRedis - подключает Redis и возвращает пул соединений.
func InitializeRedis() (*redis.Pool, error) {
	pool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", configs.RedisConfig.Address)
		},
	}
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// InitializeRouter - создаёт роутер с зарегистрированными хендлерами.
func InitializeRouter(db *sql.DB, redisPool *redis.Pool) *mux.Router {
	// Инициализация репозиториев
	userRepo := mysql.NewUserRepository(db)
	sessionRepo := ses.NewSessionRedisRepository(redisPool)

	// Инициализация юзкейсов
	authUsecase := &usecase.AuthUsecase{
		userRepository:    userRepo,
		SessionRepository: sessionRepo,
	}
	// Создание роутера
	router := mux.NewRouter()
	// Регистрация хендлеров
	http.NewUserHandler(router, *authUsecase)
	return nil
}

// Run - основная функция запуска приложения.
func Run() error {
	// Настройка логирования
	db, err := InitializeDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	// Инициализация Redis
	redisPool, err := InitializeRedis()
	if err != nil {
		return err
	}
	defer redisPool.Close()

	// Создание и настройка роутера
	router := InitializeRouter(db, redisPool)
	PORT := ":8000"
	// Запуск HTTP-сервера
	fmt.Printf("\tstarting server at %s\n", PORT)
	log.Println("starting server at", PORT)

	err = http.ListenAndServe(PORT, router)
	if err != nil {
		return err
	}

	return nil
}
