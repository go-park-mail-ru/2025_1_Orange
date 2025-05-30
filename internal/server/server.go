package server

import (
	_ "ResuMatch/docs"
	"ResuMatch/internal/config"
	"ResuMatch/internal/middleware"
	"context"
	"net/http"
	"time"

	swagger "github.com/swaggo/http-swagger"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
		httpServer: &http.Server{
			Addr:           cfg.HTTP.Host + ":" + cfg.HTTP.Port,
			ReadTimeout:    cfg.HTTP.ReadTimeout,
			WriteTimeout:   cfg.HTTP.WriteTimeout,
			MaxHeaderBytes: cfg.HTTP.MaxHeaderBytes,
		},
	}
}

func (s *Server) SetupRoutes(routeConfig func(*http.ServeMux)) {
	subrouter := http.NewServeMux()

	mainRouter := http.NewServeMux()
	mainRouter.Handle("/metrics", middleware.PrometheusHandler())
	mainRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", subrouter))

	mainRouter.HandleFunc("/api/v1/swagger/", swagger.WrapHandler)
	routeConfig(subrouter)

	handler := middleware.CreateMiddlewareChain(
		middleware.MetricsMiddleware(),
		middleware.RecoveryMiddleware(),
		middleware.CORS(s.config.HTTP.CORSAllowedOrigins),
		middleware.CSRFMiddleware(s.config.CSRF),
		middleware.RequestIDMiddleware(),
		middleware.AccessLogMiddleware(),
	)(mainRouter)

	s.httpServer.Handler = handler
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
