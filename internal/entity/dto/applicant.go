package dto

import "time"

type ApplicantProfileResponse struct {
	ID         int       `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	MiddleName string    `json:"middle_name"`
	City       string    `json:"city"`
	BirthDate  time.Time `json:"birth_date"`
	Sex        string    `json:"sex"`
	Email      string    `json:"email"`
	Status     string    `json:"status"`
	Quote      string    `json:"quote"`
	AvatarPath string    `json:"avatar_path"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ApplicantProfileUpdate struct {
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	MiddleName string    `json:"middle_name"`
	City       string    `json:"city"`
	BirthDate  time.Time `json:"birth_date"`
	Sex        string    `json:"sex"`
	Status     string    `json:"status"`
	Quote      string    `json:"quote"`
}
