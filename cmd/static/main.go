package main

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/repository/postgres"
	"ResuMatch/internal/usecase/service"
	"ResuMatch/pkg/connector"
	l "ResuMatch/pkg/logger"
)

func main() {
	cfg, err := config.LoadS3Config()
	if err != nil {
		l.Log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}

	staticConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с static postgres: %v", err)
	}

	applicantStaticRepo, err := postgres.NewStaticRepository(staticConn, cfg.Minio.Buckets.ApplicantBucket, cfg.Minio)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория статики для соискателя: %v", err)
	}

	employerStaticRepo, err := postgres.NewStaticRepository(staticConn, cfg.Minio.Buckets.EmployerBucket, cfg.Minio)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория статики для работодателя: %v", err)
	}

	applicantStaticService := service.NewStaticService(applicantStaticRepo)
	employerStaticService := service.NewStaticService(employerStaticRepo)
}
