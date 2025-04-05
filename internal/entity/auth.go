package entity

import (
	"bytes"
	"crypto/rand"
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
		return NewClientError("пароль должен содержать не менее 8 символов", ErrBadRequest)

	case len(password) > 32:
		return NewClientError("пароль должен содержать не более 32 символов", ErrBadRequest)

	case !regexp.MustCompile(`^[!@#$%^&*\w]+$`).MatchString(password):
		return NewClientError("пароль должен состоять из латинских букв, цифр и специальных символов !@#$%^&*", ErrBadRequest)

	case !regexp.MustCompile(`[A-Z]`).MatchString(password):
		return NewClientError("пароль должен содержать как минимум одну заглавную букву", ErrBadRequest)

	case !regexp.MustCompile(`[a-z]`).MatchString(password):
		return NewClientError("пароль должен содержать как минимум одну строчную букву", ErrBadRequest)

	case !regexp.MustCompile(`\d`).MatchString(password):
		return NewClientError("пароль должен содержать как минимум одну цифру", ErrBadRequest)

	case !regexp.MustCompile(`[!@#$%^&*]`).MatchString(password):
		return NewClientError("пароль должен содержать как минимум один из специальных символов !@#$%^&*", ErrBadRequest)

	default:
		return nil
	}
}

func HashPassword(password string) (salt []byte, hash []byte, err error) {
	salt = make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, nil, NewClientError("Ошибка при хешировании пароля", ErrInternal)
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
		return NewClientError("Невалидная почта", ErrBadRequest)
	}
	if len(email) > 255 {
		return NewClientError("Почта не может быть длиннее 255 символов", ErrBadRequest)
	}
	return nil
}
