package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"context"
	"fmt"
	"time"
)

type ResumeService struct {
	resumeRepository         repository.ResumeRepository
	skillRepository          repository.SkillRepository
	specializationRepository repository.SpecializationRepository
}

func NewResumeService(
	resumeRepo repository.ResumeRepository,
	skillRepo repository.SkillRepository,
	specializationRepo repository.SpecializationRepository,
) usecase.Resume {
	return &ResumeService{
		resumeRepository:         resumeRepo,
		skillRepository:          skillRepo,
		specializationRepository: specializationRepo,
	}
}

func (s *ResumeService) Create(ctx context.Context, request *dto.CreateResumeRequest) (*dto.ResumeResponse, error) {
	// Проверяем существование специализации
	specialization, err := s.specializationRepository.GetByID(ctx, request.SpecializationID)
	if err != nil {
		return nil, err
	}

	// Преобразуем DTO в сущность
	graduationYear, err := time.Parse("2006-01-02", request.GraduationYear)
	if err != nil {
		return nil, entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("неверный формат даты окончания учебы: %w", err),
		)
	}

	resume := &entity.Resume{
		ApplicantID:               request.ApplicantID,
		AboutMe:                   request.AboutMe,
		SpecializationID:          request.SpecializationID,
		Education:                 request.Education,
		EducationalInstitution:    request.EducationalInstitution,
		GraduationYear:            graduationYear,
		Skills:                    request.Skills,
		AdditionalSpecializations: request.AdditionalSpecializations,
	}

	// Валидируем резюме
	if err := resume.Validate(); err != nil {
		return nil, err
	}

	// Создаем резюме в БД
	createdResume, err := s.resumeRepository.Create(ctx, resume)
	if err != nil {
		return nil, err
	}

	// Добавляем навыки
	if len(request.Skills) > 0 {
		if err := s.resumeRepository.AddSkills(ctx, createdResume.ID, request.Skills); err != nil {
			return nil, err
		}
	}

	// Добавляем дополнительные специализации
	if len(request.AdditionalSpecializations) > 0 {
		if err := s.resumeRepository.AddSpecializations(ctx, createdResume.ID, request.AdditionalSpecializations); err != nil {
			return nil, err
		}
	}

	// Добавляем опыт работы
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

	// Получаем навыки для ответа
	skills, err := s.resumeRepository.GetSkillsByResumeID(ctx, createdResume.ID)
	if err != nil {
		return nil, err
	}

	// Получаем дополнительные специализации для ответа
	additionalSpecializations, err := s.resumeRepository.GetSpecializationsByResumeID(ctx, createdResume.ID)
	if err != nil {
		return nil, err
	}

	// Формируем ответ
	response := &dto.ResumeResponse{
		ID:                        createdResume.ID,
		ApplicantID:               createdResume.ApplicantID,
		AboutMe:                   createdResume.AboutMe,
		SpecializationID:          createdResume.SpecializationID,
		SpecializationName:        specialization.Name,
		Education:                 createdResume.Education,
		EducationalInstitution:    createdResume.EducationalInstitution,
		GraduationYear:            createdResume.GraduationYear.Format("2006-01-02"),
		CreatedAt:                 createdResume.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                 createdResume.UpdatedAt.Format(time.RFC3339),
		Skills:                    make([]dto.SkillDTO, 0, len(skills)),
		AdditionalSpecializations: make([]dto.SpecializationDTO, 0, len(additionalSpecializations)),
		WorkExperiences:           make([]dto.WorkExperienceResponse, 0, len(workExperiences)),
	}

	// Добавляем навыки в ответ
	for _, skill := range skills {
		response.Skills = append(response.Skills, dto.SkillDTO{
			ID:   skill.ID,
			Name: skill.Name,
		})
	}

	// Добавляем дополнительные специализации в ответ
	for _, spec := range additionalSpecializations {
		response.AdditionalSpecializations = append(response.AdditionalSpecializations, dto.SpecializationDTO{
			ID:   spec.ID,
			Name: spec.Name,
		})
	}

	// Добавляем опыт работы в ответ
	for _, we := range workExperiences {
		workExp := dto.WorkExperienceResponse{
			ID:           we.ID,
			EmployerName: we.EmployerName,
			Position:     we.Position,
			Duties:       we.Duties,
			Achievements: we.Achievements,
			StartDate:    we.StartDate.Format("2006-01-02"),
			UntilNow:     we.UntilNow,
		}

		if !we.UntilNow {
			workExp.EndDate = we.EndDate.Format("2006-01-02")
		}

		response.WorkExperiences = append(response.WorkExperiences, workExp)
	}

	return response, nil
}

func (s *ResumeService) GetByID(ctx context.Context, id int) (*dto.ResumeResponse, error) {
	// Получаем резюме
	resume, err := s.resumeRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Получаем основную специализацию
	specialization, err := s.specializationRepository.GetByID(ctx, resume.SpecializationID)
	if err != nil {
		return nil, err
	}

	// Получаем навыки
	skills, err := s.resumeRepository.GetSkillsByResumeID(ctx, resume.ID)
	if err != nil {
		return nil, err
	}

	// Получаем дополнительные специализации
	additionalSpecializations, err := s.resumeRepository.GetSpecializationsByResumeID(ctx, resume.ID)
	if err != nil {
		return nil, err
	}

	// Получаем опыт работы
	workExperiences, err := s.resumeRepository.GetWorkExperienceByResumeID(ctx, resume.ID)
	if err != nil {
		return nil, err
	}

	// Формируем ответ
	response := &dto.ResumeResponse{
		ID:                        resume.ID,
		ApplicantID:               resume.ApplicantID,
		AboutMe:                   resume.AboutMe,
		SpecializationID:          resume.SpecializationID,
		SpecializationName:        specialization.Name,
		Education:                 resume.Education,
		EducationalInstitution:    resume.EducationalInstitution,
		GraduationYear:            resume.GraduationYear.Format("2006-01-02"),
		CreatedAt:                 resume.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                 resume.UpdatedAt.Format(time.RFC3339),
		Skills:                    make([]dto.SkillDTO, 0, len(skills)),
		AdditionalSpecializations: make([]dto.SpecializationDTO, 0, len(additionalSpecializations)),
		WorkExperiences:           make([]dto.WorkExperienceResponse, 0, len(workExperiences)),
	}

	// Добавляем навыки в ответ
	for _, skill := range skills {
		response.Skills = append(response.Skills, dto.SkillDTO{
			ID:   skill.ID,
			Name: skill.Name,
		})
	}

	// Добавляем дополнительные специализации в ответ
	for _, spec := range additionalSpecializations {
		response.AdditionalSpecializations = append(response.AdditionalSpecializations, dto.SpecializationDTO{
			ID:   spec.ID,
			Name: spec.Name,
		})
	}

	// Добавляем опыт работы в ответ
	for _, we := range workExperiences {
		workExp := dto.WorkExperienceResponse{
			ID:           we.ID,
			EmployerName: we.EmployerName,
			Position:     we.Position,
			Duties:       we.Duties,
			Achievements: we.Achievements,
			StartDate:    we.StartDate.Format("2006-01-02"),
			UntilNow:     we.UntilNow,
		}

		if !we.UntilNow {
			workExp.EndDate = we.EndDate.Format("2006-01-02")
		}

		response.WorkExperiences = append(response.WorkExperiences, workExp)
	}

	return response, nil
}

// Update обновляет резюме
func (s *ResumeService) Update(ctx context.Context, id int, request *dto.UpdateResumeRequest) (*dto.ResumeResponse, error) {
	// Проверяем существование резюме
	existingResume, err := s.resumeRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Проверяем, что резюме принадлежит указанному соискателю
	if existingResume.ApplicantID != request.ApplicantID {
		return nil, entity.NewError(
			entity.ErrForbidden,
			fmt.Errorf("резюме с id=%d не принадлежит соискателю с id=%d", id, request.ApplicantID),
		)
	}

	// Проверяем существование специализации
	specialization, err := s.specializationRepository.GetByID(ctx, request.SpecializationID)
	if err != nil {
		return nil, err
	}

	// Преобразуем DTO в сущность
	graduationYear, err := time.Parse("2006-01-02", request.GraduationYear)
	if err != nil {
		return nil, entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("неверный формат даты окончания учебы: %w", err),
		)
	}

	resume := &entity.Resume{
		ID:                        id,
		ApplicantID:               request.ApplicantID,
		AboutMe:                   request.AboutMe,
		SpecializationID:          request.SpecializationID,
		Education:                 request.Education,
		EducationalInstitution:    request.EducationalInstitution,
		GraduationYear:            graduationYear,
		Skills:                    request.Skills,
		AdditionalSpecializations: request.AdditionalSpecializations,
	}

	// Валидируем резюме
	if err := resume.Validate(); err != nil {
		return nil, err
	}

	// Начинаем транзакцию обновления
	// 1. Обновляем основную информацию о резюме
	updatedResume, err := s.resumeRepository.Update(ctx, resume)
	if err != nil {
		return nil, err
	}

	// 2. Удаляем старые навыки и добавляем новые
	if err := s.resumeRepository.DeleteSkills(ctx, id); err != nil {
		return nil, err
	}
	if len(request.Skills) > 0 {
		if err := s.resumeRepository.AddSkills(ctx, id, request.Skills); err != nil {
			return nil, err
		}
	}

	// 3. Удаляем старые специализации и добавляем новые
	if err := s.resumeRepository.DeleteSpecializations(ctx, id); err != nil {
		return nil, err
	}
	if len(request.AdditionalSpecializations) > 0 {
		if err := s.resumeRepository.AddSpecializations(ctx, id, request.AdditionalSpecializations); err != nil {
			return nil, err
		}
	}

	// 4. Удаляем старый опыт работы и добавляем новый
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

	// Получаем навыки для ответа
	skills, err := s.resumeRepository.GetSkillsByResumeID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Получаем дополнительные специализации для ответа
	additionalSpecializations, err := s.resumeRepository.GetSpecializationsByResumeID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Формируем ответ
	response := &dto.ResumeResponse{
		ID:                        updatedResume.ID,
		ApplicantID:               updatedResume.ApplicantID,
		AboutMe:                   updatedResume.AboutMe,
		SpecializationID:          updatedResume.SpecializationID,
		SpecializationName:        specialization.Name,
		Education:                 updatedResume.Education,
		EducationalInstitution:    updatedResume.EducationalInstitution,
		GraduationYear:            updatedResume.GraduationYear.Format("2006-01-02"),
		CreatedAt:                 updatedResume.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                 updatedResume.UpdatedAt.Format(time.RFC3339),
		Skills:                    make([]dto.SkillDTO, 0, len(skills)),
		AdditionalSpecializations: make([]dto.SpecializationDTO, 0, len(additionalSpecializations)),
		WorkExperiences:           make([]dto.WorkExperienceResponse, 0, len(workExperiences)),
	}

	// Добавляем навыки в ответ
	for _, skill := range skills {
		response.Skills = append(response.Skills, dto.SkillDTO{
			ID:   skill.ID,
			Name: skill.Name,
		})
	}

	// Добавляем дополнительные специализации в ответ
	for _, spec := range additionalSpecializations {
		response.AdditionalSpecializations = append(response.AdditionalSpecializations, dto.SpecializationDTO{
			ID:   spec.ID,
			Name: spec.Name,
		})
	}

	// Добавляем опыт работы в ответ
	for _, we := range workExperiences {
		workExp := dto.WorkExperienceResponse{
			ID:           we.ID,
			EmployerName: we.EmployerName,
			Position:     we.Position,
			Duties:       we.Duties,
			Achievements: we.Achievements,
			StartDate:    we.StartDate.Format("2006-01-02"),
			UntilNow:     we.UntilNow,
		}

		if !we.UntilNow {
			workExp.EndDate = we.EndDate.Format("2006-01-02")
		}

		response.WorkExperiences = append(response.WorkExperiences, workExp)
	}

	return response, nil
}

// Delete удаляет резюме
func (s *ResumeService) Delete(ctx context.Context, id int, applicantID int) (*dto.DeleteResumeResponse, error) {
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
