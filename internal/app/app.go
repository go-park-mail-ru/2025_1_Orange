package app

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/repository/postgres"
	"ResuMatch/internal/repository/redis"
	"ResuMatch/internal/server"
	"ResuMatch/internal/transport/http"
	"ResuMatch/internal/usecase/service"
	"github.com/julienschmidt/httprouter"
	"log"
)

func Init(cfg *config.Config) *server.Server {
	// Repositories Init
	applicantRepo, err := postgres.NewApplicantRepository(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to create applicant repository: %v", err)
	}

	employerRepo, err := postgres.NewEmployerDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to create employer repository: %v", err)
	}

	sessionRepo, err := redis.NewSessionRepository(cfg.Redis, cfg.Redis.TTL)
	if err != nil {
		log.Fatalf("Failed to create session repository: %v", err)
	}

	// Use Cases Init
	authService := service.NewAuthService(sessionRepo)
	applicantService := service.NewApplicantService(applicantRepo)
	employerService := service.NewEmployerService(employerRepo)

	// Transport Init
	authHandler := http.NewAuthHandler(authService)
	applicantHandler := http.NewApplicantHandler(authService, applicantService, cfg.CSRF)
	employmentHandler := http.NewEmployerHandler(authService, employerService, cfg.CSRF)

	// Server Init
	srv := server.NewServer(cfg)

	// Router config
	srv.SetupRoutes(func(r *httprouter.Router) {
		authHandler.Configure(r)
		applicantHandler.Configure(r)
		employmentHandler.Configure(r)
	})

	return srv
}
