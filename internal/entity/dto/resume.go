package dto

import (
	"ResuMatch/internal/entity"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
)

func init() {
	govalidator.CustomTypeTagMap.Set("customYearValidation", func(i interface{}, _ interface{}) bool {
		dateStr, ok := i.(string)
		if !ok || dateStr == "" {
			return false
		}

		var year int

		// Пробуем распарсить как полную дату
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			year = t.Year()
		} else if y, err := strconv.Atoi(dateStr); err == nil { // Пробуем распарсить как просто год
			year = y
		} else {
			return false
		}

		currentYear := time.Now().Year()
		return year >= (currentYear-50) && year <= (currentYear+5)
	})

	// Регистрируем кастомный валидатор для формата даты YYYY-MM-DD
	govalidator.CustomTypeTagMap.Set("date_iso", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
		str, ok := i.(string)
		if !ok {
			return false
		}

		_, err := time.Parse("2006-01-02", str)
		return err == nil
	}))
}

// easyjson:json
type CreateResumeRequest struct {
	AboutMe                   string               `json:"about_me" valid:"stringlength(10|500),optional"`
	Specialization            string               `json:"specialization" valid:"required,stringlength(3|30)"`
	Profession                string               `json:"profession" valid:"required,stringlength(3|50)"`
	Education                 entity.EducationType `json:"education" valid:"required,in(secondary_school|incomplete_higher|higher|bachelor|master|phd)"`
	EducationalInstitution    string               `json:"educational_institution" valid:"required,stringlength(3|50)"`
	GraduationYear            string               `json:"graduation_year" valid:"required,customYearValidation"`
	Skills                    []string             `json:"skills" valid:"optional"`
	AdditionalSpecializations []string             `json:"additional_specializations" valid:"optional"`
	WorkExperiences           []WorkExperienceDTO  `json:"work_experiences" valid:"optional"`
}

// easyjson:json
type WorkExperienceDTO struct {
	EmployerName string `json:"employer_name" valid:"required,stringlength(2|50)"`
	Position     string `json:"position" valid:"required,stringlength(2|50)"`
	Duties       string `json:"duties" valid:"stringlength(5|250),optional"`
	Achievements string `json:"achievements" valid:"stringlength(0|1000),optional"`

	StartDate string `json:"start_date" valid:"required,date_iso"`
	EndDate   string `json:"end_date" valid:"date_iso,optional"`
	UntilNow  bool   `json:"until_now" valid:"optional"`
}

// easyjson:json
type ResumeResponse struct {
	ID                        int                      `json:"id"`
	ApplicantID               int                      `json:"applicant_id"`
	AboutMe                   string                   `json:"about_me,omitempty"`
	Specialization            string                   `json:"specialization,omitempty"`
	Profession                string                   `json:"profession,omitempty"`
	Education                 entity.EducationType     `json:"education,omitempty"`
	EducationalInstitution    string                   `json:"educational_institution,omitempty"`
	GraduationYear            string                   `json:"graduation_year,omitempty"`
	CreatedAt                 string                   `json:"created_at"`
	UpdatedAt                 string                   `json:"updated_at"`
	Skills                    []string                 `json:"skills"`
	AdditionalSpecializations []string                 `json:"additional_specializations"`
	WorkExperiences           []WorkExperienceResponse `json:"work_experiences"`
}

// easyjson:json
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
// easyjson:json
type UpdateResumeRequest struct {
	AboutMe                   string               `json:"about_me" valid:"stringlength(10|500), optional"`
	Specialization            string               `json:"specialization" valid:"required,stringlength(3|30)"`
	Profession                string               `json:"profession" valid:"required,stringlength(3|50)"`
	Education                 entity.EducationType `json:"education" valid:"required,in(secondary_school|incomplete_higher|higher|bachelor|master|phd)"`
	EducationalInstitution    string               `json:"educational_institution" valid:"required,stringlength(3|50)"`
	GraduationYear            string               `json:"graduation_year" valid:"required,customYearValidation"`
	Skills                    []string             `json:"skills" valid:"optional"`
	AdditionalSpecializations []string             `json:"additional_specializations" valid:"optional"`
	WorkExperiences           []WorkExperienceDTO  `json:"work_experiences" valid:"optional"`
}

// easyjson:json
type DeleteResumeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// New DTO for resume list
// easyjson:json
type ResumeShortResponse struct {
	ID             int                       `json:"id"`
	ApplicantID    int                       `json:"applicant_id,omitempty"` // Keep for backward compatibility
	Applicant      *ApplicantProfileResponse `json:"applicant"`              // Add applicant information
	Specialization string                    `json:"specialization"`
	Profession     string                    `json:"profession"`
	WorkExperience WorkExperienceShort       `json:"work_experiences"`
	CreatedAt      string                    `json:"created_at"`
	UpdatedAt      string                    `json:"updated_at"`
}

// Добавляем новое DTO для вывода списка резюме соискателя с навыками
// easyjson:json
type ResumeApplicantShortResponse struct {
	ID             int                       `json:"id"`
	ApplicantID    int                       `json:"applicant_id,omitempty"`
	Applicant      *ApplicantProfileResponse `json:"applicant"`
	Skills         []string                  `json:"skills"` // Добавлено поле навыков
	Specialization string                    `json:"specialization"`
	Profession     string                    `json:"profession"`
	WorkExperience WorkExperienceShort       `json:"work_experiences"`
	CreatedAt      string                    `json:"created_at"`
	UpdatedAt      string                    `json:"updated_at"`
}

// easyjson:json
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

type ResumeChatResponse struct {
	ID          int    `json:"id"`
	ApplicantID int    `json:"applicant_id"`
	Profession  string `json:"profession"`
}

// easyjson:json
type ResumeApplicantShortResponseList []ResumeApplicantShortResponse

// easyjson:json
type ResumeShortResponseList []ResumeShortResponse
