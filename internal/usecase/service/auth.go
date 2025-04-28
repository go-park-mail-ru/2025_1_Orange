package service

import (
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"context"
)

type AuthService struct {
	sessionRepository repository.SessionRepository
}

func NewAuthService(
	sessionRepo repository.SessionRepository,
) usecase.Auth {
	return &AuthService{
		sessionRepository: sessionRepo,
	}
}

func (a *AuthService) Logout(ctx context.Context, session string) error {
	if err := a.sessionRepository.DeleteSession(ctx, session); err != nil {
		return err
	}
	return nil
}

func (a *AuthService) LogoutAll(ctx context.Context, userID int, role string) error {
	if err := a.sessionRepository.DeleteAllSessions(ctx, userID, role); err != nil {
		return err
	}
	return nil
}

func (a *AuthService) GetUserIDBySession(ctx context.Context, session string) (int, string, error) {
	userID, role, err := a.sessionRepository.GetSession(ctx, session)
	if err != nil {
		return -1, "", err
	}
	return userID, role, nil
}

func (a *AuthService) CreateSession(ctx context.Context, userID int, role string) (string, error) {
	session, err := a.sessionRepository.CreateSession(ctx, userID, role)
	if err != nil {
		return "", err
	}
	return session, nil
}
