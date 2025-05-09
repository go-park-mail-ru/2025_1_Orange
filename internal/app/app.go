package app

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/metrics"
	"ResuMatch/internal/repository/postgres"
	"ResuMatch/internal/server"
	"ResuMatch/internal/transport/grpc/auth"
	"ResuMatch/internal/transport/grpc/static"
	handler "ResuMatch/internal/transport/http"
	"ResuMatch/internal/usecase/service"
	"ResuMatch/pkg/connector"
	l "ResuMatch/pkg/logger"
	"net/http"
)

func Init(cfg *config.Config) *server.Server {
	// Postgres Connection
	resumeConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с resume postgres: %v", err)
	}

	vacancyConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с vacancy postgres: %v", err)
	}

	skillConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с skill postgres: %v", err)
	}

	specializationConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с specialization postgres: %v", err)
	}

	cityConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с city postgres: %v", err)
	}

	applicantConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с applicant postgres: %v", err)
	}

	employerConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с employer postgres: %v", err)
	}

	// Repositories Init
	resumeRepo, err := postgres.NewResumeRepository(resumeConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория резюме: %v", err)
	}

	vacancyRepo, err := postgres.NewVacancyRepository(vacancyConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория вакансии: %v", err)
	}

	skillRepo, err := postgres.NewSkillRepository(skillConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория навыка: %v", err)
	}

	specializationRepo, err := postgres.NewSpecializationRepository(specializationConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория специализации: %v", err)
	}

	cityRepo, err := postgres.NewCityRepository(cityConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория города: %v", err)
	}

	applicantRepo, err := postgres.NewApplicantRepository(applicantConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория соискателя: %v", err)
	}

	employerRepo, err := postgres.NewEmployerRepository(employerConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория работодателя: %v", err)
	}

	// Use Cases Init
	staticService, err := static.NewGateway(cfg.Microservices.S3.Addr())
	if err != nil {
		l.Log.Errorf("Ошибка при подключении к сервису статики: %v", err)
	}

	authService, err := auth.NewGateway(cfg.Microservices.Auth.Addr())
	if err != nil {
		l.Log.Errorf("Ошибка при подключении к сервису авторизации: %v", err)
	}

	applicantService := service.NewApplicantService(applicantRepo, cityRepo, staticService)
	employerService := service.NewEmployerService(employerRepo, staticService)

	specializationService := service.NewSpecializationService(specializationRepo)

	resumeService := service.NewResumeService(resumeRepo, skillRepo, specializationRepo, applicantRepo, applicantService)
	vacancyService := service.NewVacanciesService(vacancyRepo, applicantRepo, specializationRepo, employerService)

	// Transport Init
	authHandler := handler.NewAuthHandler(authService, cfg.CSRF)
	applicantHandler := handler.NewApplicantHandler(authService, applicantService, cfg.CSRF)
	employmentHandler := handler.NewEmployerHandler(authService, employerService, cfg.CSRF)
	resumeHandler := handler.NewResumeHandler(authService, resumeService, cfg.CSRF)
	vacancyHandler := handler.NewVacancyHandler(authService, vacancyService, cfg.CSRF)
	specializationHandler := handler.NewSpecializationHandler(specializationService)

	// Metrics Init
	metrics.Init("resumatch")

	// Server Init
	srv := server.NewServer(cfg)

	// Router config
	srv.SetupRoutes(func(r *http.ServeMux) {
		authHandler.Configure(r)
		applicantHandler.Configure(r)
		employmentHandler.Configure(r)
		resumeHandler.Configure(r)
		vacancyHandler.Configure(r)
		specializationHandler.Configure(r)
	})

	return srv
}
