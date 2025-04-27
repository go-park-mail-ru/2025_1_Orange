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

// Add the required dependencies to the ResumeService struct
type ResumeService struct {
	resumeRepository         repository.ResumeRepository
	skillRepository          repository.SkillRepository
	specializationRepository repository.SpecializationRepository
	applicantRepository      repository.ApplicantRepository
	applicantService         usecase.Applicant
}

// Update the constructor to include the new dependencies
func NewResumeService(
	resumeRepo repository.ResumeRepository,
	skillRepo repository.SkillRepository,
	specializationRepo repository.SpecializationRepository,
	applicantRepo repository.ApplicantRepository,
	applicantService usecase.Applicant,
) usecase.ResumeUsecase {
	return &ResumeService{
		resumeRepository:         resumeRepo,
		skillRepository:          skillRepo,
		specializationRepository: specializationRepo,
		applicantRepository:      applicantRepo,
		applicantService:         applicantService,
	}
}

func (s *ResumeService) Create(ctx context.Context, applicantID int, request *dto.CreateResumeRequest) (*dto.ResumeResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":   requestID,
		"applicantID": applicantID,
	}).Info("Создание резюме")

	// Find specialization ID by name
	var specializationID int
	var err error
	if request.Specialization != "" {
		specializationID, err = s.resumeRepository.FindSpecializationIDByName(ctx, request.Specialization)
		if err != nil {
			return nil, err
		}
	}

	// Parse graduation year if provided
	var graduationYear time.Time
	if request.GraduationYear != "" {
		graduationYear, err = time.Parse("2006-01-02", request.GraduationYear)
		if err != nil {
			return nil, entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты окончания учебы: %w", err),
			)
		}
	}

	// Create resume entity
	resume := &entity.Resume{
		ApplicantID:            applicantID,
		AboutMe:                request.AboutMe,
		SpecializationID:       specializationID,
		Education:              request.Education,
		EducationalInstitution: request.EducationalInstitution,
		GraduationYear:         graduationYear,
		Profession:             request.Profession, // Дополнение - добавлено поле профессии
	}

	// Validate resume
	if err := resume.Validate(); err != nil {
		return nil, err
	}

	// Create resume in DB
	createdResume, err := s.resumeRepository.Create(ctx, resume)
	if err != nil {
		return nil, err
	}

	// Find skill IDs by names and add them
	if len(request.Skills) > 0 {
		skillIDs, err := s.resumeRepository.FindSkillIDsByNames(ctx, request.Skills)
		if err != nil {
			return nil, err
		}

		if len(skillIDs) > 0 {
			if err := s.resumeRepository.AddSkills(ctx, createdResume.ID, skillIDs); err != nil {
				return nil, err
			}
		}
	}

	// Find additional specialization IDs by names and add them
	if len(request.AdditionalSpecializations) > 0 {
		specIDs, err := s.resumeRepository.FindSpecializationIDsByNames(ctx, request.AdditionalSpecializations)
		if err != nil {
			return nil, err
		}

		if len(specIDs) > 0 {
			if err := s.resumeRepository.AddSpecializations(ctx, createdResume.ID, specIDs); err != nil {
				return nil, err
			}
		}
	}

	// Add work experiences if provided
	var workExperiences []entity.WorkExperience
	for _, we := range request.WorkExperiences {
		startDate, err := time.Parse("2006-01-02", we.StartDate)
		if err != nil {
			return nil, entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты начала работы: %w", err),
			)
		}

		var endDate time.Time
		if !we.UntilNow && we.EndDate != "" {
			endDate, err = time.Parse("2006-01-02", we.EndDate)
			if err != nil {
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неверный формат даты окончания работы: %w", err),
				)
			}
		}

		workExperience := entity.WorkExperience{
			ResumeID:     createdResume.ID,
			EmployerName: we.EmployerName,
			Position:     we.Position,
			Duties:       we.Duties,
			Achievements: we.Achievements,
			StartDate:    startDate,
			EndDate:      endDate,
			UntilNow:     we.UntilNow,
		}

		if err := workExperience.Validate(); err != nil {
			return nil, err
		}

		createdWorkExperience, err := s.resumeRepository.AddWorkExperience(ctx, &workExperience)
		if err != nil {
			return nil, err
		}

		workExperiences = append(workExperiences, *createdWorkExperience)
	}

	// Get specialization name
	var specializationName string
	if createdResume.SpecializationID != 0 {
		specialization, err := s.specializationRepository.GetByID(ctx, createdResume.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}

	// Get skills for response
	skills, err := s.resumeRepository.GetSkillsByResumeID(ctx, createdResume.ID)
	if err != nil {
		return nil, err
	}

	// Get additional specializations for response
	additionalSpecializations, err := s.resumeRepository.GetSpecializationsByResumeID(ctx, createdResume.ID)
	if err != nil {
		return nil, err
	}

	// Build response
	response := &dto.ResumeResponse{
		ID:                        createdResume.ID,
		ApplicantID:               createdResume.ApplicantID,
		AboutMe:                   createdResume.AboutMe,
		Specialization:            specializationName,
		Profession:                createdResume.Profession, // Дополнение - добавлено поле профессии
		Skills:                    make([]string, 0, len(skills)),
		AdditionalSpecializations: make([]string, 0, len(additionalSpecializations)),
		WorkExperiences:           make([]dto.WorkExperienceResponse, 0, len(workExperiences)),
		CreatedAt:                 createdResume.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                 createdResume.UpdatedAt.Format(time.RFC3339),
	}

	// Add education info if exists
	if createdResume.Education != "" {
		response.Education = createdResume.Education
	}

	// Add educational institution if exists
	if createdResume.EducationalInstitution != "" {
		response.EducationalInstitution = createdResume.EducationalInstitution
	}

	// Add graduation year if exists
	if !createdResume.GraduationYear.IsZero() {
		response.GraduationYear = createdResume.GraduationYear.Format("2006-01-02")
	}

	// Add skills to response (just names)
	for _, skill := range skills {
		response.Skills = append(response.Skills, skill.Name)
	}

	// Add additional specializations to response (just names)
	for _, spec := range additionalSpecializations {
		response.AdditionalSpecializations = append(response.AdditionalSpecializations, spec.Name)
	}

	// Add work experiences to response
	for _, we := range workExperiences {
		workExp := dto.WorkExperienceResponse{
			ID:           we.ID,
			EmployerName: we.EmployerName,
			Position:     we.Position,
			Duties:       we.Duties,
			Achievements: we.Achievements,
			StartDate:    we.StartDate.Format("2006-01-02"),
			UntilNow:     we.UntilNow,
			UpdatedAt:    we.UpdatedAt.Format(time.RFC3339),
		}

		if !we.UntilNow && !we.EndDate.IsZero() {
			workExp.EndDate = we.EndDate.Format("2006-01-02")
		}

		response.WorkExperiences = append(response.WorkExperiences, workExp)
	}

	return response, nil
}

func (s *ResumeService) GetByID(ctx context.Context, id int) (*dto.ResumeResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"resumeID":  id,
	}).Info("Получение резюме по ID")

	// Get resume
	resume, err := s.resumeRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get specialization name
	var specializationName string
	if resume.SpecializationID != 0 {
		specialization, err := s.specializationRepository.GetByID(ctx, resume.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}

	// Get skills
	skills, err := s.resumeRepository.GetSkillsByResumeID(ctx, resume.ID)
	if err != nil {
		return nil, err
	}

	// Get additional specializations
	additionalSpecializations, err := s.resumeRepository.GetSpecializationsByResumeID(ctx, resume.ID)
	if err != nil {
		return nil, err
	}

	// Get work experiences
	workExperiences, err := s.resumeRepository.GetWorkExperienceByResumeID(ctx, resume.ID)
	if err != nil {
		return nil, err
	}

	// Build response
	response := &dto.ResumeResponse{
		ID:                        resume.ID,
		ApplicantID:               resume.ApplicantID,
		AboutMe:                   resume.AboutMe,
		Specialization:            specializationName,
		Profession:                resume.Profession, // Дополнение - добавлено поле профессии
		Skills:                    make([]string, 0, len(skills)),
		AdditionalSpecializations: make([]string, 0, len(additionalSpecializations)),
		WorkExperiences:           make([]dto.WorkExperienceResponse, 0, len(workExperiences)),
		CreatedAt:                 resume.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                 resume.UpdatedAt.Format(time.RFC3339),
	}

	// Add education info if exists
	if resume.Education != "" {
		response.Education = resume.Education
	}

	// Add educational institution if exists
	if resume.EducationalInstitution != "" {
		response.EducationalInstitution = resume.EducationalInstitution
	}

	// Add graduation year if exists
	if !resume.GraduationYear.IsZero() {
		response.GraduationYear = resume.GraduationYear.Format("2006-01-02")
	}

	// Add skills to response (just names)
	for _, skill := range skills {
		response.Skills = append(response.Skills, skill.Name)
	}

	// Add additional specializations to response (just names)
	for _, spec := range additionalSpecializations {
		response.AdditionalSpecializations = append(response.AdditionalSpecializations, spec.Name)
	}

	// Add work experiences to response
	for _, we := range workExperiences {
		workExp := dto.WorkExperienceResponse{
			ID:           we.ID,
			EmployerName: we.EmployerName,
			Position:     we.Position,
			Duties:       we.Duties,
			Achievements: we.Achievements,
			StartDate:    we.StartDate.Format("2006-01-02"),
			UntilNow:     we.UntilNow,
			UpdatedAt:    we.UpdatedAt.Format(time.RFC3339),
		}

		if !we.UntilNow && !we.EndDate.IsZero() {
			workExp.EndDate = we.EndDate.Format("2006-01-02")
		}

		response.WorkExperiences = append(response.WorkExperiences, workExp)
	}

	return response, nil
}

func (s *ResumeService) Update(ctx context.Context, id int, applicantID int, request *dto.UpdateResumeRequest) (*dto.ResumeResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":   requestID,
		"resumeID":    id,
		"applicantID": applicantID,
	}).Info("Обновление резюме")

	// Check if resume exists
	existingResume, err := s.resumeRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if resume belongs to the applicant
	if existingResume.ApplicantID != applicantID {
		return nil, entity.NewError(
			entity.ErrForbidden,
			fmt.Errorf("резюме с id=%d не принадлежит соискателю с id=%d", id, applicantID),
		)
	}

	// Find specialization ID by name
	var specializationID int
	if request.Specialization != "" {
		specializationID, err = s.resumeRepository.FindSpecializationIDByName(ctx, request.Specialization)
		if err != nil {
			return nil, err
		}
	}

	// Parse graduation year if provided
	var graduationYear time.Time
	if request.GraduationYear != "" {
		graduationYear, err = time.Parse("2006-01-02", request.GraduationYear)
		if err != nil {
			return nil, entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты окончания учебы: %w", err),
			)
		}
	}

	// Create resume entity for update
	resume := &entity.Resume{
		ID:                     id,
		ApplicantID:            applicantID,
		AboutMe:                request.AboutMe,
		SpecializationID:       specializationID,
		Education:              request.Education,
		EducationalInstitution: request.EducationalInstitution,
		GraduationYear:         graduationYear,
		Profession:             request.Profession, // Дополнение - добавлено поле профессии
	}

	// Validate resume
	if err := resume.Validate(); err != nil {
		return nil, err
	}

	// Update resume in DB
	updatedResume, err := s.resumeRepository.Update(ctx, resume)
	if err != nil {
		return nil, err
	}

	// Update skills
	if err := s.resumeRepository.DeleteSkills(ctx, id); err != nil {
		return nil, err
	}
	if len(request.Skills) > 0 {
		skillIDs, err := s.resumeRepository.FindSkillIDsByNames(ctx, request.Skills)
		if err != nil {
			return nil, err
		}
		if len(skillIDs) > 0 {
			if err := s.resumeRepository.AddSkills(ctx, id, skillIDs); err != nil {
				return nil, err
			}
		}
	}

	// Update specializations
	if err := s.resumeRepository.DeleteSpecializations(ctx, id); err != nil {
		return nil, err
	}
	if len(request.AdditionalSpecializations) > 0 {
		specIDs, err := s.resumeRepository.FindSpecializationIDsByNames(ctx, request.AdditionalSpecializations)
		if err != nil {
			return nil, err
		}
		if len(specIDs) > 0 {
			if err := s.resumeRepository.AddSpecializations(ctx, id, specIDs); err != nil {
				return nil, err
			}
		}
	}

	// Update work experiences
	if err := s.resumeRepository.DeleteWorkExperiences(ctx, id); err != nil {
		return nil, err
	}

	var workExperiences []entity.WorkExperience
	for _, we := range request.WorkExperiences {
		startDate, err := time.Parse("2006-01-02", we.StartDate)
		if err != nil {
			return nil, entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты начала работы: %w", err),
			)
		}

		var endDate time.Time
		if !we.UntilNow && we.EndDate != "" {
			endDate, err = time.Parse("2006-01-02", we.EndDate)
			if err != nil {
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неверный формат даты окончания работы: %w", err),
				)
			}
		}

		workExperience := entity.WorkExperience{
			ResumeID:     id,
			EmployerName: we.EmployerName,
			Position:     we.Position,
			Duties:       we.Duties,
			Achievements: we.Achievements,
			StartDate:    startDate,
			EndDate:      endDate,
			UntilNow:     we.UntilNow,
		}

		if err := workExperience.Validate(); err != nil {
			return nil, err
		}

		createdWorkExperience, err := s.resumeRepository.AddWorkExperience(ctx, &workExperience)
		if err != nil {
			return nil, err
		}

		workExperiences = append(workExperiences, *createdWorkExperience)
	}

	// Get specialization name
	var specializationName string
	if updatedResume.SpecializationID != 0 {
		specialization, err := s.specializationRepository.GetByID(ctx, updatedResume.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}

	// Get skills for response
	skills, err := s.resumeRepository.GetSkillsByResumeID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get additional specializations for response
	additionalSpecializations, err := s.resumeRepository.GetSpecializationsByResumeID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build response
	response := &dto.ResumeResponse{
		ID:                        updatedResume.ID,
		ApplicantID:               updatedResume.ApplicantID,
		AboutMe:                   updatedResume.AboutMe,
		Specialization:            specializationName,
		Profession:                updatedResume.Profession, // Дополнение - добавлено поле профессии
		Skills:                    make([]string, 0, len(skills)),
		AdditionalSpecializations: make([]string, 0, len(additionalSpecializations)),
		WorkExperiences:           make([]dto.WorkExperienceResponse, 0, len(workExperiences)),
		CreatedAt:                 updatedResume.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                 updatedResume.UpdatedAt.Format(time.RFC3339),
	}

	// Add education info if exists
	if updatedResume.Education != "" {
		response.Education = updatedResume.Education
	}

	// Add educational institution if exists
	if updatedResume.EducationalInstitution != "" {
		response.EducationalInstitution = updatedResume.EducationalInstitution
	}

	// Add graduation year if exists
	if !updatedResume.GraduationYear.IsZero() {
		response.GraduationYear = updatedResume.GraduationYear.Format("2006-01-02")
	}

	// Add skills to response (just names)
	for _, skill := range skills {
		response.Skills = append(response.Skills, skill.Name)
	}

	// Add additional specializations to response (just names)
	for _, spec := range additionalSpecializations {
		response.AdditionalSpecializations = append(response.AdditionalSpecializations, spec.Name)
	}

	// Add work experiences to response
	for _, we := range workExperiences {
		workExp := dto.WorkExperienceResponse{
			ID:           we.ID,
			EmployerName: we.EmployerName,
			Position:     we.Position,
			Duties:       we.Duties,
			Achievements: we.Achievements,
			StartDate:    we.StartDate.Format("2006-01-02"),
			UntilNow:     we.UntilNow,
			UpdatedAt:    we.UpdatedAt.Format(time.RFC3339),
		}

		if !we.UntilNow && !we.EndDate.IsZero() {
			workExp.EndDate = we.EndDate.Format("2006-01-02")
		}

		response.WorkExperiences = append(response.WorkExperiences, workExp)
	}

	return response, nil
}

func (s *ResumeService) Delete(ctx context.Context, id int, applicantID int) (*dto.DeleteResumeResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":   requestID,
		"resumeID":    id,
		"applicantID": applicantID,
	}).Info("Удаление резюме")

	// Проверяем существование резюме
	existingResume, err := s.resumeRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Проверяем, что резюме принадлежит указанному соискателю
	if existingResume.ApplicantID != applicantID {
		return nil, entity.NewError(
			entity.ErrForbidden,
			fmt.Errorf("резюме с id=%d не принадлежит соискателю с id=%d", id, applicantID),
		)
	}

	// Удаляем связанные данные
	if err := s.resumeRepository.DeleteWorkExperiences(ctx, id); err != nil {
		return nil, err
	}

	if err := s.resumeRepository.DeleteSkills(ctx, id); err != nil {
		return nil, err
	}

	if err := s.resumeRepository.DeleteSpecializations(ctx, id); err != nil {
		return nil, err
	}

	// Удаляем само резюме
	if err := s.resumeRepository.Delete(ctx, id); err != nil {
		return nil, err
	}

	return &dto.DeleteResumeResponse{
		Success: true,
		Message: fmt.Sprintf("Резюме с id=%d успешно удалено", id),
	}, nil
}

// GetAll returns a list of all resumes (for employers)
func (s *ResumeService) GetAll(ctx context.Context, limit int, offset int) ([]dto.ResumeShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("Получение списка всех резюме")

	// Get all resumes with limit
	resumes, err := s.resumeRepository.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Build response
	response := make([]dto.ResumeShortResponse, 0, len(resumes))
	for _, resume := range resumes {
		// Get specialization name
		var specializationName string
		if resume.SpecializationID != 0 {
			specialization, err := s.specializationRepository.GetByID(ctx, resume.SpecializationID)
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

		// Get the most recent work experience
		workExperiences, err := s.resumeRepository.GetWorkExperienceByResumeID(ctx, resume.ID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"resumeID":  resume.ID,
				"error":     err,
			}).Error("ошибка при получении опыта работы")
			continue
		}

		// // Get applicant information
		// applicantProfile, err := s.applicantRepository.GetApplicantByID(ctx, resume.ApplicantID)
		// if err != nil {
		// 	l.Log.WithFields(logrus.Fields{
		// 		"requestID":   requestID,
		// 		"resumeID":    resume.ID,
		// 		"applicantID": resume.ApplicantID,
		// 		"error":       err,
		// 	}).Error("ошибка при получении информации о соискателе")
		// 	continue
		// }
		// Convert applicant entity to DTO
		applicantDTO, err := s.applicantService.GetUser(ctx, resume.ApplicantID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":   requestID,
				"resumeID":    resume.ID,
				"applicantID": resume.ApplicantID,
				"error":       err,
			}).Error("ошибка при конвертации соискателя в DTO")
			continue
		}

		// Create short resume response
		shortResume := dto.ResumeShortResponse{
			ID:             resume.ID,
			Applicant:      applicantDTO,
			Specialization: specializationName,
			Profession:     resume.Profession, // Дополнение - добавлено поле профессии
			CreatedAt:      resume.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      resume.UpdatedAt.Format(time.RFC3339),
		}

		// Add the most recent work experience if available
		if len(workExperiences) > 0 {
			we := workExperiences[0] // The first one is the most recent due to ORDER BY in the query
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

// GetAll returns a list of all resumes (for applicants)
func (s *ResumeService) GetAllResumesByApplicantID(ctx context.Context, applicantID int, limit int, offset int) ([]dto.ResumeApplicantShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":   requestID,
		"applicantID": applicantID,
	}).Info("Получение списка всех резюме соискателя")

	// Get all resumes with limit
	resumes, err := s.resumeRepository.GetAllResumesByApplicantID(ctx, applicantID, limit, offset)
	if err != nil {
		return nil, err
	}

	// // Get applicant information once since all resumes belong to the same applicant
	// applicantDTO, err := s.applicantService.GetUser(ctx, applicantID)
	// if err != nil {
	// 	l.Log.WithFields(logrus.Fields{
	// 		"requestID":   requestID,
	// 		"applicantID": applicantID,
	// 		"error":       err,
	// 	}).Error("ошибка при получении информации о соискателе")
	// 	return nil, err
	// }

	// Build response
	response := make([]dto.ResumeApplicantShortResponse, 0, len(resumes))
	for _, resume := range resumes {
		// Get specialization name
		var specializationName string
		if resume.SpecializationID != 0 {
			specialization, err := s.specializationRepository.GetByID(ctx, resume.SpecializationID)
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

		// Get the most recent work experience
		workExperiences, err := s.resumeRepository.GetWorkExperienceByResumeID(ctx, resume.ID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"resumeID":  resume.ID,
				"error":     err,
			}).Error("ошибка при получении опыта работы")
			continue
		}

		// Convert applicant entity to DTO
		applicantDTO, err := s.applicantService.GetUser(ctx, resume.ApplicantID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":   requestID,
				"resumeID":    resume.ID,
				"applicantID": resume.ApplicantID,
				"error":       err,
			}).Error("ошибка при конвертации соискателя в DTO")
			continue
		}

		// Получаем навыки для резюме - добавлено для нового DTO
		skills, err := s.resumeRepository.GetSkillsByResumeID(ctx, resume.ID)
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

		// Create short resume response
		shortResume := dto.ResumeApplicantShortResponse{
			ID:             resume.ID,
			Applicant:      applicantDTO,
			Skills:         skillNames,
			Specialization: specializationName,
			Profession:     resume.Profession, // Дополнение - добавлено поле профессии
			CreatedAt:      resume.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      resume.UpdatedAt.Format(time.RFC3339),
		}

		// Add the most recent work experience if available
		if len(workExperiences) > 0 {
			we := workExperiences[0] // The first one is the most recent due to ORDER BY in the query
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

// SearchResumesByProfession ищет резюме по профессии с учетом роли пользователя
func (s *ResumeService) SearchResumesByProfession(ctx context.Context, userID int, role string, profession string, limit int, offset int) ([]dto.ResumeShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"userID":     userID,
		"role":       role,
		"profession": profession,
	}).Info("Поиск резюме по профессии")

	var resumes []entity.Resume
	var err error

	// В зависимости от роли пользователя выбираем метод поиска
	if role == "applicant" {
		// Для соискателя ищем только его резюме
		resumes, err = s.resumeRepository.SearchResumesByProfessionForApplicant(ctx, userID, profession, limit, offset)
	} else {
		// Для работодателя ищем все резюме
		resumes, err = s.resumeRepository.SearchResumesByProfession(ctx, profession, limit, offset)
	}

	if err != nil {
		return nil, err
	}

	// Формируем ответ, аналогично методам GetAll и GetAllResumesByApplicantID
	response := make([]dto.ResumeShortResponse, 0, len(resumes))
	for _, resume := range resumes {
		// Получаем имя специализации
		var specializationName string
		if resume.SpecializationID != 0 {
			specialization, err := s.specializationRepository.GetByID(ctx, resume.SpecializationID)
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

		// Получаем последний опыт работы
		workExperiences, err := s.resumeRepository.GetWorkExperienceByResumeID(ctx, resume.ID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"resumeID":  resume.ID,
				"error":     err,
			}).Error("ошибка при получении опыта работы")
			continue
		}

		// Получаем информацию о соискателе
		applicantDTO, err := s.applicantService.GetUser(ctx, resume.ApplicantID)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":   requestID,
				"resumeID":    resume.ID,
				"applicantID": resume.ApplicantID,
				"error":       err,
			}).Error("ошибка при конвертации соискателя в DTO")
			continue
		}

		// Создаем краткий ответ о резюме
		shortResume := dto.ResumeShortResponse{
			ID:             resume.ID,
			Applicant:      applicantDTO,
			Specialization: specializationName,
			Profession:     resume.Profession,
			CreatedAt:      resume.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      resume.UpdatedAt.Format(time.RFC3339),
		}

		// Добавляем последний опыт работы, если он есть
		if len(workExperiences) > 0 {
			we := workExperiences[0] // Первый - самый последний из-за ORDER BY в запросе
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
