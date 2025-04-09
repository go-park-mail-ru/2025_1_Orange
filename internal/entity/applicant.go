package entity

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
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
	AvatarPath   string          `db:"avatar_path"`
	PasswordHash []byte          `db:"-"`
	PasswordSalt []byte          `db:"-"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
}

func ValidateFirstName(firstName string) error {
	if utf8.RuneCountInString(firstName) > 30 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("имя не может быть длиннее 30 символов"),
		)
	}
	return nil
}

func ValidateLastName(lastName string) error {
	if utf8.RuneCountInString(lastName) > 30 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("фамилия не может быть длиннее 30 символов"),
		)
	}
	return nil
}

func ValidateMiddleName(middleName string) error {
	if middleName != "" && utf8.RuneCountInString(middleName) > 30 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("отчество не может быть длиннее 30 символов"),
		)
	}
	return nil
}

func ValidateSex(sex string) error {
	if sex != "" && !(sex == "M" || sex == "F") {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("пол должен быть 'M' или 'F'"),
		)
	}
	return nil
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

func ValidateQuote(quote string) error {
	if utf8.RuneCountInString(quote) > 255 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("цитата не может быть длиннее 255 символов"),
		)
	}
	return nil
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

func ValidateAvatar(fileHeader *multipart.FileHeader) error {
	if fileHeader.Size > 5<<20 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("размер изображения не должен превышать 5MB"),
		)
	}

	fileExt := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if fileExt != ".jpg" && fileExt != ".jpeg" && fileExt != ".png" {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("допустимы только форматы jpg, jpeg, png"),
		)
	}

	return nil
}
