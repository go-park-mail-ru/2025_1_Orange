package dto

import "ResuMatch/internal/entity"

type CreateResumeRequest struct {
	ApplicantID               int                  `json:"applicant_id" validate:"required"`
	AboutMe                   string               `json:"about_me"`
	SpecializationID          int                  `json:"specialization_id"`
	Education                 entity.EducationType `json:"education"`
	EducationalInstitution    string               `json:"educational_institution"`
	GraduationYear            string               `json:"graduation_year"`
	Skills                    []int                `json:"skills"`
	AdditionalSpecializations []int                `json:"additional_specializations"`
	WorkExperiences           []WorkExperienceDTO  `json:"work_experiences"`
}

type WorkExperienceDTO struct {
	EmployerName string `json:"employer_name" validate:"required,max=64"`
	Position     string `json:"position" validate:"required,max=64"`
	Duties       string `json:"duties"`
	Achievements string `json:"achievements"`
	StartDate    string `json:"start_date" validate:"required"`
	EndDate      string `json:"end_date"`
	UntilNow     bool   `json:"until_now"`
}

type ResumeResponse struct {
	ID                        int                      `json:"id"`
	ApplicantID               int                      `json:"applicant_id"`
	AboutMe                   string                   `json:"about_me,omitempty"`
	SpecializationID          int                      `json:"specialization_id,omitempty"`
	SpecializationName        string                   `json:"specialization_name,omitempty"`
	Education                 entity.EducationType     `json:"education,omitempty"`
	EducationalInstitution    string                   `json:"educational_institution,omitempty"`
	GraduationYear            string                   `json:"graduation_year,omitempty"`
	CreatedAt                 string                   `json:"created_at"`
	UpdatedAt                 string                   `json:"updated_at"`
	Skills                    []SkillDTO               `json:"skills"`
	AdditionalSpecializations []SpecializationDTO      `json:"additional_specializations"`
	WorkExperiences           []WorkExperienceResponse `json:"work_experiences"`
}

type SkillDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SpecializationDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type WorkExperienceResponse struct {
	ID           int    `json:"id"`
	EmployerName string `json:"employer_name"`
	Position     string `json:"position"`
	Duties       string `json:"duties,omitempty"`
	Achievements string `json:"achievements,omitempty"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date,omitempty"`
	UntilNow     bool   `json:"until_now"`
	UpdatedAt    string `json:"updated_at"`
}

type UpdateResumeRequest struct {
	ApplicantID               int                  `json:"applicant_id" validate:"required"`
	AboutMe                   string               `json:"about_me"`
	SpecializationID          int                  `json:"specialization_id"`
	Education                 entity.EducationType `json:"education"`
	EducationalInstitution    string               `json:"educational_institution"`
	GraduationYear            string               `json:"graduation_year"`
	Skills                    []int                `json:"skills"`
	AdditionalSpecializations []int                `json:"additional_specializations"`
	WorkExperiences           []WorkExperienceDTO  `json:"work_experiences"`
}

type DeleteResumeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
