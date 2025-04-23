package entity

import (
	"time"
)

type Employer struct {
	ID           int       `db:"id"`
	CompanyName  string    `db:"company_name"`
	LegalAddress string    `db:"legal_address"`
	Email        string    `db:"email"`
	Slogan       string    `db:"slogan"`
	Website      string    `db:"website"`
	Vk           string    `db:"vk"`
	Telegram     string    `db:"telegram"`
	Facebook     string    `db:"facebook"`
	Description  string    `db:"description"`
	LogoID       int       `db:"logo_id"`
	PasswordHash []byte    `db:"-"`
	PasswordSalt []byte    `db:"-"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
