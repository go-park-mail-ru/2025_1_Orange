package entity

import (
	"fmt"
	"unicode/utf8"
)

type Employer struct {
	ID           int    `json:"id"`
	CompanyName  string `json:"company_name"`
	LegalAddress string `json:"legal_address,omitempty"`
	AuthBase
}

func (e *Employer) Validate() error {
	if err := ValidateEmail(e.Email); err != nil {
		return err
	}
	if utf8.RuneCountInString(e.CompanyName) > 64 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("название компании не может быть длиннее 64 символов"),
		)
	}

	if utf8.RuneCountInString(e.LegalAddress) > 255 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("юридический адрес компании не может быть длиннее 255 символов"),
		)
	}

	return nil
}
