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
			Addr:           ":" + cfg.HTTP.Port,
			ReadTimeout:    cfg.HTTP.ReadTimeout,
			WriteTimeout:   cfg.HTTP.WriteTimeout,
			MaxHeaderBytes: cfg.HTTP.MaxHeaderBytes,
		},
	}
}

//func (s *Server) SetupRoutes(routeConfig func(*mux.Router)) {
//	router := mux.NewRouter()
//
//	// Middleware
//	csrf := middleware.NewCSRF(s.config.Secrets.CSRF, "__csrf_token", s.config.Cookies)
//	cors := middleware.CORS(s.config.HTTP.CORSAllowedOrigins)
//	router.Use(cors, csrf.CSRFMiddleware)
//
//	// Router config
//	routeConfig(router)
//
//	s.httpServer.Handler = router
//}

func (s *Server) SetupRoutes(routeConfig func(*httprouter.Router)) {
	router := httprouter.New()

	// Middleware chain
	handler := middleware.CORS(s.config.HTTP.CORSAllowedOrigins)(router)
	csrf := middleware.NewCSRF(s.config.Secrets.CSRF, "__csrf_token", s.config.Cookies)
	handler = csrf.CSRFMiddleware(handler)

	// Router config
	routeConfig(router)

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
