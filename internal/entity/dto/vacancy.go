package dto

import "time"

type Vacancy struct {
	ID                   int       `json:"id"`
	Title                string    `json:"title`
	IsActive             bool      `json:"is_active`
	EmployerID           int       `json:"employer_id"`
	SpecializationID     int       `json:"specialization_id,omitempty"`
	WorkFormat           string    `json:"work_format,omitempty"`
	Employment           string    `json:"employment,omitempty"`
	Schedule             string    `json:"schedule,omitempty"`
	WorkingHours         string    `json:"working_hours,omitempty"`
	SalaryFrom           int       `json:"salary_from,omitempty"`
	SalaryTo             int       `json:"salary_to,omitempty"`
	TaxesIncluded        bool      `json:"taxes_included,omitempty"`
	Experience           string    `json:"experience,omitempty"`
	Description          string    `json:"description"`
	Tasks                string    `json:"tasks,omitempty"`
	Requirements         string    `json:"requirements,omitempty"`
	OptionalRequirements string    `json:"optional_requirements,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

const (
	UserTypeApplicant = "applicant"
	UserTypeEmployer  = "employer"
)

type UserFromSession struct {
	ID       int
	UserType string
}
