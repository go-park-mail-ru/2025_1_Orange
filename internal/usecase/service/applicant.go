package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

type ApplicantService struct {
	applicantRepository repository.ApplicantRepository
	cityRepository      repository.CityRepository
}

func NewApplicantService(
	applicantRepository repository.ApplicantRepository,
	cityRepository repository.CityRepository,
) usecase.Applicant {
	return &ApplicantService{
		applicantRepository: applicantRepository,
		cityRepository:      cityRepository,
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
		AvatarPath: applicantEntity.AvatarPath,
		CreatedAt:  applicantEntity.CreatedAt,
		UpdatedAt:  applicantEntity.UpdatedAt,
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

func (a *ApplicantService) UploadAvatar(ctx context.Context, userID int, fileHeader *multipart.FileHeader) (*dto.UploadAvatarResponse, error) {
	if err := entity.ValidateAvatar(fileHeader); err != nil {
		return nil, err
	}

	fileExt := strings.ToLower(filepath.Ext(fileHeader.Filename))
	relativePath := fmt.Sprintf("/img/applicant/%d%s", userID, fileExt)
	fullPath := filepath.Join("assets", relativePath)

	file, err := fileHeader.Open()
	if err != nil {
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось открыть файл: %w", err),
		)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	if err := os.WriteFile(fullPath, fileBytes, 0644); err != nil {
		return nil, fmt.Errorf("не удалось сохранить аватар: %w", err)
	}

	updateFields := map[string]interface{}{
		"avatar_path": relativePath,
	}

	if err := a.applicantRepository.UpdateApplicant(ctx, userID, updateFields); err != nil {
		os.Remove(fullPath)
		return nil, fmt.Errorf("не удалось обновить путь до аватара: %w", err)
	}
	return &dto.UploadAvatarResponse{AvatarPath: relativePath}, nil
}
