package server

import (
	_ "ResuMatch/docs"
	"ResuMatch/internal/config"
	"ResuMatch/internal/middleware"
	"context"
	swagger "github.com/swaggo/http-swagger"
	"net/http"
	"time"
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
	mainRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", subrouter))

	mainRouter.HandleFunc("/swagger/", swagger.WrapHandler)
	routeConfig(subrouter)

	subrouter.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("/app/assets"))))

	handler := middleware.CreateMiddlewareChain(
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
