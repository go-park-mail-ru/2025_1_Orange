package router

import (
	"ResuMatch/internal/handlers"
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/vacancies", handlers.GetVacancies)

	return mux
}
