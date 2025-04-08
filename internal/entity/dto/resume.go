package dto

type CreateResumeRequest struct {
	ApplicantID               int                 `json:"applicant_id" validate:"required"`
	AboutMe                   string              `json:"about_me" validate:"required"`
	SpecializationID          int                 `json:"specialization_id" validate:"required"`
	Education                 int                 `json:"education" validate:"required"`
	EducationalInstitution    string              `json:"educational_institution" validate:"required"`
	GraduationYear            string              `json:"graduation_year" validate:"required"`
	Skills                    []int               `json:"skills" validate:"required"`
	AdditionalSpecializations []int               `json:"additional_specializations"`
	WorkExperiences           []WorkExperienceDTO `json:"work_experiences"`
}

type WorkExperienceDTO struct {
	EmployerName string `json:"employer_name" validate:"required"`
	Position     string `json:"position" validate:"required"`
	Duties       string `json:"duties" validate:"required"`
	Achievements string `json:"achievements"`
	StartDate    string `json:"start_date" validate:"required"`
	EndDate      string `json:"end_date"`
	UntilNow     bool   `json:"until_now"`
}

type ResumeResponse struct {
	ID                        int                      `json:"id"`
	ApplicantID               int                      `json:"applicant_id"`
	AboutMe                   string                   `json:"about_me"`
	SpecializationID          int                      `json:"specialization_id"`
	SpecializationName        string                   `json:"specialization_name"`
	Education                 int                      `json:"education"`
	EducationalInstitution    string                   `json:"educational_institution"`
	GraduationYear            string                   `json:"graduation_year"`
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
	Duties       string `json:"duties"`
	Achievements string `json:"achievements"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	UntilNow     bool   `json:"until_now"`
}
