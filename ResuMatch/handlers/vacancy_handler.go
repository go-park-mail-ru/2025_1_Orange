package handlers

import (
	"ResuMatch/data"
	"encoding/json"
	"net/http"
)

func GetVacancies(w http.ResponseWriter, r *http.Request) {

	// Проверяем, что запрос идет методом GET
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Устанавливаем заголовки ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Кодируем данные в JSON и отправляем
	if err := json.NewEncoder(w).Encode(data.Vacancies); err != nil {
		http.Error(w, "Не удалось закодировать ответ", http.StatusInternalServerError)
	}
}
