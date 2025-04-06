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

type Error struct {
	svcErr error
	appErr error
}

func NewError(appErr, svcErr error) error {
	return Error{
		svcErr: appErr,
		appErr: svcErr,
	}
}

func (e Error) Error() string {
	return errors.Join(e.svcErr, e.appErr).Error()
}

func (e Error) AppErr() error {
	return e.appErr
}

func (e Error) SvcErr() error {
	return e.svcErr
}
