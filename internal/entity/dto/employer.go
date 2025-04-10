package dto

import "time"

type EmployerProfileResponse struct {
	ID           int       `json:"id"`
	CompanyName  string    `json:"company_name"`
	LegalAddress string    `json:"legal_address"`
	Email        string    `json:"email"`
	Slogan       string    `json:"slogan"`
	Website      string    `json:"website"`
	Description  string    `json:"description"`
	LogoPath     string    `json:"logo_path"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type EmployerProfileUpdate struct {
	CompanyName  string `json:"company_name"`
	LegalAddress string `json:"legal_address"`
	Slogan       string `json:"slogan"`
	Website      string `json:"website"`
	Description  string `json:"description"`
}
