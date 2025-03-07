package models

type User struct {
	Id               uint64 `json:"id"`
	Name             string `json:"name"`
	Birthdate        string `json:"birth_date"`
	Photo            string `json:"photo"`
	Login            string `json:"login"`
	Password         string `json:"password"`
	RegistrationDate string `json:"registration_date"`
	Email            string `json:"email"`
	Role             string `json:"role"`
}
