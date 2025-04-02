package entity

import "unicode/utf8"

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
		return NewClientError("Имя не может быть длиннее 30 символов", ErrBadRequest)
	}
	if utf8.RuneCountInString(a.LastName) > 30 {
		return NewClientError("Фамилия не может быть длиннее 30 символов", ErrBadRequest)
	}

	return nil
}
