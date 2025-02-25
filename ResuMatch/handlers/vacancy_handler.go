package handlers

import (
	"ResuMatch/data"
	"encoding/json"
	"net/http"
)

func GetVacancies(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем заголовки ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Кодируем данные в JSON и отправляем
	json.NewEncoder(w).Encode(data.Vacancies)
}
