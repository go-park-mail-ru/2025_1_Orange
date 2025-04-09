package entity

import (
	"fmt"
	"time"
	"unicode/utf8"
)

type Employer struct {
	ID           int       `db:"id"`
	CompanyName  string    `db:"company_name"`
	LegalAddress string    `db:"legal_address"`
	Email        string    `db:"email"`
	Slogan       string    `db:"slogan"`
	Website      string    `db:"website"`
	Description  string    `db:"description"`
	LogoID       int       `db:"logo_id"`
	PasswordHash []byte    `db:"-"`
	PasswordSalt []byte    `db:"-"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func ValidateCompanyName(companyName string) error {
	if utf8.RuneCountInString(companyName) > 64 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("название компании не может быть длиннее 64 символов"),
		)
	}
	return nil
}

func ValidateLegalAddress(legalAddress string) error {
	if utf8.RuneCountInString(legalAddress) > 255 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("юридический адрес компании не может быть длиннее 255 символов"),
		)
	}
	return nil
}

func ValidateSlogan(slogan string) error {
	if utf8.RuneCountInString(slogan) > 255 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("слоган компании не может быть длиннее 255 символов"),
		)
	}
	return nil
}

func ValidateWebsite(website string) error {
	if utf8.RuneCountInString(website) > 64 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("адрес сайта компании не может быть длиннее 255 символов"),
		)
	}
	return nil
}

func ValidateDescription(description string) error {
	if utf8.RuneCountInString(description) > 2000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("описание компании не может быть длиннее 2000 символов"),
		)
	}
	return nil
}
