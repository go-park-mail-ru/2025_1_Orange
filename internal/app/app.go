package app

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/repository/postgres"
	"ResuMatch/internal/server"
	"ResuMatch/internal/transport/grpc/auth"
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
<<<<<<< HEAD
=======
	if err != nil {
		l.Log.Errorf("Failed to connect to vacancy postgres: %v", err)
	}

	skillConn, err := connector.NewPostgresConnection(cfg.Postgres)
>>>>>>> 442b3e9 (Add vacancy routing and fix issues with get vacancies)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с vacancy postgres: %v", err)
	}

<<<<<<< HEAD
	skillConn, err := connector.NewPostgresConnection(cfg.Postgres)
=======
	specializationConn, err := connector.NewPostgresConnection(cfg.Postgres)
>>>>>>> 442b3e9 (Add vacancy routing and fix issues with get vacancies)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с skill postgres: %v", err)
	}

<<<<<<< HEAD
	specializationConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с specialization postgres: %v", err)
	}

=======
>>>>>>> 442b3e9 (Add vacancy routing and fix issues with get vacancies)
	cityConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с city postgres: %v", err)
	}

	staticConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с static postgres: %v", err)
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

	vacancyRepo, err := postgres.NewVacancyRepository(vacancyConn)
	if err != nil {
		l.Log.Errorf("Failed to create vacancy repository: %v", err)
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

	applicantStaticRepo, err := postgres.NewStaticRepository(staticConn, cfg.Minio.Buckets.ApplicantBucket, cfg.Minio.Config)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория статики для соискателя: %v", err)
	}

	employerStaticRepo, err := postgres.NewStaticRepository(staticConn, cfg.Minio.Buckets.EmployerBucket, cfg.Minio.Config)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория статики для работодателя: %v", err)
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
<<<<<<< HEAD
	applicantStaticService := service.NewStaticService(applicantStaticRepo)
	employerStaticService := service.NewStaticService(employerStaticRepo)

	authService, err := auth.NewGateway(cfg.Microservices.Auth.Addr())
	if err != nil {
		l.Log.Errorf("Ошибка при подключении к сервису авторизации: %v", err)
	}

	applicantService := service.NewApplicantService(applicantRepo, cityRepo, applicantStaticRepo)
	employerService := service.NewEmployerService(employerRepo, employerStaticRepo)

	// resumeService := service.NewResumeService(resumeRepo, skillRepo, specializationRepo)
	resumeService := service.NewResumeService(resumeRepo, skillRepo, specializationRepo, applicantRepo, applicantService)
	vacancyService := service.NewVacanciesService(vacancyRepo, applicantRepo, specializationRepo, employerService)
=======
	staticService := service.NewStaticService(staticRepo)
	authService := service.NewAuthService(sessionRepo, applicantRepo, employerRepo)
	applicantService := service.NewApplicantService(applicantRepo, cityRepo, staticRepo)
	employerService := service.NewEmployerService(employerRepo, staticRepo)

	// resumeService := service.NewResumeService(resumeRepo, skillRepo, specializationRepo)
	resumeService := service.NewResumeService(resumeRepo, skillRepo, specializationRepo, applicantRepo, applicantService)
	vacancyService := service.NewVacanciesService(vacancyRepo, applicantRepo, specializationRepo)
>>>>>>> b8e1fb1 (Fix vacancy)

	// Transport Init
	authHandler := handler.NewAuthHandler(authService, cfg.CSRF)
	applicantHandler := handler.NewApplicantHandler(authService, applicantService, applicantStaticService, cfg.CSRF)
	employmentHandler := handler.NewEmployerHandler(authService, employerService, employerStaticService, cfg.CSRF)
	resumeHandler := handler.NewResumeHandler(authService, resumeService, cfg.CSRF)
	vacancyHandler := handler.NewVacancyHandler(authService, vacancyService, cfg.CSRF)

	// Server Init
	srv := server.NewServer(cfg)

	// Router config
	srv.SetupRoutes(func(r *http.ServeMux) {
		authHandler.Configure(r)
		applicantHandler.Configure(r)
		employmentHandler.Configure(r)
		resumeHandler.Configure(r)
		vacancyHandler.Configure(r)
	})

	return srv
}
