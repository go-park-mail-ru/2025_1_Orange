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
	employerService          usecase.Employer
}

func NewVacanciesService(vacancyRepo repository.VacancyRepository,
	applicantRepo repository.ApplicantRepository,
	specializationRepo repository.SpecializationRepository,
	employerService usecase.Employer,
) usecase.Vacancy {
	return &VacanciesService{
		vacanciesRepository:      vacancyRepo,
		applicantRepository:      applicantRepo,
		specializationRepository: specializationRepo,
		employerService:          employerService,
	}
}

func (vs *VacanciesService) CreateVacancy(ctx context.Context, employerID int, request *dto.VacancyCreate) (*dto.VacancyResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"employerID": employerID,
	}).Info("Создание вакансии")

	var specializationID int
	var err error
	if request.Specialization != "" {
		specializationID, err = vs.vacanciesRepository.FindSpecializationIDByName(ctx, request.Specialization)
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

	createdVacancy, err := vs.vacanciesRepository.Create(ctx, vacancy)
	if err != nil {
		return nil, err
	}

	if len(request.Skills) > 0 {
		skillIDs, err := vs.vacanciesRepository.FindSkillIDsByNames(ctx, request.Skills)
		if err != nil {
			return nil, err
		}

		if len(skillIDs) > 0 {
			if err := vs.vacanciesRepository.AddSkills(ctx, createdVacancy.ID, skillIDs); err != nil {
				return nil, err
			}
		}
	}
	var specializationName string
	if createdVacancy.SpecializationID != 0 {
		specialization, err := vs.specializationRepository.GetByID(ctx, createdVacancy.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}

	skills, err := vs.vacanciesRepository.GetSkillsByVacancyID(ctx, createdVacancy.ID)
	if err != nil {
		return nil, err
	}

	experienceStr := fmt.Sprintf(createdVacancy.Experience)

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
		Skills:               make([]string, 0, len(skills)),
		OptionalRequirements: createdVacancy.OptionalRequirements,
		CreatedAt:            createdVacancy.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            createdVacancy.UpdatedAt.Format(time.RFC3339),
	}

	for _, skill := range skills {
		response.Skills = append(response.Skills, skill)
	}

	return response, nil
}

func (vs *VacanciesService) GetVacancy(ctx context.Context, id, currentUserID int, userRole string) (*dto.VacancyResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"vacancyID": id,
	}).Info("Получение вакансии по ID")

	vacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var specializationName string
	if vacancy.SpecializationID != 0 { // 0 или другое значение по умолчанию
		specialization, err := vs.specializationRepository.GetByID(ctx, vacancy.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
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
	if userRole == "applicant" && currentUserID != 0 {
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
		Specialization:       specializationName,
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

func (vs *VacanciesService) UpdateVacancy(ctx context.Context, id int, employerID int, request *dto.VacancyUpdate) (*dto.VacancyResponse, error) {
	requestID := utils.GetRequestID(ctx)

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

	experienceStr := fmt.Sprintf(updatedVacancy.Experience)
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

func (s *VacanciesService) GetAll(ctx context.Context, currentUserID int, userRole string, limit int, offset int) ([]dto.VacancyShortResponse, error) {

	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("Получение списка всех вакансий")

	vacancies, err := s.vacanciesRepository.GetAll(ctx, limit, offset)

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

		responded := false
		if userRole == "applicant" && currentUserID != 0 {
			responded, err = s.vacanciesRepository.ResponseExists(ctx, vacancy.ID, currentUserID)
			if err != nil {
				return nil, err
			}
		}

		employerDTO, err := s.employerService.GetUser(ctx, vacancy.EmployerID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":   requestID,
				"resumeID":    vacancy.ID,
				"applicantID": vacancy.EmployerID,
				"error":       err,
			}).Error("ошибка при конвертации работодателя в DTO")
			continue
		}

		shortVacancy := dto.VacancyShortResponse{
			ID:             vacancy.ID,
			Title:          vacancy.Title,
			Employer:       employerDTO,
			Specialization: specializationName,
			WorkFormat:     vacancy.WorkFormat,
			Employment:     vacancy.Employment,
			WorkingHours:   vacancy.WorkingHours,
			SalaryFrom:     vacancy.SalaryFrom,
			SalaryTo:       vacancy.SalaryTo,
			TaxesIncluded:  vacancy.TaxesIncluded,
			CreatedAt:      vacancy.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      vacancy.UpdatedAt.Format(time.RFC3339),
			City:           vacancy.City,
			Responded:      responded,
		}

		response = append(response, shortVacancy)
	}

	return response, nil
}
func (vs *VacanciesService) ApplyToVacancy(ctx context.Context, vacancyID, applicantID int) error {
	// Проверяем существование вакансии
	if _, err := vs.vacanciesRepository.GetByID(ctx, vacancyID); err != nil {
		return fmt.Errorf("vacancy not found: %w", err)
	}

	// Проверяем, не откликался ли уже
	hasResponded, err := vs.vacanciesRepository.ResponseExists(ctx, vacancyID, applicantID)
	if err != nil {
		return fmt.Errorf("failed to check existing responses: %w", err)
	}
	if hasResponded {
		return entity.NewError(entity.ErrAlreadyExists,
			fmt.Errorf("you have already applied to this vacancy"))
	}

	return vs.vacanciesRepository.CreateResponse(ctx, vacancyID, applicantID)
}

func (vs *VacanciesService) GetActiveVacanciesByEmployerID(ctx context.Context, employerID, userID int, userRole string, limit int, offset int) ([]dto.VacancyShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"employerID": employerID,
	}).Info("Получение вакансии по ID работодателя")

	vacancies, err := vs.vacanciesRepository.GetActiveVacanciesByEmployerID(ctx, employerID, limit, offset)
	if err != nil {
		return nil, err
	}

	response := make([]dto.VacancyShortResponse, 0, len(vacancies))
	for _, vacancy := range vacancies {
		var specializationName string
		if vacancy.SpecializationID != 0 {
			specialization, err := vs.specializationRepository.GetByID(ctx, vacancy.SpecializationID)
			if err != nil {
				l.Log.WithFields(logrus.Fields{
					"requestID":  requestID,
					"vacancyID":  vacancy.ID,
					"employerID": employerID,
					"error":      err,
				}).Error("ошибка при получении специализации")
				continue
			}
			specializationName = specialization.Name
		}

		responded := false
		if userRole == "applicant" && userID != 0 {
			responded, err = vs.vacanciesRepository.ResponseExists(ctx, vacancy.ID, userID)
			if err != nil {
				return nil, err
			}
		}

		employerDTO, err := vs.employerService.GetUser(ctx, vacancy.EmployerID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":   requestID,
				"resumeID":    vacancy.ID,
				"applicantID": vacancy.EmployerID,
				"error":       err,
			}).Error("ошибка при конвертации работодателя в DTO")
			continue
		}

		shortVacancy := dto.VacancyShortResponse{
			ID:             vacancy.ID,
			Title:          vacancy.Title,
			Employer:       employerDTO,
			Specialization: specializationName,
			WorkFormat:     vacancy.WorkFormat,
			Employment:     vacancy.Employment,
			WorkingHours:   vacancy.WorkingHours,
			SalaryFrom:     vacancy.SalaryFrom,
			SalaryTo:       vacancy.SalaryTo,
			TaxesIncluded:  vacancy.TaxesIncluded,
			CreatedAt:      vacancy.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      vacancy.UpdatedAt.Format(time.RFC3339),
			City:           vacancy.City,
			Responded:      responded,
		}

		response = append(response, shortVacancy)
	}

	return response, nil
}

func (vs *VacanciesService) GetVacanciesByApplicantID(ctx context.Context, applicantID int, limit int, offset int) ([]dto.VacancyShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":   requestID,
		"applicantID": applicantID,
	}).Info("Получение вакансии по ID работодателя")

	vacancies, err := vs.vacanciesRepository.GetVacanciesByApplicantID(ctx, applicantID, limit, offset)
	if err != nil {
		return nil, err
	}

	response := make([]dto.VacancyShortResponse, 0, len(vacancies))
	for _, vacancy := range vacancies {
		var specializationName string
		if vacancy.SpecializationID != 0 {
			specialization, err := vs.specializationRepository.GetByID(ctx, vacancy.SpecializationID)
			if err != nil {
				l.Log.WithFields(logrus.Fields{
					"requestID":   requestID,
					"vacancyID":   vacancy.ID,
					"applicantID": applicantID,
					"error":       err,
				}).Error("ошибка при получении специализации")
				continue
			}
			specializationName = specialization.Name
		}

		responded := false
		if applicantID != 0 {
			responded, err = vs.vacanciesRepository.ResponseExists(ctx, vacancy.ID, applicantID)
			if err != nil {
				return nil, err
			}
		}

		employerDTO, err := vs.employerService.GetUser(ctx, vacancy.EmployerID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":   requestID,
				"resumeID":    vacancy.ID,
				"applicantID": vacancy.EmployerID,
				"error":       err,
			}).Error("ошибка при конвертации работодателя в DTO")
			continue
		}

		shortVacancy := dto.VacancyShortResponse{
			ID:             vacancy.ID,
			Title:          vacancy.Title,
			Employer:       employerDTO,
			Specialization: specializationName,
			WorkFormat:     vacancy.WorkFormat,
			Employment:     vacancy.Employment,
			WorkingHours:   vacancy.WorkingHours,
			SalaryFrom:     vacancy.SalaryFrom,
			SalaryTo:       vacancy.SalaryTo,
			TaxesIncluded:  vacancy.TaxesIncluded,
			CreatedAt:      vacancy.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      vacancy.UpdatedAt.Format(time.RFC3339),
			City:           vacancy.City,
			Responded:      responded,
		}

		response = append(response, shortVacancy)
	}

	return response, nil
}

// SearchVacancies ищет вакансии по заданному запросу с учетом роли пользователя
func (s *VacanciesService) SearchVacancies(ctx context.Context, userID int, userRole string, searchQuery string, limit int, offset int) ([]dto.VacancyShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"userID":    userID,
		"role":      userRole,
		"query":     searchQuery,
	}).Info("Поиск вакансий")

	var vacancies []*entity.Vacancy
	var err error

	// В зависимости от роли пользователя выбираем метод поиска
	if userRole == "employer" {
		// Для работодателя ищем только его вакансии
		vacancies, err = s.vacanciesRepository.SearchVacanciesByEmployerID(ctx, userID, searchQuery, limit, offset)
	} else {
		// Для соискателя или неавторизованного пользователя ищем все вакансии
		vacancies, err = s.vacanciesRepository.SearchVacancies(ctx, searchQuery, limit, offset)
	}

	if err != nil {
		return nil, err
	}

	// Формируем ответ, аналогично методу GetAll
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

		responded := false
		if userRole == "applicant" && userID != 0 {
			responded, err = s.vacanciesRepository.ResponseExists(ctx, vacancy.ID, userID)
			if err != nil {
				return nil, err
			}
		}

		// Получаем информацию о соискателе
		employerDTO, err := s.employerService.GetUser(ctx, vacancy.EmployerID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":   requestID,
				"resumeID":    vacancy.ID,
				"applicantID": vacancy.EmployerID,
				"error":       err,
			}).Error("ошибка при конвертации работодателя в DTO")
			continue
		}

		shortVacancy := dto.VacancyShortResponse{
			ID:             vacancy.ID,
			Title:          vacancy.Title,
			Employer:       employerDTO,
			Specialization: specializationName,
			WorkFormat:     vacancy.WorkFormat,
			Employment:     vacancy.Employment,
			WorkingHours:   vacancy.WorkingHours,
			SalaryFrom:     vacancy.SalaryFrom,
			SalaryTo:       vacancy.SalaryTo,
			TaxesIncluded:  vacancy.TaxesIncluded,
			CreatedAt:      vacancy.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      vacancy.UpdatedAt.Format(time.RFC3339),
			City:           vacancy.City,
			Responded:      responded,
		}

		response = append(response, shortVacancy)
	}

	return response, nil
}
