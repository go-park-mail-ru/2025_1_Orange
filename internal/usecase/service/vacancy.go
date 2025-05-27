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
	resumeRepository         repository.ResumeRepository
	applicantService         usecase.Applicant
}

func NewVacanciesService(vacancyRepo repository.VacancyRepository,
	applicantRepo repository.ApplicantRepository,
	specializationRepo repository.SpecializationRepository,
	employerService usecase.Employer,
	resumeRepository repository.ResumeRepository,
	applicantService usecase.Applicant,
) usecase.Vacancy {
	return &VacanciesService{
		vacanciesRepository:      vacancyRepo,
		applicantRepository:      applicantRepo,
		specializationRepository: specializationRepo,
		employerService:          employerService,
		resumeRepository:         resumeRepository,
		applicantService:         applicantService,
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
		response.Skills = append(response.Skills, skill.Name)
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
	if vacancy.SpecializationID != 0 {
		specialization, err := vs.specializationRepository.GetByID(ctx, vacancy.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}

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

	liked := false
	if userRole == "applicant" && currentUserID != 0 {
		liked, err = vs.vacanciesRepository.LikeExists(ctx, vacancy.ID, currentUserID)
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
		Skills:               make([]string, 0, len(skills)),
		City:                 vacancy.City,
		Responded:            responded,
		Liked:                liked,
	}

	for _, skill := range skills {
		response.Skills = append(response.Skills, skill.Name)
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
		response.Skills = append(response.Skills, skill.Name)
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

		liked := false
		if userRole == "applicant" && currentUserID != 0 {
			liked, err = s.vacanciesRepository.LikeExists(ctx, vacancy.ID, currentUserID)
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
			Liked:          liked,
		}

		response = append(response, shortVacancy)
	}

	return response, nil
}

func (vs *VacanciesService) ApplyToVacancy(ctx context.Context, vacancyID, applicantID, resumeID int) (entity.Notification, error) {
	notification := entity.Notification{}
	vacancy, err := vs.vacanciesRepository.GetByID(ctx, vacancyID)
	if err != nil {
		return notification, fmt.Errorf("vacancy not found: %w", err)
	}

	hasResponded, err := vs.vacanciesRepository.ResponseExists(ctx, vacancyID, applicantID)
	if err != nil {
		return notification, fmt.Errorf("failed to check existing responses: %w", err)
	}
	if hasResponded {
		return notification, vs.vacanciesRepository.DeleteResponse(ctx, vacancyID, applicantID, resumeID)
	}

	notification = entity.Notification{
		Type:         entity.ApplyNotificationType,
		SenderID:     applicantID,
		SenderRole:   entity.ApplicantRole,
		ReceiverID:   vacancy.EmployerID,
		ReceiverRole: entity.EmployerRole,
		ObjectID:     vacancy.ID,
		ResumeID:     resumeID,
	}

	return notification, vs.vacanciesRepository.CreateResponse(ctx, vacancyID, applicantID, resumeID)
}

func (vs *VacanciesService) GetRespondedResumeOnVacancy(ctx context.Context, vacancyID int, limit, offset int) ([]dto.ResumeApplicantShortResponse, error) {

	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"vacancyID": vacancyID,
	}).Info("Получение списка резюме откликнувшихся на вакансию")

	responses, err := vs.vacanciesRepository.GetVacancyResponses(ctx, vacancyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get vacancy responses: %w", err)
	}

	response := make([]dto.ResumeApplicantShortResponse, 0, len(responses))

	for _, r := range responses {
		resume, err := vs.resumeRepository.GetByID(ctx, r.ResumeID)
		if err != nil {
			return nil, err
		}
		var specializationName string
		if resume.SpecializationID != 0 {
			specialization, err := vs.specializationRepository.GetByID(ctx, resume.SpecializationID)
			if err != nil {
				l.Log.WithFields(logrus.Fields{
					"requestID":        requestID,
					"resumeID":         resume.ID,
					"specializationID": resume.SpecializationID,
					"error":            err,
				}).Error("ошибка при получении специализации")
				continue
			}
			specializationName = specialization.Name
		}

		workExperiences, err := vs.resumeRepository.GetWorkExperienceByResumeID(ctx, resume.ID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"resumeID":  resume.ID,
				"error":     err,
			}).Error("ошибка при получении опыта работы")
			continue
		}

		applicantDTO, err := vs.applicantService.GetUser(ctx, resume.ApplicantID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":   requestID,
				"resumeID":    resume.ID,
				"applicantID": resume.ApplicantID,
				"error":       err,
			}).Error("ошибка при конвертации соискателя в DTO")
			continue
		}

		skills, err := vs.resumeRepository.GetSkillsByResumeID(ctx, resume.ID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"resumeID":  resume.ID,
				"error":     err,
			}).Error("ошибка при получении навыков резюме")
			continue
		}

		// Преобразуем навыки в массив строк
		skillNames := make([]string, 0, len(skills))
		for _, skill := range skills {
			skillNames = append(skillNames, skill.Name)
		}

		shortResume := dto.ResumeApplicantShortResponse{
			ID:             resume.ID,
			Applicant:      applicantDTO,
			Skills:         skillNames,
			Specialization: specializationName,
			Profession:     resume.Profession,
			CreatedAt:      resume.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      resume.UpdatedAt.Format(time.RFC3339),
		}

		if len(workExperiences) > 0 {
			we := workExperiences[0]
			workExp := dto.WorkExperienceShort{
				ID:           we.ID,
				EmployerName: we.EmployerName,
				Position:     we.Position,
				Duties:       we.Duties,
				Achievements: we.Achievements,
				StartDate:    we.StartDate.Format("2006-01-02"),
				UntilNow:     we.UntilNow,
			}

			if !we.UntilNow && !we.EndDate.IsZero() {
				workExp.EndDate = we.EndDate.Format("2006-01-02")
			}

			shortResume.WorkExperience = workExp
		}

		response = append(response, shortResume)
	}
	return response, nil
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

		liked := false
		if userRole == "applicant" && userID != 0 {
			liked, err = vs.vacanciesRepository.LikeExists(ctx, vacancy.ID, userID)
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
			Liked:          liked,
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

		liked := false
		if applicantID != 0 {
			liked, err = vs.vacanciesRepository.LikeExists(ctx, vacancy.ID, applicantID)
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
			Liked:          liked,
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

	// Для соискателя или неавторизованного пользователя ищем все вакансии
	vacancies, err = s.vacanciesRepository.SearchVacancies(ctx, searchQuery, limit, offset)

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

		liked := false
		if userRole == "applicant" && userID != 0 {
			liked, err = s.vacanciesRepository.LikeExists(ctx, vacancy.ID, userID)
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
			Liked:          liked,
		}

		response = append(response, shortVacancy)
	}

	return response, nil
}

// SearchVacanciesBySpecializations ищет вакансии по списку специализаций
func (s *VacanciesService) SearchVacanciesBySpecializations(ctx context.Context, userID int, userRole string, specializations []string, limit int, offset int) ([]dto.VacancyShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":       requestID,
		"userID":          userID,
		"role":            userRole,
		"specializations": specializations,
		"limit":           limit,
		"offset":          offset,
	}).Info("Поиск вакансий по специализациям")

	// Находим ID специализаций по их названиям
	specializationIDs, err := s.vacanciesRepository.FindSpecializationIDsByNames(ctx, specializations)
	if err != nil {
		return nil, err
	}

	// Если не найдено ни одной специализации, возвращаем пустой список
	if len(specializationIDs) == 0 {
		return []dto.VacancyShortResponse{}, nil
	}

	// Ищем вакансии по ID специализаций
	vacancies, err := s.vacanciesRepository.SearchVacanciesBySpecializations(ctx, specializationIDs, limit, offset)
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

		liked := false
		if userRole == "applicant" && userID != 0 {
			liked, err = s.vacanciesRepository.LikeExists(ctx, vacancy.ID, userID)
			if err != nil {
				return nil, err
			}
		}

		// Получаем информацию о работодателе
		employerDTO, err := s.employerService.GetUser(ctx, vacancy.EmployerID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":  requestID,
				"vacancyID":  vacancy.ID,
				"employerID": vacancy.EmployerID,
				"error":      err,
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
			Liked:          liked,
		}

		response = append(response, shortVacancy)
	}

	return response, nil
}

// SearchVacanciesByQueryAndSpecializations ищет вакансии по текстовому запросу и списку специализаций
func (s *VacanciesService) SearchVacanciesByQueryAndSpecializations(ctx context.Context, userID int, userRole string, searchQuery string, specializations []string, minSalary int, employment, experience []string, limit int, offset int) ([]dto.VacancyShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":       requestID,
		"userID":          userID,
		"role":            userRole,
		"query":           searchQuery,
		"specializations": specializations,
		"minSalary":       minSalary,
		"employment":      employment,
		"experience":      experience,
		"limit":           limit,
		"offset":          offset,
	}).Info("Комбинированный поиск вакансий по запросу и специализациям")

	// Валидация входных параметров
	validEmployment := map[string]bool{
		"full_time":  true,
		"part_time":  true,
		"contract":   true,
		"internship": true,
		"freelance":  true,
		"watch":      true,
	}
	validExperience := map[string]bool{
		"no_matter":     true,
		"no_experience": true,
		"1_3_years":     true,
		"3_6_years":     true,
		"6_plus_years":  true,
	}

	for _, emp := range employment {
		if !validEmployment[emp] {
			return nil, entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("некорректное значение employment: %s", emp),
			)
		}
	}

	for _, exp := range experience {
		if !validExperience[exp] {
			return nil, entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("некорректное значение experience: %s", exp),
			)
		}
	}

	if minSalary < 0 {
		return nil, entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("минимальная зарплата не может быть отрицательной"),
		)
	}

	var specializationIDs []int
	var err error

	if len(specializations) > 0 {
		specializationIDs, err = s.vacanciesRepository.FindSpecializationIDsByNames(ctx, specializations)
		if err != nil {
			return nil, err
		}
	}

	// // Если не найдено ни одной специализации, возвращаем пустой список
	// if len(specializationIDs) == 0 {
	// 	return []dto.VacancyShortResponse{}, nil
	// }

	// Ищем вакансии по текстовому запросу и ID специализаций
	vacancies, err := s.vacanciesRepository.SearchVacanciesByQueryAndSpecializations(ctx, searchQuery, specializationIDs, minSalary, employment, experience, limit, offset)
	if err != nil {
		return nil, err
	}

	// Формируем ответ, аналогично другим методам поиска
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

		liked := false
		if userRole == "applicant" && userID != 0 {
			liked, err = s.vacanciesRepository.LikeExists(ctx, vacancy.ID, userID)
			if err != nil {
				return nil, err
			}
		}

		// Получаем информацию о работодателе
		employerDTO, err := s.employerService.GetUser(ctx, vacancy.EmployerID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":  requestID,
				"vacancyID":  vacancy.ID,
				"employerID": vacancy.EmployerID,
				"error":      err,
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
			Liked:          liked,
		}

		response = append(response, shortVacancy)
	}

	return response, nil
}

func (vs *VacanciesService) LikeVacancy(ctx context.Context, vacancyID, applicantID int) error {
	// Проверяем существование вакансии
	if _, err := vs.vacanciesRepository.GetByID(ctx, vacancyID); err != nil {
		return fmt.Errorf("vacancy not found: %w", err)
	}

	hasLiked, err := vs.vacanciesRepository.LikeExists(ctx, vacancyID, applicantID)
	if err != nil {
		return fmt.Errorf("failed to check existing like: %w", err)
	}
	if hasLiked {
		return vs.vacanciesRepository.DeleteLike(ctx, vacancyID, applicantID)
	}

	return vs.vacanciesRepository.CreateLike(ctx, vacancyID, applicantID)
}
func (vs *VacanciesService) GetLikedVacancies(ctx context.Context, applicantID int, limit, offset int) ([]dto.VacancyShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":   requestID,
		"applicantID": applicantID,
	}).Info("Получение понравившихся вакансии по ID соискателя")

	vacancies, err := vs.vacanciesRepository.GetlikedVacancies(ctx, applicantID, limit, offset)
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

		employerDTO, err := vs.employerService.GetUser(ctx, vacancy.EmployerID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":  requestID,
				"resumeID":   vacancy.ID,
				"employerID": vacancy.EmployerID,
				"error":      err,
			}).Error("ошибка при конвертации работодателя в DTO")
			continue
		}

		responded := false
		if applicantID != 0 {
			responded, err = vs.vacanciesRepository.ResponseExists(ctx, vacancy.ID, applicantID)
			if err != nil {
				return nil, err
			}
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
			Liked:          true,
		}

		response = append(response, shortVacancy)
	}

	return response, nil
}
