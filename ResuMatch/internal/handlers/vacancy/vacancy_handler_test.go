package vacancy

import (
	"ResuMatch/internal/data"
	"ResuMatch/internal/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Проверка того, что GetVacancies возвращает правильный формат данных
func TestGetVacancies_Success(t *testing.T) {
	// Создаем запрос для теста
	req, err := http.NewRequest("GET", "/vacancies", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем записывающее средство для захвата ответа
	rr := httptest.NewRecorder()

	// Мокаем обработчик (в данном случае, реальный обработчик)
	handler := http.HandlerFunc(GetVacancies)

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем статус-код ответа
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Обработчик вернул неверный статус-код: получил %v, ожидался %v", status, http.StatusOK)
	}

	// Проверяем, что возвращаемые данные — это JSON
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Обработчик вернул некорректный заголовок Content-Type: получил %v, ожидался %v", contentType, "application/json")
	}

	// Проверяем содержимое ответа (корректность JSON)
	var vacancies []models.Vacancy
	if err := json.NewDecoder(rr.Body).Decode(&vacancies); err != nil {
		t.Fatalf("Не удалось декодировать тело ответа: %v", err)
	}

	// Проверяем количество вакансий
	if len(vacancies) != len(data.Vacancies) {
		t.Errorf("Обработчик вернул неправильное количество вакансий: получил %v, ожидалось %v", len(vacancies), len(data.Vacancies))
	}

	// // Дополнительные проверки для первого элемента вакансий
	// expectedVacancy := data.Vacancies[0]
	// if vacancies[0].ID != expectedVacancy.ID {
	// 	t.Errorf("Ожидался ID вакансии %v, но получен %v", expectedVacancy.ID, vacancies[0].ID)
	// }
	// if vacancies[0].Title != expectedVacancy.Title {
	// 	t.Errorf("Ожидалось название вакансии %v, но получено %v", expectedVacancy.Title, vacancies[0].Title)
	// }
	// if vacancies[0].Company != expectedVacancy.Company {
	// 	t.Errorf("Ожидалась компания %v, но получена %v", expectedVacancy.Company, vacancies[0].Company)
	// }
	// // Можно дальше добавить еще проверок, если они нужны ...
}

// Тест на неправильный HTTP-метод (POST вместо GET)
func TestGetVacancies_WrongMethod(t *testing.T) {
	req, err := http.NewRequest("POST", "/vacancies", nil)
	if err != nil {
		t.Fatalf("Ошибка при создании запроса: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetVacancies)
	handler.ServeHTTP(rr, req)

	// Ожидаем, что сервер вернет ошибку
	if rr.Code == http.StatusOK {
		t.Errorf("Ожидался код ошибки, но сервер вернул 200 OK")
	}
}
