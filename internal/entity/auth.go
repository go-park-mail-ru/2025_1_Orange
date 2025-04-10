package entity

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/argon2"
	"regexp"
)

const (
	TimeCost        = 2
	MemoryCost      = 64 * 1024
	ParallelThreads = 2
	HashLength      = 32
)

type AuthBase struct {
	Email        string `json:"email"`
	PasswordHash []byte `json:"-"`
	PasswordSalt []byte `json:"-"`
}

func ValidatePassword(password string) error {
	switch {
	case len(password) < 8:
		return NewError(
			ErrBadRequest,
			fmt.Errorf("пароль должен содержать не менее 8 символов"),
		)

	case len(password) > 32:
		return NewError(
			ErrBadRequest,
			fmt.Errorf("пароль должен содержать не более 32 символов"),
		)

	case !regexp.MustCompile(`^[!@#$%^&*\w]+$`).MatchString(password):
		return NewError(
			ErrBadRequest,
			fmt.Errorf("пароль должен состоять из латинских букв, цифр и специальных символов !@#$%^&*"),
		)

	case !regexp.MustCompile(`[A-Z]`).MatchString(password):
		return NewError(
			ErrBadRequest,
			fmt.Errorf("пароль должен содержать как минимум одну заглавную букву"),
		)

	case !regexp.MustCompile(`[a-z]`).MatchString(password):
		return NewError(
			ErrBadRequest,
			fmt.Errorf("пароль должен содержать как минимум одну строчную букву"),
		)

	case !regexp.MustCompile(`\d`).MatchString(password):
		return NewError(
			ErrBadRequest,
			fmt.Errorf("пароль должен содержать как минимум одну цифру"),
		)

	case !regexp.MustCompile(`[!@#$%^&*]`).MatchString(password):
		return NewError(
			ErrBadRequest,
			fmt.Errorf("пароль должен содержать как минимум один из специальных символов !@#$%^&*"),
		)

	default:
		return nil
	}
}

func HashPassword(password string) (salt []byte, hash []byte, err error) {
	salt = make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, nil, NewError(
			ErrInternal,
			fmt.Errorf("ошибка при хешировании пароля"),
		)
	}

	hash = argon2.IDKey(
		[]byte(password),
		salt,
		TimeCost,
		MemoryCost,
		ParallelThreads,
		HashLength,
	)
	return salt, hash, nil
}

func (a *AuthBase) CheckPassword(password string) bool {
	return bytes.Equal(
		argon2.IDKey(
			[]byte(password),
			a.PasswordSalt,
			TimeCost,
			MemoryCost,
			ParallelThreads,
			HashLength,
		),
		a.PasswordHash,
	)
}

func ValidateEmail(email string) error {
	re := regexp.MustCompile("^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}(?:\\.[A-Za-z]{2,})?$")
	if !re.MatchString(email) {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("невалидная почта"),
		)
	}

	if len(email) > 255 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("почта не может быть длиннее 255 символов"),
		)
	}
	return nil
}
