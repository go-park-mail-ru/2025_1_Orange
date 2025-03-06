package data

import (
	"auth/models"
)

var Users = []models.User{
	{
		Id:               1,
		Name:             "Елена Смирнова",
		Login:            "elena.smirnova",
		Password:         "hashed_password_1",
		Birthdate:        "1985-12-20",
		Photo:            "https://example.com/images/elena_smirnova.png",
		RegistrationDate: "2023-10-20",
		Email:            "elena.smirnova@example.org",
		Role:             "administrator",
	},
	{
		Id:               2,
		Name:             "Иван Иванов",
		Login:            "ivan.ivanov",
		Password:         "hashed_password_2",
		Birthdate:        "1990-05-15",
		Photo:            "https://example.com/images/ivan_ivanov.jpg",
		RegistrationDate: "2023-10-27",
		Email:            "ivan.ivanov@example.com",
		Role:             "user",
	},
	{
		Id:               3,
		Name:             "Джон Доу",
		Login:            "john.doe",
		Password:         "hashed_password_3",
		Birthdate:        "1978-08-01",
		Photo:            "",
		RegistrationDate: "2023-09-10",
		Email:            "john.doe@example.net",
		Role:             "moderator",
	},
}
