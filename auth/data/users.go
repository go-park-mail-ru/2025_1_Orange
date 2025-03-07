package data

import (
	"auth/models"
)

var Users = map[string]models.User{
	"rvasily": {
		Id:               1,
		Name:             "Василий Рогов",
		Birthdate:        "1980-01-01",
		Photo:            "",
		Login:            "rvasily",
		Password:         "love",
		RegistrationDate: "2023-01-01",
		Email:            "rvasily@example.com",
		Role:             "user",
	},
	"elena.smirnova": {
		Id:               2,
		Name:             "Елена Смирнова",
		Birthdate:        "1985-05-20",
		Photo:            "https://example.com/images/elena_smirnova.jpg",
		Login:            "elena.smirnova",
		Password:         "password123",
		RegistrationDate: "2023-02-15",
		Email:            "elena.smirnova@example.com",
		Role:             "admin",
	},
	"john.doe": {
		Id:               3,
		Name:             "Джон Доу",
		Birthdate:        "1992-11-10",
		Photo:            "",
		Login:            "john.doe",
		Password:         "securepass",
		RegistrationDate: "2023-03-01",
		Email:            "john.doe@example.net",
		Role:             "moderator",
	},
	"jane.smith": {
		Id:               4,
		Name:             "Джейн Смит",
		Birthdate:        "1975-07-04",
		Photo:            "https://example.com/images/jane_smith.png",
		Login:            "jane.smith",
		Password:         "strongpassword",
		RegistrationDate: "2023-04-22",
		Email:            "jane.smith@example.org",
		Role:             "user",
	},
	"peter.jones": {
		Id:               5,
		Name:             "Питер Джонс",
		Birthdate:        "1988-09-28",
		Photo:            "",
		Login:            "peter.jones",
		Password:         "anotherpass",
		RegistrationDate: "2023-05-10",
		Email:            "peter.jones@example.com",
		Role:             "user",
	},
}
