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

// Create method update
func (s *ResumeService) Create(ctx context.Context, request *dto.CreateResumeRequest) (*dto.ResumeResponse, error) {
	// Check if specialization exists if provided
	var specialization *entity.Specialization
	var err error

	if request.SpecializationID != 0 {
		specialization, err = s.specializationRepository.GetByID(ctx, request.SpecializationID)
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
		ApplicantID:               request.ApplicantID,
		AboutMe:                   request.AboutMe,
		SpecializationID:          request.SpecializationID,
		Education:                 request.Education,
		EducationalInstitution:    request.EducationalInstitution,
		GraduationYear:            graduationYear,
		Skills:                    request.Skills,
		AdditionalSpecializations: request.AdditionalSpecializations,
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

	// Add skills if provided
	if len(request.Skills) > 0 {
		if err := s.resumeRepository.AddSkills(ctx, createdResume.ID, request.Skills); err != nil {
			return nil, err
		}
	}

	// Add additional specializations if provided
	if len(request.AdditionalSpecializations) > 0 {
		if err := s.resumeRepository.AddSpecializations(ctx, createdResume.ID, request.AdditionalSpecializations); err != nil {
			return nil, err
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
		Skills:                    make([]dto.SkillDTO, 0, len(skills)),
		AdditionalSpecializations: make([]dto.SpecializationDTO, 0, len(additionalSpecializations)),
		WorkExperiences:           make([]dto.WorkExperienceResponse, 0, len(workExperiences)),
		CreatedAt:                 createdResume.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                 createdResume.UpdatedAt.Format(time.RFC3339),
	}

	// Add specialization info if exists
	if createdResume.SpecializationID != 0 {
		response.SpecializationID = createdResume.SpecializationID
		if specialization != nil {
			response.SpecializationName = specialization.Name
		}
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

	// Add skills to response
	for _, skill := range skills {
		response.Skills = append(response.Skills, dto.SkillDTO{
			ID:   skill.ID,
			Name: skill.Name,
		})
	}

	// Add additional specializations to response
	for _, spec := range additionalSpecializations {
		response.AdditionalSpecializations = append(response.AdditionalSpecializations, dto.SpecializationDTO{
			ID:   spec.ID,
			Name: spec.Name,
		})
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

// GetByID method update
func (s *ResumeService) GetByID(ctx context.Context, id int) (*dto.ResumeResponse, error) {
	// Get resume
	resume, err := s.resumeRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get specialization if exists
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
		Skills:                    make([]dto.SkillDTO, 0, len(skills)),
		AdditionalSpecializations: make([]dto.SpecializationDTO, 0, len(additionalSpecializations)),
		WorkExperiences:           make([]dto.WorkExperienceResponse, 0, len(workExperiences)),
		CreatedAt:                 resume.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                 resume.UpdatedAt.Format(time.RFC3339),
	}

	// Add specialization info if exists
	if resume.SpecializationID != 0 {
		response.SpecializationID = resume.SpecializationID
		response.SpecializationName = specializationName
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

	// Add skills to response
	for _, skill := range skills {
		response.Skills = append(response.Skills, dto.SkillDTO{
			ID:   skill.ID,
			Name: skill.Name,
		})
	}

	// Add additional specializations to response
	for _, spec := range additionalSpecializations {
		response.AdditionalSpecializations = append(response.AdditionalSpecializations, dto.SpecializationDTO{
			ID:   spec.ID,
			Name: spec.Name,
		})
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

// Update method update
func (s *ResumeService) Update(ctx context.Context, id int, request *dto.UpdateResumeRequest) (*dto.ResumeResponse, error) {
	// Check if resume exists
	existingResume, err := s.resumeRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if resume belongs to the applicant
	if existingResume.ApplicantID != request.ApplicantID {
		return nil, entity.NewError(
			entity.ErrForbidden,
			fmt.Errorf("резюме с id=%d не принадлежит соискателю с id=%d", id, request.ApplicantID),
		)
	}

	// Check if specialization exists if provided
	var specialization *entity.Specialization
	if request.SpecializationID != 0 {
		specialization, err = s.specializationRepository.GetByID(ctx, request.SpecializationID)
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
		if err := s.resumeRepository.AddSkills(ctx, id, request.Skills); err != nil {
			return nil, err
		}
	}

	// Update specializations
	if err := s.resumeRepository.DeleteSpecializations(ctx, id); err != nil {
		return nil, err
	}
	if len(request.AdditionalSpecializations) > 0 {
		if err := s.resumeRepository.AddSpecializations(ctx, id, request.AdditionalSpecializations); err != nil {
			return nil, err
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
		Skills:                    make([]dto.SkillDTO, 0, len(skills)),
		AdditionalSpecializations: make([]dto.SpecializationDTO, 0, len(additionalSpecializations)),
		WorkExperiences:           make([]dto.WorkExperienceResponse, 0, len(workExperiences)),
		CreatedAt:                 updatedResume.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                 updatedResume.UpdatedAt.Format(time.RFC3339),
	}

	// Add specialization info if exists
	if updatedResume.SpecializationID != 0 {
		response.SpecializationID = updatedResume.SpecializationID
		if specialization != nil {
			response.SpecializationName = specialization.Name
		}
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

	// Add skills to response
	for _, skill := range skills {
		response.Skills = append(response.Skills, dto.SkillDTO{
			ID:   skill.ID,
			Name: skill.Name,
		})
	}

	// Add additional specializations to response
	for _, spec := range additionalSpecializations {
		response.AdditionalSpecializations = append(response.AdditionalSpecializations, dto.SpecializationDTO{
			ID:   spec.ID,
			Name: spec.Name,
		})
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
