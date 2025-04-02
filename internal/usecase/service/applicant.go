package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"context"
)

type ApplicantService struct {
	applicantRepository repository.ApplicantRepository
}

func NewApplicantService(applicantRepository repository.ApplicantRepository) usecase.Applicant {
	return &ApplicantService{
		applicantRepository: applicantRepository,
	}
}

func (a *ApplicantService) Register(registerDTO *dto.ApplicantRegister) (int, error) {
	applicant := new(entity.Applicant)

	if err := entity.ValidateEmail(registerDTO.Email); err != nil {
		return 0, err
	}
	if err := entity.ValidatePassword(registerDTO.Password); err != nil {
		return 0, err
	}

	salt, hash, err := entity.HashPassword(registerDTO.Password)
	if err != nil {
		return 0, err
	}

	applicant.Email = registerDTO.Email
	applicant.PasswordHash = hash
	applicant.PasswordSalt = salt
	applicant.FirstName = registerDTO.FirstName
	applicant.LastName = registerDTO.LastName

	applicant, err = a.applicantRepository.Create(context.Background(), applicant)
	if err != nil {
		return -1, err
	}
	return applicant.ID, nil
}

func (a *ApplicantService) Login(loginDTO *dto.ApplicantLogin) (int, error) {
	if err := entity.ValidateEmail(loginDTO.Email); err != nil {
		return 0, err
	}

	applicant, err := a.applicantRepository.GetByEmail(context.Background(), loginDTO.Email)
	if err != nil {
		return 0, err
	}

	if applicant.CheckPassword(loginDTO.Password) {
		return applicant.ID, nil
	}
	return 0, entity.NewClientError("Неверный пароль", entity.ErrForbidden)
}

func (a *ApplicantService) GetUser(applicantID int) (*dto.ApplicantProfile, error) {
	applicant, err := a.applicantRepository.GetByID(context.Background(), applicantID)
	if err != nil {
		return nil, err
	}
	return &dto.ApplicantProfile{
		ID:        applicant.ID,
		FirstName: applicant.FirstName,
		LastName:  applicant.LastName,
		Email:     applicant.Email,
	}, nil

}
