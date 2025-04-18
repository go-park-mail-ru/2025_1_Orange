package dto

import (
	"ResuMatch/internal/entity"
)

// Updated CreateResumeRequest - removed applicant_id, changed specialization and skills to strings
type CreateResumeRequest struct {
	AboutMe                   string               `json:"about_me"`
	Specialization            string               `json:"specialization"`
	Education                 entity.EducationType `json:"education"`
	EducationalInstitution    string               `json:"educational_institution"`
	GraduationYear            string               `json:"graduation_year"`
	Skills                    []string             `json:"skills"`
	AdditionalSpecializations []string             `json:"additional_specializations"`
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
	Specialization            string                   `json:"specialization,omitempty"`
	Education                 entity.EducationType     `json:"education,omitempty"`
	EducationalInstitution    string                   `json:"educational_institution,omitempty"`
	GraduationYear            string                   `json:"graduation_year,omitempty"`
	CreatedAt                 string                   `json:"created_at"`
	UpdatedAt                 string                   `json:"updated_at"`
	Skills                    []string                 `json:"skills"`
	AdditionalSpecializations []string                 `json:"additional_specializations"`
	WorkExperiences           []WorkExperienceResponse `json:"work_experiences"`
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

// Updated UpdateResumeRequest - similar changes as CreateResumeRequest
type UpdateResumeRequest struct {
	AboutMe                   string               `json:"about_me"`
	Specialization            string               `json:"specialization"`
	Education                 entity.EducationType `json:"education"`
	EducationalInstitution    string               `json:"educational_institution"`
	GraduationYear            string               `json:"graduation_year"`
	Skills                    []string             `json:"skills"`
	AdditionalSpecializations []string             `json:"additional_specializations"`
	WorkExperiences           []WorkExperienceDTO  `json:"work_experiences"`
}

type DeleteResumeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// New DTO for resume list
type ResumeShortResponse struct {
	ID             int                 `json:"id"`
	ApplicantID    int                 `json:"applicant_id"`
	Specialization string              `json:"specialization"`
	WorkExperience WorkExperienceShort `json:"work_experiences"`
	CreatedAt      string              `json:"created_at"`
	UpdatedAt      string              `json:"updated_at"`
}

type WorkExperienceShort struct {
	ID           int    `json:"id"`
	EmployerName string `json:"employer_name"`
	Position     string `json:"position"`
	Duties       string `json:"duties,omitempty"`
	Achievements string `json:"achievements,omitempty"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date,omitempty"`
	UntilNow     bool   `json:"until_now"`
}
