package dto

import "time"

// easyjson:json
type EmployerProfileResponse struct {
	ID           int       `json:"id"`
	CompanyName  string    `json:"company_name"`
	LegalAddress string    `json:"legal_address"`
	Email        string    `json:"email"`
	Slogan       string    `json:"slogan"`
	Website      string    `json:"website"`
	Description  string    `json:"description"`
	Vk           string    `json:"vk"`
	Telegram     string    `json:"telegram"`
	Facebook     string    `json:"facebook"`
	LogoPath     string    `json:"logo_path"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// easyjson:json
type EmployerProfileUpdate struct {
	CompanyName  string `json:"company_name" valid:"runelength(2|30),optional"`
	LegalAddress string `json:"legal_address" valid:"runelength(5|100),optional"`
	Slogan       string `json:"slogan" valid:"runelength(10|255),optional"`
	Website      string `json:"website" valid:"url,runelength(7|128),optional"`
	Description  string `json:"description" valid:"runelength(0|2000),optional"`
	Vk           string `json:"vk" valid:"url,runelength(7|128),optional"`
	Telegram     string `json:"telegram" valid:"url,runelength(7|128),optional"`
	Facebook     string `json:"facebook" valid:"url,runelength(7|128),optional"`
}
