package router

import (
	"ResuMatch/handlers"
	"ResuMatch/repository/profile"
	"ResuMatch/repository/session"
	"ResuMatch/usecase"
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	sessionRepo := session.SessionRepo{}
	userRepo := profile.UserRepo{}
	core := usecase.NewCore(sessionRepo, userRepo)

	mux.HandleFunc("/login", handlers.NewMyHandler(core).Signin)
	mux.HandleFunc("/login", handlers.NewMyHandler(core).Signup)
	mux.HandleFunc("/login", handlers.NewMyHandler(core).Logout)
	mux.HandleFunc("/vacancies", handlers.GetVacancies)

	return mux
}
