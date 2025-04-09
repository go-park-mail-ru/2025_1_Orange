package entity

import (
	"errors"
)

var (
	ErrBadRequest    = errors.New("bad request")
	ErrForbidden     = errors.New("forbidden")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInternal      = errors.New("internal server error")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
)

const (
	PSQLUniqueViolation   = "23505"
	PSQLNotNullViolation  = "23502"
	PSQLDatatypeViolation = "22P02"
	PSQLCheckViolation    = "23514"
)

type Error struct {
	clientErr   error
	internalErr error
}

func NewError(clientErr, internalErr error) error {
	return Error{
		clientErr:   clientErr,
		internalErr: internalErr,
	}
}

func (e Error) Error() string {
	return errors.Join(e.clientErr, e.internalErr).Error()
}

func (e Error) InternalErr() error {
	return e.internalErr
}

func (e Error) ClientErr() error {
	return e.clientErr
}
