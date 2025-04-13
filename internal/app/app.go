package app

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/repository/postgres"
	"ResuMatch/internal/repository/redis"
	"ResuMatch/internal/server"
	handler "ResuMatch/internal/transport/http"
	"ResuMatch/internal/usecase/service"
	"ResuMatch/pkg/connector"
	"log"
	"net/http"
)

func Init(cfg *config.Config) *server.Server {
	// Замечание 10 - Добавление коннектора
	// Создаем общее соединение с PostgreSQL для репозиториев резюме
	pgConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to create PostgreSQL connection: %v", err)
	}

	// Repositories Init
	// Замечание 10 - Добавление коннектора
	// Используем общее соединение для репозиториев резюме
	resumeRepo, err := postgres.NewResumeRepository(pgConn)
	if err != nil {
		log.Fatalf("Failed to create resume repository: %v", err)
	}

	skillRepo, err := postgres.NewSkillRepository(pgConn)
	if err != nil {
		log.Fatalf("Failed to create skill repository: %v", err)
	}

	specializationRepo, err := postgres.NewSpecializationRepository(pgConn)
	if err != nil {
		log.Fatalf("Failed to create specialization repository: %v", err)
	}

	// Оставляем старые репозитории без изменений
	applicantRepo, err := postgres.NewApplicantRepository(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to create applicant repository: %v", err)
	}

	employerRepo, err := postgres.NewEmployerRepository(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to create employer repository: %v", err)
	}

	sessionRepo, err := redis.NewSessionRepository(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to create session repository: %v", err)
	}

	// Use Cases Init
	authService := service.NewAuthService(sessionRepo, applicantRepo, employerRepo)
	applicantService := service.NewApplicantService(applicantRepo)
	employerService := service.NewEmployerService(employerRepo)
	resumeService := service.NewResumeService(resumeRepo, skillRepo, specializationRepo)

	// Transport Init
	authHandler := handler.NewAuthHandler(authService)
	applicantHandler := handler.NewApplicantHandler(authService, applicantService, cfg.CSRF)
	employmentHandler := handler.NewEmployerHandler(authService, employerService, cfg.CSRF)
	resumeHandler := handler.NewResumeHandler(authService, resumeService, cfg.CSRF)

	// Server Init
	srv := server.NewServer(cfg)

	// Router config
	srv.SetupRoutes(func(r *http.ServeMux) {
		authHandler.Configure(r)
		applicantHandler.Configure(r)
		employmentHandler.Configure(r)
		resumeHandler.Configure(r)
	})

	return srv
}
