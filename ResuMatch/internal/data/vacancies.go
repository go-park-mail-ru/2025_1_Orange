// package data

// import (
// 	"ResuMatch/internal/models"
// 	"time"
// )

// var Vacancies = []models.Vacancy{
// 	{
// 		ID:          1,
// 		Title:       "Golang Developer",
// 		Company:     "TechCorp",
// 		Location:    "Москва",
// 		Salary:      "200000 RUB",
// 		Description: "Разработка серверных API.",
// 		CreatedAt:   time.Now(),
// 		UpdatedAt:   time.Now(),
// 		Active:      true,
// 		PostedBy:    101,
// 		EmployerID:  1,
// 	},
// 	{
// 		ID:          2,
// 		Title:       "Frontend Developer",
// 		Company:     "WebSoft",
// 		Location:    "Санкт-Петербург",
// 		Salary:      "180000 RUB",
// 		Description: "Разработка интерфейсов.",
// 		CreatedAt:   time.Now(),
// 		UpdatedAt:   time.Now(),
// 		Active:      true,
// 		PostedBy:    102,
// 		EmployerID:  1,
// 	},
// 	{
// 		ID:          3,
// 		Title:       "Data Scientist",
// 		Company:     "AI Solutions",
// 		Location:    "Новосибирск",
// 		Salary:      "250000 RUB",
// 		Description: "Работа с ML-моделями, анализ больших данных.",
// 		CreatedAt:   time.Now(),
// 		UpdatedAt:   time.Now(),
// 		Active:      true,
// 		PostedBy:    103,
// 		EmployerID:  2,
// 	},
// 	{
// 		ID:          4,
// 		Title:       "DevOps Engineer",
// 		Company:     "CloudTech",
// 		Location:    "Казань",
// 		Salary:      "230000 RUB",
// 		Description: "Автоматизация CI/CD, работа с Kubernetes, Docker.",
// 		CreatedAt:   time.Now(),
// 		UpdatedAt:   time.Now(),
// 		Active:      false, // Вакансия неактивна
// 		PostedBy:    104,
// 		EmployerID:  3,
// 	},
// 	{
// 		ID:          5,
// 		Title:       "Backend Developer (Python)",
// 		Company:     "FinTech Group",
// 		Location:    "Екатеринбург",
// 		Salary:      "210000 RUB",
// 		Description: "Разработка высоконагруженных финансовых сервисов.",
// 		CreatedAt:   time.Now(),
// 		UpdatedAt:   time.Now(),
// 		Active:      true,
// 		PostedBy:    105,
// 		EmployerID:  4,
// 	},
// 	{
// 		ID:          6,
// 		Title:       "Mobile Developer (Flutter)",
// 		Company:     "AppWorks",
// 		Location:    "Челябинск",
// 		Salary:      "190000 RUB",
// 		Description: "Разработка кроссплатформенных мобильных приложений.",
// 		CreatedAt:   time.Now(),
// 		UpdatedAt:   time.Now(),
// 		Active:      true,
// 		PostedBy:    106,
// 		EmployerID:  5,
// 	},
// }

package data

import "ResuMatch/internal/models"

var Vacancies = []models.Vacancy{
	{
		ID:         12,
		Profession: "Фулстек веб разработчик",
		Salary:     "120000 - 200000 ₽ в месяц",
		Company:    "VK",
		City:       "Москва",
		Badges: []models.Badge{
			{Name: "Удаленно"},
			{Name: "40 часов в неделю"},
		},
		DayCreated: 1,
		Count:      12,
	},
	{
		ID:         13,
		Profession: "Мобильный разработчик (iOS)",
		Salary:     "150000 - 220000 ₽ в месяц",
		Company:    "Yandex",
		City:       "Санкт-Петербург",
		Badges: []models.Badge{
			{Name: "Гибкий график"},
			{Name: "Оплата обедов"},
		},
		DayCreated: 3,
		Count:      8,
	},
	{
		ID:         14,
		Profession: "Аналитик данных",
		Salary:     "180000 - 250000 ₽ в месяц",
		Company:    "Sber",
		City:       "Москва",
		Badges: []models.Badge{
			{Name: "Удаленная работа"},
			{Name: "ДМС"},
		},
		DayCreated: 5,
		Count:      15,
	},
}
