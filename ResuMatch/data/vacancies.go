package data

import (
	"ResuMatch/models"
	"time"
)

var Vacancies = []models.Vacancy{
	{
		ID:          1,
		Title:       "Golang Developer",
		Company:     "TechCorp",
		Location:    "Москва",
		Salary:      "200000 RUB",
		Description: "Разработка серверных API.",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Active:      true,
		PostedBy:    101,
	},
	{
		ID:          2,
		Title:       "Frontend Developer",
		Company:     "WebSoft",
		Location:    "Санкт-Петербург",
		Salary:      "180000 RUB",
		Description: "Разработка интерфейсов.",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Active:      true,
		PostedBy:    102,
	},
	{
		ID:          3,
		Title:       "Data Scientist",
		Company:     "AI Solutions",
		Location:    "Новосибирск",
		Salary:      "250000 RUB",
		Description: "Работа с ML-моделями, анализ больших данных.",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Active:      true,
		PostedBy:    103,
	},
	{
		ID:          4,
		Title:       "DevOps Engineer",
		Company:     "CloudTech",
		Location:    "Казань",
		Salary:      "230000 RUB",
		Description: "Автоматизация CI/CD, работа с Kubernetes, Docker.",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Active:      false, // Вакансия неактивна
		PostedBy:    104,
	},
	{
		ID:          5,
		Title:       "Backend Developer (Python)",
		Company:     "FinTech Group",
		Location:    "Екатеринбург",
		Salary:      "210000 RUB",
		Description: "Разработка высоконагруженных финансовых сервисов.",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Active:      true,
		PostedBy:    105,
	},
	{
		ID:          6,
		Title:       "Mobile Developer (Flutter)",
		Company:     "AppWorks",
		Location:    "Челябинск",
		Salary:      "190000 RUB",
		Description: "Разработка кроссплатформенных мобильных приложений.",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Active:      true,
		PostedBy:    106,
	},
}
