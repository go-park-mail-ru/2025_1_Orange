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
