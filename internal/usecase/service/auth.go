package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"context"
	"errors"
)

type AuthService struct {
	sessionRepository   repository.SessionRepository
	applicantRepository repository.ApplicantRepository
	employerRepository  repository.EmployerRepository
}

func NewAuthService(
	sessionRepo repository.SessionRepository,
	applicantRepo repository.ApplicantRepository,
	employerRepo repository.EmployerRepository,
) usecase.Auth {
	return &AuthService{
		sessionRepository:   sessionRepo,
		applicantRepository: applicantRepo,
		employerRepository:  employerRepo,
	}
}

func (a *AuthService) EmailExists(ctx context.Context, email string) (*dto.EmailExistsResponse, error) {
	if err := entity.ValidateEmail(email); err != nil {
		return nil, err
	}

	applicant, err := a.applicantRepository.GetApplicantByEmail(ctx, email)
	if err == nil && applicant != nil {
		return &dto.EmailExistsResponse{
			Exists: true,
			Role:   "applicant",
		}, err
	}

	var e entity.Error
	if errors.As(err, &e) && !errors.Is(e.SvcErr(), entity.ErrNotFound) {
		return nil, err
	}

	employer, err := a.employerRepository.GetEmployerByEmail(ctx, email)
	if err == nil && employer != nil {
		return &dto.EmailExistsResponse{
			Exists: true,
			Role:   "employer",
		}, err
	}

	return nil, err
}

func (a *AuthService) Logout(s string) error {
	if err := a.sessionRepository.DeleteSession(s); err != nil {
		return err
	}
	return nil
}

func (a *AuthService) LogoutAll(userID int, role string) error {
	if err := a.sessionRepository.DeleteAllSessions(userID, role); err != nil {
		return err
	}
	return nil
}

func (a *AuthService) GetUserIDBySession(session string) (int, string, error) {
	userID, role, err := a.sessionRepository.GetSession(session)
	if err != nil {
		return -1, "", err
	}
	return userID, role, nil
}

func (a *AuthService) CreateSession(userID int, role string) (string, error) {
	session, err := a.sessionRepository.CreateSession(userID, role)
	if err != nil {
		return "", err
	}
	return session, nil
}
