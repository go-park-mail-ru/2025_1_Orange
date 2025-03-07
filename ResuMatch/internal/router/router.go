package router

import (
	"ResuMatch/internal/handlers"
	"ResuMatch/internal/repository/profile"
	"ResuMatch/internal/repository/session"
	"ResuMatch/internal/usecase"
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	sessionRepo := session.Sessionrepo{}
	userRepo := profile.UserRepo{}
	core := usecase.NewCore(sessionRepo, userRepo)

	mux.HandleFunc("/signin", handlers.NewMyHandler(core).Signin)
	mux.HandleFunc("/signup", handlers.NewMyHandler(core).Signup)
	mux.HandleFunc("/logout", handlers.NewMyHandler(core).Logout)
	mux.HandleFunc("/vacancies", handlers.GetVacancies)

	return mux
}
