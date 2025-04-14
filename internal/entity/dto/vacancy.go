package dto

import (
	"time"
)

// VacancyShort представляет сокращенную информацию о вакансии
type VacancyShortResponse struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	EmployerID     int    `json:"employer_id"`
	Specialization string `json:"specialization"`
	WorkFormat     string `json:"work_format"`
	Employment     string `json:"employment"`
	WorkingHours   int    `json:"working_hours"`
	SalaryFrom     int    `json:"salary_from"`
	SalaryTo       int    `json:"salary_to"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type VacancyCreate struct {
	Title                string   `json:"title"`
	SpecializationID     int      `json:"specialization_id"`
	WorkFormat           string   `json:"work_format"`
	Employment           string   `json:"employment"`
	Schedule             string   `json:"schedule"`
	WorkingHours         int      `json:"working_hours"`
	SalaryFrom           int      `json:"salary_from"`
	SalaryTo             int      `json:"salary_to"`
	TaxesIncluded        string   `json:"taxes_included"`
	Experience           int      `json:"experience"`
	City                 []string `json:"city"`
	Skills               []string `json:"skills"`
	Description          string   `json:"description"`
	Tasks                string   `json:"tasks"`
	Requirements         string   `json:"requirements"`
	OptionalRequirements string   `json:"optional_requirements"`
}

type VacancyUpdate struct {
	Title                string   `json:"title"`
	SpecializationID     int      `json:"specialization_id"`
	WorkFormat           string   `json:"work_format"`
	Employment           string   `json:"employment"`
	Schedule             string   `json:"schedule"`
	WorkingHours         int      `json:"working_hours"`
	SalaryFrom           int      `json:"salary_from"`
	SalaryTo             int      `json:"salary_to"`
	TaxesIncluded        string   `json:"taxes_included"`
	Experience           int      `json:"experience"`
	City                 []string `json:"city"`
	Skills               []string `json:"skills"`
	Description          string   `json:"description"`
	Tasks                string   `json:"tasks"`
	Requirements         string   `json:"requirements"`
	OptionalRequirements string   `json:"optional_requirements"`
}

type VacancyResponse struct {
	ID                   int       `json:"id"`
	EmployerID           int       `json:"employer_id"`
	Title                string    `json:"title"`
	SpecializationID     string    `json:"specialization_id"`
	WorkFormat           string    `json:"work_format"`
	Employment           string    `json:"employment"`
	Schedule             string    `json:"schedule"`
	WorkingHours         int       `json:"working_hours"`
	SalaryFrom           int       `json:"salary_from"`
	SalaryTo             int       `json:"salary_to"`
	TaxesIncluded        string    `json:"taxes_included"`
	Experience           string    `json:"experience"`
	City                 []string  `json:"city"`
	Skills               []string  `json:"skills"`
	Description          string    `json:"description"`
	Tasks                string    `json:"tasks"`
	Requirements         string    `json:"requirements"`
	OptionalRequirements string    `json:"optional_requirements"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type VacancyResponsed struct {
	ID          int       `json:"id"`
	VacancyID   int       `json:"vacancy_id"`
	ApplicantID int       `json:"applicant_id"`
	ResumeID    []int     `json:"resume_id,omitempty"`
	AppliedAt   time.Time `json:"applied_at"`
}

type DeleteVacancy struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
<<<<<<< HEAD

type ApplyToVacancyRequest struct {
	ResumeID int `json:"resume_id"`
}
=======
>>>>>>> a6396a4 (Fix mistakes)
