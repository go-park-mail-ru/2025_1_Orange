package router

import (
	"ResuMatch/internal/handlers/auth"
	"ResuMatch/internal/handlers/vacancy"
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/signin", auth.NewMyHandler().Signin)
	mux.HandleFunc("/signup", auth.NewMyHandler().Signup)
	mux.HandleFunc("/logout", auth.NewMyHandler().Logout)
	mux.HandleFunc("/auth", auth.NewMyHandler().Auth)
	mux.HandleFunc("/check-email", auth.NewMyHandler().CheckEmail)
	mux.HandleFunc("/vacancies", vacancy.GetVacancies)

	return mux
}
