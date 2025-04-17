package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type VacanciesService struct {
	vacanciesRepository      repository.VacancyRepository
	applicantRepository      repository.ApplicantRepository
	specializationRepository repository.SpecializationRepository
}

func NewVacanciesService(vacancyRepo repository.VacancyRepository,
	applicantRepo repository.ApplicantRepository,
	specializationRepo repository.SpecializationRepository,
) usecase.Vacancy {
	return &VacanciesService{
		vacanciesRepository:      vacancyRepo,
		applicantRepository:      applicantRepo,
		specializationRepository: specializationRepo,
	}
}

func (s *VacanciesService) CreateVacancy(ctx context.Context, employerID int, request *dto.VacancyCreate) (*dto.VacancyResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"employerID": employerID,
	}).Info("Создание вакансии")

	var specializationID int
	var err error
	if request.Specialization != "" {
		specializationID, err = s.vacanciesRepository.FindSpecializationIDByName(ctx, request.Specialization)
		if err != nil {
			return nil, err
		}
	}

	vacancy := &entity.Vacancy{
		Title:                request.Title,
		IsActive:             true,
		EmployerID:           employerID,
		SpecializationID:     specializationID,
		WorkFormat:           request.WorkFormat,
		Employment:           request.Employment,
		Schedule:             request.Schedule,
		WorkingHours:         request.WorkingHours,
		SalaryFrom:           request.SalaryFrom,
		SalaryTo:             request.SalaryTo,
		TaxesIncluded:        request.TaxesIncluded,
		Experience:           request.Experience,
		Description:          request.Description,
		Tasks:                request.Tasks,
		Requirements:         request.Requirements,
		OptionalRequirements: request.OptionalRequirements,
		City:                 request.City,
	}

	if err := vacancy.Validate(); err != nil {
		return nil, err
	}

	createdVacancy, err := s.vacanciesRepository.Create(ctx, vacancy)
	if err != nil {
		return nil, err
	}

	if len(request.Skills) > 0 {
		skillIDs, err := s.vacanciesRepository.FindSkillIDsByNames(ctx, request.Skills)
		if err != nil {
			return nil, err
		}

		if len(skillIDs) > 0 {
			if err := s.vacanciesRepository.AddSkills(ctx, createdVacancy.ID, skillIDs); err != nil {
				return nil, err
			}
		}
	}
	var specializationName string
	if createdVacancy.SpecializationID != 0 {
		specialization, err := s.specializationRepository.GetByID(ctx, createdVacancy.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}

	skills, err := s.vacanciesRepository.GetSkillsByVacancyID(ctx, createdVacancy.ID)
	if err != nil {
		return nil, err
	}

	experienceStr := fmt.Sprintf("%d+ лет", createdVacancy.Experience)

	response := &dto.VacancyResponse{
		ID:                   createdVacancy.ID,
		EmployerID:           createdVacancy.EmployerID,
		Title:                createdVacancy.Title,
		Specialization:       specializationName,
		WorkFormat:           createdVacancy.WorkFormat,
		Employment:           createdVacancy.Employment,
		Schedule:             createdVacancy.Schedule,
		WorkingHours:         createdVacancy.WorkingHours,
		SalaryFrom:           createdVacancy.SalaryFrom,
		SalaryTo:             createdVacancy.SalaryTo,
		TaxesIncluded:        createdVacancy.TaxesIncluded,
		Experience:           experienceStr,
		City:                 createdVacancy.City,
		Description:          createdVacancy.Description,
		Tasks:                createdVacancy.Tasks,
		Requirements:         createdVacancy.Requirements,
		OptionalRequirements: createdVacancy.OptionalRequirements,
		CreatedAt:            createdVacancy.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            createdVacancy.UpdatedAt.Format(time.RFC3339),
	}

	response.Skills = make([]string, 0, len(skills))
	for _, skill := range skills {
		response.Skills = append(response.Skills, skill)
	}

	return response, nil
}

func (vs *VacanciesService) GetVacancy(ctx context.Context, id, currentUserID int) (*dto.VacancyResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"vacancyID": id,
	}).Info("Получение вакансии по ID")

	vacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// var specializationName string
	// if vacancy.Specialization != "" {
	// 	specialization, err := vs.specializationRepository.GetByID(ctx, vacancy.SpecializationID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	specializationName = specialization.Name
	// }

	skills, err := vs.vacanciesRepository.GetSkillsByVacancyID(ctx, vacancy.ID)
	if err != nil {
		return nil, err
	}

	responded := false
	if currentUserID != 0 {
		responded, err = vs.vacanciesRepository.ResponseExists(ctx, vacancy.ID, currentUserID)
		if err != nil {
			return nil, err
		}
	}

	experienceStr := fmt.Sprintf(vacancy.Experience)

	response := &dto.VacancyResponse{
		ID:                   vacancy.ID,
		EmployerID:           vacancy.EmployerID,
		Title:                vacancy.Title,
		Specialization:       vacancy.Specialization,
		WorkFormat:           vacancy.WorkFormat,
		Employment:           vacancy.Employment,
		Schedule:             vacancy.Schedule,
		WorkingHours:         vacancy.WorkingHours,
		SalaryFrom:           vacancy.SalaryFrom,
		SalaryTo:             vacancy.SalaryTo,
		TaxesIncluded:        vacancy.TaxesIncluded,
		Experience:           experienceStr,
		Description:          vacancy.Description,
		Tasks:                vacancy.Tasks,
		Requirements:         vacancy.Requirements,
		OptionalRequirements: vacancy.OptionalRequirements,
		CreatedAt:            vacancy.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            vacancy.UpdatedAt.Format(time.RFC3339),
		Skills:               skills,
		City:                 vacancy.City,
		Responded:            responded,
	}

	for _, skill := range skills {
		response.Skills = append(response.Skills, skill)
	}

	return response, nil
}

func (vs *VacanciesService) UpdateVacancy(ctx context.Context, id int, request *dto.VacancyUpdate) (*dto.VacancyResponse, error) {
	requestID := utils.GetRequestID(ctx)

	employerID, ok := ctx.Value("employerID").(int)
	if !ok {
		return nil, entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("не удалось получить ID работодателя из контекста"),
		)
	}

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"vacancyID":  id,
		"employerID": employerID,
	}).Info("Обновление вакансии")

	existingVacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingVacancy.EmployerID != employerID {
		return nil, entity.NewError(
			entity.ErrForbidden,
			fmt.Errorf("вакансия с id=%d не принадлежит работодателю с id=%d", id, employerID),
		)
	}

	var specializationID int
	if request.Specialization != "" {
		specializationID, err = vs.vacanciesRepository.FindSpecializationIDByName(ctx, request.Specialization)
		if err != nil {
			return nil, err
		}
	}

	vacancy := &entity.Vacancy{
		ID:                   id,
		EmployerID:           employerID,
		Title:                request.Title,
		SpecializationID:     specializationID,
		WorkFormat:           request.WorkFormat,
		Employment:           request.Employment,
		Schedule:             request.Schedule,
		WorkingHours:         request.WorkingHours,
		SalaryFrom:           request.SalaryFrom,
		SalaryTo:             request.SalaryTo,
		TaxesIncluded:        request.TaxesIncluded,
		Experience:           request.Experience,
		Description:          request.Description,
		Tasks:                request.Tasks,
		Requirements:         request.Requirements,
		OptionalRequirements: request.OptionalRequirements,
		City:                 request.City,
	}

	if err := vacancy.Validate(); err != nil {
		return nil, err
	}

	updatedVacancy, err := vs.vacanciesRepository.Update(ctx, vacancy)
	if err != nil {
		return nil, err
	}

	if err := vs.vacanciesRepository.DeleteSkills(ctx, id); err != nil {
		return nil, err
	}
	if len(request.Skills) > 0 {
		skillIDs, err := vs.vacanciesRepository.FindSkillIDsByNames(ctx, request.Skills)
		if err != nil {
			return nil, err
		}
		if len(skillIDs) > 0 {
			if err := vs.vacanciesRepository.AddSkills(ctx, id, skillIDs); err != nil {
				return nil, err
			}
		}
	}

	var specializationName string
	if updatedVacancy.SpecializationID != 0 {
		specialization, err := vs.specializationRepository.GetByID(ctx, updatedVacancy.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}
	skills, err := vs.vacanciesRepository.GetSkillsByVacancyID(ctx, id)
	if err != nil {
		return nil, err
	}

	experienceStr := fmt.Sprintf("%d+ лет", updatedVacancy.Experience)
	response := &dto.VacancyResponse{
		ID:                   updatedVacancy.ID,
		EmployerID:           updatedVacancy.EmployerID,
		Title:                updatedVacancy.Title,
		Specialization:       specializationName,
		WorkFormat:           updatedVacancy.WorkFormat,
		Employment:           updatedVacancy.Employment,
		Schedule:             updatedVacancy.Schedule,
		WorkingHours:         updatedVacancy.WorkingHours,
		SalaryFrom:           updatedVacancy.SalaryFrom,
		SalaryTo:             updatedVacancy.SalaryTo,
		TaxesIncluded:        updatedVacancy.TaxesIncluded,
		Experience:           experienceStr,
		Description:          updatedVacancy.Description,
		Tasks:                updatedVacancy.Tasks,
		Requirements:         updatedVacancy.Requirements,
		OptionalRequirements: updatedVacancy.OptionalRequirements,
		CreatedAt:            updatedVacancy.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            updatedVacancy.UpdatedAt.Format(time.RFC3339),
		Skills:               make([]string, 0, len(skills)),
		City:                 updatedVacancy.City,
	}

	for _, skill := range skills {
		response.Skills = append(response.Skills, skill)
	}
	return response, nil
}

func (vs *VacanciesService) DeleteVacancy(ctx context.Context, id int, employerID int) (*dto.DeleteVacancy, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"vacancyID":  id,
		"employerID": employerID,
	}).Info("Удаление вакансии")

	existingVacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingVacancy.EmployerID != employerID {
		return nil, entity.NewError(
			entity.ErrForbidden,
			fmt.Errorf("вакансия с id=%d не принадлежит работодателю с id=%d", id, employerID),
		)
	}

	if err := vs.vacanciesRepository.DeleteSkills(ctx, id); err != nil {
		return nil, err
	}

	if err := vs.vacanciesRepository.DeleteCity(ctx, id); err != nil {
		return nil, err
	}

	if err := vs.vacanciesRepository.Delete(ctx, id); err != nil {
		return nil, err
	}

	return &dto.DeleteVacancy{
		Success: true,
		Message: fmt.Sprintf("Вакансия с id=%d успешно удалена", id),
	}, nil
}

func (s *VacanciesService) GetAll(ctx context.Context) ([]dto.VacancyShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("Получение списка всех вакансий")

	vacancies, err := s.vacanciesRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]dto.VacancyShortResponse, 0, len(vacancies))
	for _, vacancy := range vacancies {
		var specializationName string
		if vacancy.SpecializationID != 0 {
			specialization, err := s.specializationRepository.GetByID(ctx, vacancy.SpecializationID)
			if err != nil {
				l.Log.WithFields(logrus.Fields{
					"requestID":        requestID,
					"vacancyID":        vacancy.ID,
					"specializationID": vacancy.SpecializationID,
					"error":            err,
				}).Error("ошибка при получении специализации")
				continue
			}
			specializationName = specialization.Name
		}

		shortVacancy := dto.VacancyShortResponse{
			ID:             vacancy.ID,
			Title:          vacancy.Title,
			EmployerID:     vacancy.EmployerID,
			Specialization: specializationName,
			WorkFormat:     vacancy.WorkFormat,
			Employment:     vacancy.Employment,
			WorkingHours:   vacancy.WorkingHours,
			SalaryFrom:     vacancy.SalaryFrom,
			SalaryTo:       vacancy.SalaryTo,
			CreatedAt:      vacancy.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      vacancy.UpdatedAt.Format(time.RFC3339),
			City:           vacancy.City,
		}

		response = append(response, shortVacancy)
	}

	return response, nil
}
func (s *VacanciesService) ApplyToVacancy(ctx context.Context, vacancyID, applicantID int) error {
	// Проверяем существование вакансии
	if _, err := s.vacanciesRepository.GetByID(ctx, vacancyID); err != nil {
		return fmt.Errorf("vacancy not found: %w", err)
	}

	// Проверяем, не откликался ли уже
	hasResponded, err := s.vacanciesRepository.ResponseExists(ctx, vacancyID, applicantID)
	if err != nil {
		return fmt.Errorf("failed to check existing responses: %w", err)
	}
	if hasResponded {
		return entity.NewError(entity.ErrAlreadyExists,
			fmt.Errorf("you have already applied to this vacancy"))
	}

	return s.vacanciesRepository.CreateResponse(ctx, vacancyID, applicantID)
}
