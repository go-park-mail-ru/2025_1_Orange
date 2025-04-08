package entity

import (
	"fmt"
	"time"
	"unicode/utf8"
)

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

func (v *Vacancy) Validate() error {
	// Проверка обязательных полей
	if v.Title == "" {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("название вакансии обязательно"),
		)
	}

	if v.Description == "" {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("описание вакансии обязательно"),
		)
	}

	// Проверка длины полей
	if utf8.RuneCountInString(v.Title) > 64 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("название вакансии не может быть длиннее 64 символов"),
		)
	}

	if utf8.RuneCountInString(v.Description) > 5000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("описание вакансии не может быть длиннее 5000 символов"),
		)
	}

	if v.Tasks != "" && utf8.RuneCountInString(v.Tasks) > 2000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("задачи вакансии не могут быть длиннее 2000 символов"),
		)
	}

	if v.Requirements != "" && utf8.RuneCountInString(v.Requirements) > 2000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("требования вакансии не могут быть длиннее 2000 символов"),
		)
	}

	// Проверка зарплатного диапазона
	if v.SalaryFrom > 0 && v.SalaryTo > 0 && v.SalaryFrom > v.SalaryTo {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("минимальная зарплата не может быть больше максимальной"),
		)
	}

	validWorkFormats := map[string]bool{"office": true, "remote": true, "hybrid": true}
	if v.WorkFormat != "" && !validWorkFormats[v.WorkFormat] {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("недопустимый формат работы"),
		)
	}

	validEmployment := map[string]bool{"full": true, "part": true, "project": true}
	if v.Employment != "" && !validEmployment[v.Employment] {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("недопустимый тип занятости"),
		)
	}
	return nil
}

const (
	UserTypeApplicant = "applicant"
	UserTypeEmployer  = "employer"
)

type UserFromSession struct {
	ID       uint64
	UserType string
}
