package router

import (
	"ResuMatch/internal/handlers/auth"
	"ResuMatch/internal/handlers/vacancy"
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	handler := auth.NewMyHandler()

	mux.HandleFunc("/signin", handler.Signin)
	mux.HandleFunc("/signup", handler.Signup)
	mux.HandleFunc("/logout", handler.Logout)
	mux.HandleFunc("/auth", handler.Auth)
	mux.HandleFunc("/check-email", handler.CheckEmail)
	mux.HandleFunc("/vacancies", vacancy.GetVacancies)

	return mux
}
