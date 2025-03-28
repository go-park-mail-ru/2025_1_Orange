package usecase

import (
	"ResuMatch/internal/domain/mocks"
	"ResuMatch/internal/repository/mysql"
	"ResuMatch/internal/repository/redis"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type IUserUsecase interface {
	SignUp(ctx context.Context, user *mocks.DBUser, expiryUnixSeconds int64) (string, error)
	Login(ctx context.Context, email, password string, expiryUnixSeconds int64) (string, error)
	Logout(ctx context.Context, sid string) error
}

type AuthUsecase struct {
	userRepository    mysql.IUserRepository
	sessionRepository redis.ISessionRepository
}

func NewUserUsecase(
	userRepository mysql.IUserRepository,
	sessionRepository redis.ISessionRepository,
) IUserUsecase {
	return &AuthUsecase{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
	}
}

func (au *AuthUsecase) SignUp(ctx context.Context, user *mocks.DBUser, expiryUnixSeconds int64) (string, error) {
	if err := validateSignUpData(user); err != nil {
		return "", err
	}

	err := au.userRepository.AddUser(ctx, user)
	if err != nil {
		return "", err
	}

	userID, err := au.userRepository.GetUserIdByEmail(ctx, user.Email)
	if err != nil {
		return "", err
	}

	sessionID := uuid.NewString()
	session := &mocks.Session{
		SID:     sessionID,
		UserID:  userID,
		Expires: time.Now().Add(time.Second * time.Duration(expiryUnixSeconds)),
	}

	err = au.sessionRepository.CreateSession(session)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func (au *AuthUsecase) Login(ctx context.Context, email, password string, expiryUnixSeconds int64) (string, error) {
	// Проверка валидности email и пароля
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return "", errors.New("email and password must not be empty")
	}

	// Проверка пользователя в базе данных
	userID, err := au.userRepository.GetUserIdByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	err = au.userRepository.CheckPasswordById(ctx, userID, password)
	if err != nil {
		return "", err
	}
	// Создание новой сессии
	sessionID := uuid.NewString()
	session := &mocks.Session{
		SID:     sessionID,
		UserID:  userID,
		Expires: time.Now().Add(time.Second * time.Duration(expiryUnixSeconds)),
	}

	err = au.sessionRepository.CreateSession(session)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func (au *AuthUsecase) Logout(ctx context.Context, sid string) error {

	err := au.sessionRepository.DeleteSession(sid)
	if err != nil {
		return err
	}

	return nil
}

// func (au *AuthUsecase) GetInfo(ctx context.Context) (*mocks.DBUser, error) {
// 	user, err := au.userRepository.GetUserInfo(ctx, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return user, nil
// }

func validateSignUpData(user *mocks.DBUser) error {
	if strings.TrimSpace(user.Email) == "" {
		return errors.New("email is required")
	}

	if len(user.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	return nil
}
