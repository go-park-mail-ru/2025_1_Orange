package service

import (
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
)

type AuthService struct {
	sessionRepository repository.SessionRepository
}

func NewAuthService(sessionRepository repository.SessionRepository) usecase.Auth {
	return &AuthService{
		sessionRepository: sessionRepository,
	}
}

func (a AuthService) Logout(s string) error {
	if err := a.sessionRepository.DeleteSession(s); err != nil {
		return err
	}
	return nil
}

func (a AuthService) LogoutAll(userID int, role string) error {
	if err := a.sessionRepository.DeleteAllSessions(userID, role); err != nil {
		return err
	}
	return nil
}

func (a AuthService) GetUserIDBySession(session string) (int, string, error) {
	userID, role, err := a.sessionRepository.GetSession(session)
	if err != nil {
		return 0, "", err
	}
	return userID, role, nil
}

func (a AuthService) CreateSession(userID int, role string) (string, error) {
	session, err := a.sessionRepository.CreateSession(userID, role)
	if err != nil {
		return "", err
	}
	return session, nil
}
