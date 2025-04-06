package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"context"
	"fmt"
)

type EmployerService struct {
	employerRepository repository.EmployerRepository
}

func NewEmployerService(employerRepository repository.EmployerRepository) usecase.Employer {
	return &EmployerService{
		employerRepository: employerRepository,
	}
}

func (e *EmployerService) Register(ctx context.Context, registerDTO *dto.EmployerRegister) (int, error) {
	employer := new(entity.Employer)

	if err := entity.ValidateEmail(registerDTO.Email); err != nil {
		return -1, err
	}
	if err := entity.ValidatePassword(registerDTO.Password); err != nil {
		return -1, err
	}

	salt, hash, err := entity.HashPassword(registerDTO.Password)
	if err != nil {
		return -1, err
	}

	employer.Email = registerDTO.Email
	employer.PasswordHash = hash
	employer.PasswordSalt = salt
	employer.CompanyName = registerDTO.CompanyName
	employer.LegalAddress = registerDTO.LegalAddress

	employer, err = e.employerRepository.Create(ctx, employer)
	if err != nil {
		return -1, err
	}
	return employer.ID, nil
}

func (e *EmployerService) Login(ctx context.Context, loginDTO *dto.EmployerLogin) (int, error) {
	if err := entity.ValidateEmail(loginDTO.Email); err != nil {
		return -1, err
	}

	employer, err := e.employerRepository.GetByEmail(ctx, loginDTO.Email)
	if err != nil {
		return -1, err
	}
	if employer.CheckPassword(loginDTO.Password) {
		return employer.ID, nil
	}
	return -1, entity.NewError(
		entity.ErrForbidden,
		fmt.Errorf("неверный пароль"),
	)
}

func (e *EmployerService) GetUser(ctx context.Context, employerID int) (*dto.EmployerProfile, error) {
	employer, err := e.employerRepository.GetByID(ctx, employerID)
	if err != nil {
		return nil, err
	}
	return &dto.EmployerProfile{
		ID:           employer.ID,
		CompanyName:  employer.CompanyName,
		Email:        employer.Email,
		LegalAddress: employer.LegalAddress,
	}, nil
}
