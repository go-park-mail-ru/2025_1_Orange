package dto

import (
	"time"
)

// VacancyShort представляет сокращенную информацию о вакансии
type VacancyShortResponse struct {
	ID             int    `json:"id" validate:"required,min=1"`
	Title          string `json:"title" validate:"required,min=10,max=100"`
	EmployerID     int    `json:"employer_id" validate:"required,min=1"`
	Specialization string `json:"specialization" validate:"required,min=3,max=50"`
	WorkFormat     string `json:"work_format" validate:"required,oneof=remote office hybrid traveling"`
	Employment     string `json:"employment" validate:"required,oneof=full_time part_time contract freelance internship watch"`
	WorkingHours   int    `json:"working_hours" validate:"required,min=1,max=168"`
	SalaryFrom     int    `json:"salary_from" validate:"required,min=0"`
	SalaryTo       int    `json:"salary_to" validate:"required,min=0,gtfield=SalaryFrom"`
	TaxesIncluded  bool   `json:"taxes_included"`
	CreatedAt      string `json:"created_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	UpdatedAt      string `json:"updated_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	City           string `json:"city" validate:"required,min=2,max=50"`
	Responded      bool   `json:"responded"`
}

type VacancyCreate struct {
	Title                string   `json:"title" validate:"required,min=3,max=50,validTitle"`
	Specialization       string   `json:"specialization" validate:"required,min=3,max=50,validText"`
	City                 string   `json:"city" validate:"required,min=3,max=50,validCity"`
	Employment           string   `json:"employment" validate:"required,oneof=full_time part_time contract internship freelance watch"`
	Schedule             string   `json:"schedule" validate:"required,oneof=5/2 2/2 6/1 3/3 on_weekend by_agreement"`
	WorkingHours         int      `json:"working_hours" validate:"required,min=1,max=96"`
	WorkFormat           string   `json:"work_format" validate:"required,oneof=office hybrid remote traveling"`
	SalaryFrom           int      `json:"salary_from" validate:"required,min=15000,max=1000000"`
	SalaryTo             int      `json:"salary_to" validate:"required,min=0,max=1000000,gtefield=SalaryFrom"`
	TaxesIncluded        bool     `json:"taxes_included"`
	Experience           string   `json:"experience" validate:"required,oneof=no_experience 1_3_years 3_6_years 6_plus_years"`
	Description          string   `json:"description" validate:"required,min=10,max=500,validText"`
	Tasks                string   `json:"tasks" validate:"min=10,max=500,validText"`
	Requirements         string   `json:"requirements" validate:"min=10,max=500,validText"`
	OptionalRequirements string   `json:"optional_requirements" validate:"min=10,max=500,validText"`
	Skills               []string `json:"skills" validate:"dive,min=2,max=30,validText"`
}

type VacancyUpdate struct {
	Title                string   `json:"title" validate:"omitempty,min=10,max=100"`
	Specialization       string   `json:"specialization" validate:"required,min=3,max=50,validText"`
	WorkFormat           string   `json:"work_format" validate:"required,oneof=office hybrid remote traveling"`
	Employment           string   `json:"employment" validate:"required,oneof=full_time part_time contract internship freelance watch"`
	Schedule             string   `json:"schedule" validate:"required,oneof=5/2 2/2 6/1 3/3 on_weekend by_agreement"`
	WorkingHours         int      `json:"working_hours" validate:"required,min=1,max=96"`
	SalaryFrom           int      `json:"salary_from" validate:"required,min=15000,max=1000000"`
	SalaryTo             int      `json:"salary_to" validate:"required,min=0,max=1000000,gtefield=SalaryFrom"`
	TaxesIncluded        bool     `json:"taxes_included"`
	Experience           string   `json:"experience" validate:"required,oneof=no_experience 1_3_years 3_6_years 6_plus_years"`
	City                 string   `json:"city" validate:"required,min=3,max=50,validCity"`
	Skills               []string `json:"skills" validate:"dive,min=2,max=30,validText"`
	Description          string   `json:"description" validate:"required,min=10,max=500,validText"`
	Tasks                string   `json:"tasks" validate:"min=10,max=500,validText"`
	Requirements         string   `json:"requirements" validate:"min=10,max=500,validText"`
	OptionalRequirements string   `json:"optional_requirements" validate:"min=10,max=500,validText"`
}

type VacancyResponse struct {
	ID                   int      `json:"id" validate:"required,min=1"`
	EmployerID           int      `json:"employer_id" validate:"required,min=1"`
	Title                string   `json:"title" validate:"required,min=10,max=100"`
	Specialization       string   `json:"specialization" validate:"required,min=3,max=50"`
	WorkFormat           string   `json:"work_format" validate:"required,oneof=remote office hybrid traveling"`
	Employment           string   `json:"employment" validate:"required,oneof=full_time part_time contract freelance internship watch"`
	Schedule             string   `json:"schedule"`
	WorkingHours         int      `json:"working_hours" validate:"required,min=1,max=168"`
	SalaryFrom           int      `json:"salary_from" validate:"required,min=0"`
	SalaryTo             int      `json:"salary_to" validate:"required,min=0,gtfield=SalaryFrom"`
	TaxesIncluded        bool     `json:"taxes_included"`
	Experience           string   `json:"experience"`
	City                 string   `json:"city" validate:"required,min=2,max=50"`
	Skills               []string `json:"skills" validate:"required,min=1,max=10,dive,min=2,max=30"`
	Description          string   `json:"description" validate:"required,min=10,max=5000"`
	Tasks                string   `json:"tasks" validate:"required,min=10,max=2000"`
	Requirements         string   `json:"requirements" validate:"required,min=10,max=2000"`
	OptionalRequirements string   `json:"optional_requirements" validate:"max=2000"`
	CreatedAt            string   `json:"created_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	UpdatedAt            string   `json:"updated_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Responded            bool     `json:"responded"`
}

type VacancyResponsed struct {
	ID          int       `json:"id" validate:"required,min=1"`
	VacancyID   int       `json:"vacancy_id" validate:"required,min=1"`
	ApplicantID int       `json:"applicant_id" validate:"required,min=1"`
	ResumeID    []int     `json:"resume_id,omitempty" validate:"omitempty,dive,min=1"`
	AppliedAt   time.Time `json:"applied_at" validate:"required"`
}

type DeleteVacancy struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ApplyToVacancyRequest struct {
	ResumeID int `json:"resume_id,omitempty"`
}
