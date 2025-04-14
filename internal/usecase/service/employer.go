package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"ResuMatch/pkg/sanitizer"
	"context"
	"fmt"
)

type EmployerService struct {
	employerRepository repository.EmployerRepository
	staticRepository   repository.StaticRepository
}

func NewEmployerService(employerRepository repository.EmployerRepository, staticRepository repository.StaticRepository) usecase.Employer {
	return &EmployerService{
		employerRepository: employerRepository,
		staticRepository:   staticRepository,
	}
}

func (e *EmployerService) employerEntityToDTO(ctx context.Context, employer *entity.Employer) (*dto.EmployerProfileResponse, error) {
	profile := &dto.EmployerProfileResponse{
		ID:           employer.ID,
		CompanyName:  employer.CompanyName,
		LegalAddress: employer.LegalAddress,
		Email:        employer.Email,
		Slogan:       employer.Slogan,
		Website:      employer.Website,
		Description:  employer.Description,
		Vk:           employer.Vk,
		Telegram:     employer.Telegram,
		Facebook:     employer.Facebook,
		CreatedAt:    employer.CreatedAt,
		UpdatedAt:    employer.UpdatedAt,
	}

	if employer.LogoID > 0 {
		logo, err := e.staticRepository.GetStatic(ctx, employer.LogoID)
		if err != nil {
			return nil, err
		}
		profile.LogoPath = logo
	}

	return profile, nil
}

func (e *EmployerService) Register(ctx context.Context, registerDTO *dto.EmployerRegister) (int, error) {
	employer := new(entity.Employer)

	if err := entity.ValidateEmail(registerDTO.Email); err != nil {
		return -1, err
	}

	if err := entity.ValidatePassword(registerDTO.Password); err != nil {
		return -1, err
	}

	if err := entity.ValidateCompanyName(registerDTO.CompanyName); err != nil {
		return -1, err
	}

	if err := entity.ValidateLegalAddress(registerDTO.LegalAddress); err != nil {
		return -1, err
	}

	salt, hash, err := entity.HashPassword(registerDTO.Password)
	if err != nil {
		return -1, err
	}

	sanitizedCompanyName := sanitizer.StrictPolicy.Sanitize(registerDTO.CompanyName)
	sanitizedLegalAddress := sanitizer.StrictPolicy.Sanitize(registerDTO.LegalAddress)
	employer, err = e.employerRepository.CreateEmployer(ctx, registerDTO.Email, sanitizedCompanyName, sanitizedLegalAddress, hash, salt)
	if err != nil {
		return -1, err
	}
	return employer.ID, nil
}

func (e *EmployerService) Login(ctx context.Context, loginDTO *dto.Login) (int, error) {
	if err := entity.ValidateEmail(loginDTO.Email); err != nil {
		return -1, err
	}

	if err := entity.ValidatePassword(loginDTO.Password); err != nil {
		return -1, err
	}

	employer, err := e.employerRepository.GetEmployerByEmail(ctx, loginDTO.Email)
	if err != nil {
		return -1, err
	}
	if entity.CheckPassword(loginDTO.Password, employer.PasswordHash, employer.PasswordSalt) {
		return employer.ID, nil
	}
	return -1, entity.NewError(
		entity.ErrForbidden,
		fmt.Errorf("неверный пароль"),
	)
}

func (e *EmployerService) GetUser(ctx context.Context, employerID int) (*dto.EmployerProfileResponse, error) {
	employer, err := e.employerRepository.GetEmployerByID(ctx, employerID)
	if err != nil {
		return nil, err
	}
	return e.employerEntityToDTO(ctx, employer)
}

func (e *EmployerService) UpdateProfile(ctx context.Context, userID int, employerDTO *dto.EmployerProfileUpdate) error {
	updateFields := make(map[string]interface{})

	if employerDTO.CompanyName != "" {
		if err := entity.ValidateCompanyName(employerDTO.CompanyName); err != nil {
			return err
		}
		updateFields["company_name"] = sanitizer.StrictPolicy.Sanitize(employerDTO.CompanyName)
	}
	if employerDTO.LegalAddress != "" {
		if err := entity.ValidateLegalAddress(employerDTO.LegalAddress); err != nil {
			return err
		}
		updateFields["legal_address"] = sanitizer.StrictPolicy.Sanitize(employerDTO.LegalAddress)
	}
	if employerDTO.Slogan != "" {
		if err := entity.ValidateSlogan(employerDTO.Slogan); err != nil {
			return err
		}
		updateFields["slogan"] = sanitizer.StrictPolicy.Sanitize(employerDTO.Slogan)
	}
	if employerDTO.Website != "" {
		if err := entity.ValidateURL(employerDTO.Website); err != nil {
			return err
		}
		updateFields["website"] = sanitizer.StrictPolicy.Sanitize(employerDTO.Website)
	}
	if employerDTO.Description != "" {
		if err := entity.ValidateDescription(employerDTO.Description); err != nil {
			return err
		}
		updateFields["description"] = sanitizer.StrictPolicy.Sanitize(employerDTO.Description)
	}
	if employerDTO.Vk != "" {
		if err := entity.ValidateURL(employerDTO.Vk); err != nil {
			return err
		}
		updateFields["vk"] = sanitizer.StrictPolicy.Sanitize(employerDTO.Vk)
	}
	if employerDTO.Telegram != "" {
		if err := entity.ValidateURL(employerDTO.Telegram); err != nil {
			return err
		}
		updateFields["telegram"] = sanitizer.StrictPolicy.Sanitize(employerDTO.Telegram)
	}
	if employerDTO.Facebook != "" {
		if err := entity.ValidateURL(employerDTO.Facebook); err != nil {
			return err
		}
		updateFields["facebook"] = sanitizer.StrictPolicy.Sanitize(employerDTO.Facebook)
	}

	if len(updateFields) == 0 {
		return entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("no fields to update"),
		)
	}

	return e.employerRepository.UpdateEmployer(ctx, userID, updateFields)
}

func (e *EmployerService) UpdateLogo(ctx context.Context, userID, logoID int) error {
	err := e.employerRepository.UpdateEmployer(ctx, userID, map[string]interface{}{"logo_id": logoID})
	if err != nil {
		return err
	}
	return nil
}
