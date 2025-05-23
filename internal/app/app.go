package app

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/metrics"
	"ResuMatch/internal/repository/postgres"
	"ResuMatch/internal/server"
	"ResuMatch/internal/transport/grpc/auth"
	"ResuMatch/internal/transport/grpc/static"
	handler "ResuMatch/internal/transport/http"
	"ResuMatch/internal/transport/ws"
	"ResuMatch/internal/usecase/service"
	"ResuMatch/pkg/connector"
	l "ResuMatch/pkg/logger"
	"net/http"
)

func Init(cfg *config.Config) *server.Server {
	// Postgres Connection
	postgresConn, err := connector.NewPostgresConnection(cfg.Postgres)
	if err != nil {
		l.Log.Errorf("Не удалось установить соединение соединение с postgres: %v", err)
	}

	//resumeConn, err := connector.NewPostgresConnection(cfg.Postgres)
	//if err != nil {
	//	l.Log.Errorf("Не удалось установить соединение соединение с resume postgres: %v", err)
	//}
	//
	//vacancyConn, err := connector.NewPostgresConnection(cfg.Postgres)
	//if err != nil {
	//	l.Log.Errorf("Не удалось установить соединение соединение с vacancy postgres: %v", err)
	//}
	//
	//skillConn, err := connector.NewPostgresConnection(cfg.Postgres)
	//if err != nil {
	//	l.Log.Errorf("Не удалось установить соединение соединение с skill postgres: %v", err)
	//}
	//
	//specializationConn, err := connector.NewPostgresConnection(cfg.Postgres)
	//if err != nil {
	//	l.Log.Errorf("Не удалось установить соединение соединение с specialization postgres: %v", err)
	//}
	//
	//cityConn, err := connector.NewPostgresConnection(cfg.Postgres)
	//if err != nil {
	//	l.Log.Errorf("Не удалось установить соединение соединение с city postgres: %v", err)
	//}
	//
	//applicantConn, err := connector.NewPostgresConnection(cfg.Postgres)
	//if err != nil {
	//	l.Log.Errorf("Не удалось установить соединение соединение с applicant postgres: %v", err)
	//}
	//
	//employerConn, err := connector.NewPostgresConnection(cfg.Postgres)
	//if err != nil {
	//	l.Log.Errorf("Не удалось установить соединение соединение с employer postgres: %v", err)
	//}

	// Repositories Init
	resumeRepo, err := postgres.NewResumeRepository(postgresConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория резюме: %v", err)
	}

	vacancyRepo, err := postgres.NewVacancyRepository(postgresConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория вакансии: %v", err)
	}

	skillRepo, err := postgres.NewSkillRepository(postgresConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория навыка: %v", err)
	}

	specializationRepo, err := postgres.NewSpecializationRepository(postgresConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория специализации: %v", err)
	}

	cityRepo, err := postgres.NewCityRepository(postgresConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория города: %v", err)
	}

	applicantRepo, err := postgres.NewApplicantRepository(postgresConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория соискателя: %v", err)
	}

	employerRepo, err := postgres.NewEmployerRepository(postgresConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория работодателя: %v", err)
	}

	notificationRepo, err := postgres.NewNotificationRepository(postgresConn)
	if err != nil {
		l.Log.Errorf("Ошибка создания репозитория уведомлений: %v", err)
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

	resumeService := service.NewResumeService(resumeRepo, skillRepo, specializationRepo, applicantRepo, applicantService, cfg.Resume)
	vacancyService := service.NewVacanciesService(vacancyRepo, applicantRepo, specializationRepo, employerService)
	notificationService := service.NewNotificationService(notificationRepo)

	// Transport Init
	wsPool := ws.NewWebsocketPool(authService)

	authHandler := handler.NewAuthHandler(authService, cfg.CSRF)
	applicantHandler := handler.NewApplicantHandler(authService, applicantService, cfg.CSRF)
	employmentHandler := handler.NewEmployerHandler(authService, employerService, cfg.CSRF)
	resumeHandler := handler.NewResumeHandler(authService, resumeService, cfg.CSRF, wsPool, notificationService)
	vacancyHandler := handler.NewVacancyHandler(authService, vacancyService, cfg.CSRF, wsPool, notificationService)
	specializationHandler := handler.NewSpecializationHandler(specializationService)
	notificationHandler := handler.NewNotificationHandler(notificationService, authService, wsPool)
	// Metrics Init
	metrics.Init("resumatch")

	// Server Init
	srv := server.NewServer(cfg, wsPool)

	// Router config
	srv.SetupRoutes(func(r *http.ServeMux) {
		authHandler.Configure(r)
		applicantHandler.Configure(r)
		employmentHandler.Configure(r)
		resumeHandler.Configure(r)
		vacancyHandler.Configure(r)
		specializationHandler.Configure(r)
		notificationHandler.Configure(r)
	})

	return srv
}
