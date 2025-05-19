package dto

import (
	"time"
)

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
	Vk         string    `json:"vk"`
	Telegram   string    `json:"telegram"`
	Facebook   string    `json:"facebook"`
	AvatarPath string    `json:"avatar_path"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ApplicantProfileUpdate struct {
	FirstName  string    `json:"first_name" valid:"runelength(2|30),optional"`
	LastName   string    `json:"last_name" valid:"runelength(2|30),optional"`
	MiddleName string    `json:"middle_name" valid:"runelength(2|30),optional"`
	City       string    `json:"city" valid:"runelength(2|30),optional"`
	BirthDate  time.Time `json:"birth_date" valid:"-"`
	Sex        string    `json:"sex" valid:"in(M|F),optional"`
	Status     string    `json:"status" valid:"-"`
	Quote      string    `json:"quote" valid:"runelength(10|255),optional"`
	Vk         string    `json:"vk" valid:"url,runelength(7|128),optional"`
	Telegram   string    `json:"telegram" valid:"url,runelength(7|128),optional"`
	Facebook   string    `json:"facebook" valid:"url,runelength(7|128),optional"`
}

type ChatShortResponseApplicant struct {
	ID         int    `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	AvatarPath string `json:"avatar_path"`
}
