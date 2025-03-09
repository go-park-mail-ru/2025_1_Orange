package router

import (
	"ResuMatch/internal/handlers/auth"
	"ResuMatch/internal/handlers/vacancy"
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

	mux.HandleFunc("/signin", auth.NewMyHandler(core).Signin)
	mux.HandleFunc("/signup", auth.NewMyHandler(core).Signup)
	mux.HandleFunc("/logout", auth.NewMyHandler(core).Logout)
	mux.HandleFunc("/auth", auth.NewMyHandler(core).Auth)
	mux.HandleFunc("/check-email", auth.NewMyHandler(core).CheckEmail)
	mux.HandleFunc("/vacancies", vacancy.GetVacancies)

	return mux
}
