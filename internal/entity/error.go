package entity

import (
	"errors"
	"fmt"
)

type ClientError struct {
	ErrorMessage string
	Data         error
}

func (err ClientError) Error() string {
	return fmt.Sprint(err.ErrorMessage)
}

func NewClientError(msg string, err error) error {
	return ClientError{
		ErrorMessage: msg,
		Data:         err,
	}
}

var (
	ErrBadRequest    = errors.New("bad request")
	ErrForbidden     = errors.New("forbidden")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrRedis         = errors.New("redis error")
	ErrPostgres      = errors.New("postgres error")
	ErrInternal      = errors.New("internal server error")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
)
