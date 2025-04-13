package entity

import (
	"fmt"
	"unicode/utf8"
)

type City struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func ValidateCity(city string) error {
	if utf8.RuneCountInString(city) > 30 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("название города не должно превышать 30 символов"),
		)
	}
	return nil
}
