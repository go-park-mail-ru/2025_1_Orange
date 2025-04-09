package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"context"
	"fmt"
)

type ApplicantService struct {
	applicantRepository repository.ApplicantRepository
	cityRepository      repository.CityRepository
	staticRepository    repository.StaticRepository
}

func NewApplicantService(
	applicantRepository repository.ApplicantRepository,
	cityRepository repository.CityRepository,
	staticRepository repository.StaticRepository,
) usecase.Applicant {
	return &ApplicantService{
		applicantRepository: applicantRepository,
		cityRepository:      cityRepository,
		staticRepository:    staticRepository,
	}
}

func (a *ApplicantService) applicantEntityToDTO(ctx context.Context, applicantEntity *entity.Applicant) (*dto.ApplicantProfileResponse, error) {
	profile := &dto.ApplicantProfileResponse{
		ID:         applicantEntity.ID,
		FirstName:  applicantEntity.FirstName,
		LastName:   applicantEntity.LastName,
		MiddleName: applicantEntity.MiddleName,
		Email:      applicantEntity.Email,
		BirthDate:  applicantEntity.BirthDate,
		Sex:        applicantEntity.Sex,
		Status:     string(applicantEntity.Status),
		Quote:      applicantEntity.Quote,
		CreatedAt:  applicantEntity.CreatedAt,
		UpdatedAt:  applicantEntity.UpdatedAt,
	}

	if applicantEntity.AvatarID > 0 {
		avatar, err := a.staticRepository.GetStatic(ctx, applicantEntity.AvatarID)
		if err != nil {
			return nil, err
		}
		profile.AvatarPath = avatar
	}

	if applicantEntity.CityID > 0 {
		city, err := a.cityRepository.GetByID(ctx, applicantEntity.CityID)
		if err != nil {
			return nil, err
		}
		profile.City = city.Name
	}

	return profile, nil
}

func (a *ApplicantService) Register(ctx context.Context, registerDTO *dto.ApplicantRegister) (int, error) {

	if err := entity.ValidateEmail(registerDTO.Email); err != nil {
		return -1, err
	}

	if err := entity.ValidatePassword(registerDTO.Password); err != nil {
		return -1, err
	}

	if err := entity.ValidateFirstName(registerDTO.FirstName); err != nil {
		return -1, err
	}

	if err := entity.ValidateLastName(registerDTO.LastName); err != nil {
		return -1, err
	}

	salt, hash, err := entity.HashPassword(registerDTO.Password)
	if err != nil {
		return -1, err
	}

	applicant, err := a.applicantRepository.CreateApplicant(ctx, registerDTO.Email, registerDTO.FirstName, registerDTO.LastName, hash, salt)
	if err != nil {
		return -1, err
	}
	return applicant.ID, nil
}

func (a *ApplicantService) Login(ctx context.Context, loginDTO *dto.ApplicantLogin) (int, error) {
	if err := entity.ValidateEmail(loginDTO.Email); err != nil {
		return -1, err
	}

	if err := entity.ValidatePassword(loginDTO.Password); err != nil {
		return -1, err
	}

	applicant, err := a.applicantRepository.GetApplicantByEmail(ctx, loginDTO.Email)
	if err != nil {
		return -1, err
	}

	if entity.CheckPassword(loginDTO.Password, applicant.PasswordHash, applicant.PasswordSalt) {
		return applicant.ID, nil
	}
	return -1, entity.NewError(
		entity.ErrForbidden,
		fmt.Errorf("неверный пароль"),
	)
}

func (a *ApplicantService) GetUser(ctx context.Context, applicantID int) (*dto.ApplicantProfileResponse, error) {
	applicant, err := a.applicantRepository.GetApplicantByID(ctx, applicantID)
	if err != nil {
		return nil, err
	}
	return a.applicantEntityToDTO(ctx, applicant)
}

func (a *ApplicantService) UpdateProfile(ctx context.Context, userID int, applicantDTO *dto.ApplicantProfileUpdate) error {
	updateFields := make(map[string]interface{})

	if applicantDTO.FirstName != "" {
		if err := entity.ValidateFirstName(applicantDTO.FirstName); err != nil {
			return err
		}
		updateFields["first_name"] = applicantDTO.FirstName
	}
	if applicantDTO.LastName != "" {
		if err := entity.ValidateLastName(applicantDTO.LastName); err != nil {
			return err
		}
		updateFields["last_name"] = applicantDTO.LastName
	}
	if applicantDTO.MiddleName != "" {
		if err := entity.ValidateMiddleName(applicantDTO.MiddleName); err != nil {
			return err
		}
		updateFields["middle_name"] = applicantDTO.MiddleName
	}
	if !applicantDTO.BirthDate.IsZero() {
		if err := entity.ValidateBirthDate(applicantDTO.BirthDate); err != nil {
			return err
		}
		updateFields["birth_date"] = applicantDTO.BirthDate
	}
	if applicantDTO.Sex != "" {
		if err := entity.ValidateSex(applicantDTO.Sex); err != nil {
			return err
		}
		updateFields["sex"] = applicantDTO.Sex
	}
	if applicantDTO.Status != "" {
		if err := entity.ValidateStatus(applicantDTO.Status); err != nil {
			return err
		}
		updateFields["status"] = applicantDTO.Status
	}
	if applicantDTO.Quote != "" {
		if err := entity.ValidateQuote(applicantDTO.Quote); err != nil {
			return err
		}
		updateFields["quote"] = applicantDTO.Quote
	}
	if applicantDTO.City != "" {
		city, err := a.cityRepository.GetByName(ctx, applicantDTO.City)
		if err != nil {
			return err
		}
		updateFields["city_id"] = city.Name
	}

	if len(updateFields) == 0 {
		return entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("no fields to update"),
		)
	}

	return a.applicantRepository.UpdateApplicant(ctx, userID, updateFields)
}

func (a *ApplicantService) UpdateAvatar(ctx context.Context, userID, avatarID int) error {
	err := a.applicantRepository.UpdateApplicant(ctx, userID, map[string]interface{}{"avatar_id": avatarID})
	if err != nil {
		return err
	}
	return nil
}
