package main

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/metrics"
	"ResuMatch/internal/repository/redis"
	"ResuMatch/internal/transport/grpc/auth"
	authPROTO "ResuMatch/internal/transport/grpc/auth/proto"
	"ResuMatch/internal/transport/grpc/interceptors"
	"ResuMatch/internal/usecase/service"
	"ResuMatch/pkg/connector"
	l "ResuMatch/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.LoadAuthConfig()
	if err != nil {
		l.Log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}

	metrics.Init("resumatch")

	// Redis Connection
	sessionConn, err := connector.NewRedisConnection(cfg.Redis)
	if err != nil {
		l.Log.Errorf("Не удалось установить соедиение с Redis: %v", err)
	}

	// Redis repository
	sessionRepo, err := redis.NewSessionRepository(sessionConn, cfg.Redis.TTL)
	if err != nil {
		l.Log.Errorf("Не удалось создать репозиторий сессий: %v", err)
	}

	// Auth UC
	authService := service.NewAuthService(sessionRepo)

	// grpc
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.RequestIDServerInterceptor()),
	)

	authGRPC := auth.NewGRPC(authService)
	authPROTO.RegisterAuthServiceServer(grpcServer, authGRPC)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(cfg.MetricPort, nil); err != nil {
			l.Log.Fatalf("Ошибка запуска HTTP-сервера для метрик авторизации: %v", err)
		}
	}()

	listener, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		l.Log.Fatalf("Невозможно прослушать порт по адресу %s: %v", cfg.Addr(), err)
	}

	l.Log.Infof("Запуск сервиса авторизации на %s", cfg.Addr())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			l.Log.Fatal("ошибка сервера gRPC:", err)
		}
	}()

	sig := <-quit
	l.Log.Infof("Завершение работы сервиса авторизации по сигналу: %v", sig)

	grpcServer.GracefulStop()
	l.Log.Info("Сервис авторизации остановлен")

}
