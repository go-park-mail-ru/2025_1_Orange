package main

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/repository/postgres"
	"ResuMatch/internal/transport/grpc/poll"
	pollPROTO "ResuMatch/internal/transport/grpc/poll/proto"
	"ResuMatch/internal/usecase/service"
	"ResuMatch/pkg/connector"
	l "ResuMatch/pkg/logger"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.LoadPollConfig()
	if err != nil {
		l.Log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}

	pollConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Failed to connect to poll postgres: %v", err)
	}

	pollRepo, err := postgres.NewPollRepository(pollConn)
	if err != nil {
		l.Log.Errorf("Failed to create poll repository: %v", err)
	}

	// Poll UC
	pollService := service.NewPollService(pollRepo)

	// grpc
	grpcServer := grpc.NewServer()

	pollGRPC := poll.NewGRPC(pollService)
	pollPROTO.RegisterPollServiceServer(grpcServer, pollGRPC)

	listener, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		l.Log.Fatalf("Невозможно прослушать порт по адресу %s: %v", cfg.Addr(), err)
	}

	l.Log.Infof("Запуск сервиса отзывов на %s", cfg.Addr())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			l.Log.Fatal("ошибка сервера gRPC:", err)
		}
	}()

	sig := <-quit
	l.Log.Infof("Завершение работы сервиса отзывов по сигналу: %v", sig)

	grpcServer.GracefulStop()
	l.Log.Info("Сервис отзывов остановлен")
}
