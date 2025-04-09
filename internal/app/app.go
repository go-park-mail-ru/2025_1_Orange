package app

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/repository/postgres"
	"ResuMatch/internal/repository/redis"
	"ResuMatch/internal/server"
	handler "ResuMatch/internal/transport/http"
	"ResuMatch/internal/usecase/service"
	"log"
	"net/http"
)

func Init(cfg *config.Config) *server.Server {
	// Repositories Init
	cityRepo, err := postgres.NewCityRepository(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to create city repository: %v", err)
	}

	staticRepo, err := postgres.NewStaticRepository(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to create static repository: %v", err)
	}

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
	staticService := service.NewStaticService(staticRepo)
	authService := service.NewAuthService(sessionRepo, applicantRepo, employerRepo)
	applicantService := service.NewApplicantService(applicantRepo, cityRepo, staticRepo)
	employerService := service.NewEmployerService(employerRepo, staticRepo)
	// Transport Init
	authHandler := handler.NewAuthHandler(authService)
	applicantHandler := handler.NewApplicantHandler(authService, applicantService, staticService, cfg.CSRF)
	employmentHandler := handler.NewEmployerHandler(authService, employerService, staticService, cfg.CSRF)

	// Server Init
	srv := server.NewServer(cfg)

	// Router config
	srv.SetupRoutes(func(r *http.ServeMux) {
		authHandler.Configure(r)
		applicantHandler.Configure(r)
		employmentHandler.Configure(r)
	})

	return srv
}
