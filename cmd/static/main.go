package main

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/metrics"
	"ResuMatch/internal/repository/postgres"
	"ResuMatch/internal/transport/grpc/interceptors"
	"ResuMatch/internal/transport/grpc/static"
	staticPROTO "ResuMatch/internal/transport/grpc/static/proto"
	"ResuMatch/internal/usecase/service"
	"ResuMatch/pkg/connector"
	l "ResuMatch/pkg/logger"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.LoadS3Config()
	if err != nil {
		l.Log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}

	metrics.Init("resumatch")

	staticConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("не удалось установить соединение соединение с static postgres: %v", err)
	}

	staticRepo, err := postgres.NewStaticRepository(staticConn, cfg.Minio.Bucket, cfg.Minio)
	if err != nil {
		l.Log.Errorf("ошибка создания репозитория статики: %v", err)
	}

	staticService := service.NewStaticService(staticRepo)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.RequestIDServerInterceptor()),
	)

	staticGRPC := static.NewGRPC(staticService)
	staticPROTO.RegisterStaticServiceServer(grpcServer, staticGRPC)

	listener, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		l.Log.Fatalf("Невозможно прослушать порт по адресу %s: %v", cfg.Addr(), err)
	}

	l.Log.Infof("Запуск сервиса статики на %s", cfg.Addr())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			l.Log.Fatal("ошибка сервера gRPC:", err)
		}
	}()

	sig := <-quit
	l.Log.Infof("Завершение работы сервиса статики по сигналу: %v", sig)

	grpcServer.GracefulStop()
	l.Log.Info("Сервис статики остановлен")
}
