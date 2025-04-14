package app

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/repository/postgres"
	"ResuMatch/internal/repository/redis"
	"ResuMatch/internal/server"
	handler "ResuMatch/internal/transport/http"
	"ResuMatch/internal/usecase/service"
	"ResuMatch/pkg/connector"
	l "ResuMatch/pkg/logger"
	"net/http"
)

func Init(cfg *config.Config) *server.Server {
	// Postgres Connection

	vacancyConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Failed to connect to vacancy postgres: %v", err)
	}

	specializationConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Failed to connect to specialization postgres: %v", err)
	}

	cityConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Failed to connect to city postgres: %v", err)

	}

	staticConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Failed to connect to static postgres: %v", err)
	}

	applicantConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Failed to connect to applicant postgres: %v", err)
	}

	employerConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Failed to connect to employer postgres: %v", err)
	}

	// Redis Connection
	sessionConn, err := connector.NewRedisConnection(cfg.Redis)
	if err != nil {
		l.Log.Errorf("Failed to connect to session redis: %v", err)
	}

	// Repositories Init

	vacancyRepo, err := postgres.NewVacancyRepository(vacancyConn)
	if err != nil {
		l.Log.Errorf("Failed to create vacancy repository: %v", err)
	}

	specializationRepo, err := postgres.NewSpecializationRepository(specializationConn)
	if err != nil {
		l.Log.Errorf("Failed to create specialization repository: %v", err)
	}

	cityRepo, err := postgres.NewCityRepository(cityConn)
	if err != nil {
		l.Log.Errorf("Failed to create city repository: %v", err)
	}

	staticRepo, err := postgres.NewStaticRepository(staticConn)
	if err != nil {
		l.Log.Errorf("Failed to create static repository: %v", err)
	}

	applicantRepo, err := postgres.NewApplicantRepository(applicantConn)
	if err != nil {
		l.Log.Errorf("Failed to create applicant repository: %v", err)
	}

	employerRepo, err := postgres.NewEmployerRepository(employerConn)
	if err != nil {
		l.Log.Errorf("Failed to create employer repository: %v", err)
	}

	sessionRepo, err := redis.NewSessionRepository(sessionConn, cfg.Redis.TTL)
	if err != nil {
		l.Log.Errorf("Failed to create session repository: %v", err)
	}

	// Use Cases Init
<<<<<<< HEAD
=======
	authService := service.NewAuthService(sessionRepo, applicantRepo, employerRepo)
>>>>>>> 73e0bd5 (Fix app)
	staticService := service.NewStaticService(staticRepo)
	authService := service.NewAuthService(sessionRepo, applicantRepo, employerRepo)
	applicantService := service.NewApplicantService(applicantRepo, cityRepo, staticRepo)
	employerService := service.NewEmployerService(employerRepo, staticRepo)
	vacancyService := service.NewVacanciesService(vacancyRepo, cityRepo, applicantRepo, specializationRepo)

	authHandler := handler.NewAuthHandler(authService, cfg.CSRF)
	applicantHandler := handler.NewApplicantHandler(authService, applicantService, staticService, cfg.CSRF)
	employmentHandler := handler.NewEmployerHandler(authService, employerService, staticService, cfg.CSRF)
	vacancyHandler := handler.NewVacancyHandler(authService, vacancyService, cfg.CSRF)
	// Server Init
	srv := server.NewServer(cfg)

	// Router config
	srv.SetupRoutes(func(r *http.ServeMux) {
		authHandler.Configure(r)
		applicantHandler.Configure(r)
		employmentHandler.Configure(r)
		vacancyHandler.Configure(r)
	})

	return srv
}
