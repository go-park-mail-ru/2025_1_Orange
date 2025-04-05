package server

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/middleware"
	"context"
	"github.com/julienschmidt/httprouter"
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

func (s *Server) SetupRoutes(routeConfig func(*httprouter.Router)) {
	router := httprouter.New()

	// Middleware chain
	handler := middleware.CreateMiddlewareChain(
		middleware.CORS(s.config.HTTP.CORSAllowedOrigins),
		middleware.CSRFMiddleware(s.config.CSRF),
	)
	//handler := middleware.CORS(s.config.HTTP.CORSAllowedOrigins)(router)
	//csrf := middleware.CSRFMiddleware(s.config.CSRF)
	//handler = csrf(handler)

	// Router config
	routeConfig(router)

	s.httpServer.Handler = handler(router)
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
