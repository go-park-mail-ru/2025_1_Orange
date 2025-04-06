package entity

import (
	"fmt"
	"unicode/utf8"
)

type Applicant struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AuthBase
}

func (a *Applicant) Validate() error {
	if err := ValidateEmail(a.Email); err != nil {
		return err
	}

	if utf8.RuneCountInString(a.FirstName) > 30 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("имя не может быть длиннее 30 символов"),
		)
	}

	if utf8.RuneCountInString(a.LastName) > 30 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("фамилия не может быть длиннее 30 символов"),
		)
	}

	return nil
}
