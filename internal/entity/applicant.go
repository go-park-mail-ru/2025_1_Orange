package entity

import (
	"fmt"
	"time"
)

type ApplicantStatus string

const (
	StatusActivelySearching ApplicantStatus = "actively_searching"
	StatusOpenToOffers      ApplicantStatus = "open_to_offers"
	StatusConsideringOffer  ApplicantStatus = "considering_offer"
	StatusStartingSoon      ApplicantStatus = "starting_soon"
	StatusNotSearching      ApplicantStatus = "not_searching"
)

type Applicant struct {
	ID           int             `db:"id"`
	FirstName    string          `db:"first_name"`
	LastName     string          `db:"last_name"`
	MiddleName   string          `db:"middle_name"`
	Email        string          `db:"email"`
	CityID       int             `db:"city_id"`
	BirthDate    time.Time       `db:"birth_date"`
	Sex          string          `db:"sex"` // "M" или "F"
	Status       ApplicantStatus `db:"status"`
	Quote        string          `db:"quote"`
	Vk           string          `db:"vk"`
	Telegram     string          `db:"telegram"`
	Facebook     string          `db:"facebook"`
	AvatarID     int             `db:"avatar_id"`
	PasswordHash []byte          `db:"-"`
	PasswordSalt []byte          `db:"-"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
}

func ValidateStatus(status string) error {
	if !validStatus(status) {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("неверный статус соискателя"),
		)
	}
	return nil
}

func validStatus(status string) bool {
	switch status {
	case string(StatusActivelySearching),
		string(StatusOpenToOffers),
		string(StatusConsideringOffer),
		string(StatusStartingSoon),
		string(StatusNotSearching):
		return true
	}
	return false
}

func ValidateBirthDate(birthDate time.Time) error {
	if birthDate.After(time.Now()) {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("дата рождения не может быть позже текущей даты"),
		)
	}
	return nil
}
