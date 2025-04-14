package entity

import (
	"fmt"
	"time"
	"unicode/utf8"
)

type Vacancy struct {
	ID                      int                       `json:"id"`
	Title                   string                    `json:"title"`
	IsActive                bool                      `json:"is_active"`
	EmployerID              int                       `json:"employer_id"`
	SpecializationID        int                       `json:"specialization_id"`
	WorkFormat              string                    `json:"work_format"`
	Employment              string                    `json:"employment"`
	Schedule                string                    `json:"schedule"`
	WorkingHours            int                       `json:"working_hours"`
	SalaryFrom              int                       `json:"salary_from"`
	SalaryTo                int                       `json:"salary_to"`
	TaxesIncluded           string                    `json:"taxes_included"`
	Experience              int                       `json:"experience"`
	Description             string                    `json:"description"`
	Tasks                   string                    `json:"tasks"`
	Requirements            string                    `json:"requirements"`
	OptionalRequirements    string                    `json:"optional_requirements"`
	CreatedAt               time.Time                 `json:"created_at"`
	UpdatedAt               time.Time                 `json:"updated_at"`
	Skills                  []Skill                   `json:"-"`
	City                    []City                    `json:"-"`
	SupplementaryConditions []SupplementaryConditions `json:"-"`
<<<<<<< HEAD
<<<<<<< HEAD
	Responded               bool                      `json:"responded"`
=======
>>>>>>> c773955 (Made vacansies usecases and handlers)
=======
	Responded               bool                      `json:"responded"`
>>>>>>> a6396a4 (Fix mistakes)
}

// VacancyShort представляет сокращенную информацию о вакансии
type VacancyShort struct {
	ID             int32     `json:"id"`
	Title          string    `json:"title"`
	Employer       Employer  `json:"employer"`
	Specialization string    `json:"specialization"`
	City           string    `json:"city"`
	WorkFormat     string    `json:"work_format"`
	Employment     string    `json:"employment"`
	WorkingHours   int32     `json:"working_hours"`
	SalaryFrom     int32     `json:"salary_from"`
	SalaryTo       int32     `json:"salary_to"`
	TaxesIncluded  bool      `json:"taxes_included"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// VacancyCreate представляет данные для создания вакансии
type VacancyCreate struct {
	Title                string   `json:"title"`
	Specialization       string   `json:"specialization"`
	WorkFormat           string   `json:"work_format"`
	Employment           string   `json:"employment"`
	Schedule             string   `json:"schedule"`
	WorkingHours         int32    `json:"working_hours"`
	SalaryFrom           int32    `json:"salary_from"`
	SalaryTo             int32    `json:"salary_to"`
	TaxesIncluded        bool     `json:"taxes_included"`
	Experience           string   `json:"experience"`
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
	Specialization       string    `json:"specialization"`
	WorkFormat           string    `json:"work_format"`
	Employment           string    `json:"employment"`
	Schedule             string    `json:"schedule"`
	WorkingHours         int32     `json:"working_hours"`
	SalaryFrom           int32     `json:"salary_from"`
	SalaryTo             int32     `json:"salary_to"`
	TaxesIncluded        bool      `json:"taxes_included"`
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

type VacancyResponses struct {
	ID          int       `json:"id"`
	VacancyID   int       `json:"vacancy_id"`
	ApplicantID int       `json:applicant_id`
	AppliedAt   time.Time `json:applied_at`
}

type VacancyLike struct {
	ID          int       `json:"id"`
	VacancyID   int       `json:"vacancy_id"`
	ApplicantID int       `json:applicant_id`
	LikedAt     time.Time `json:liked_at`
}

type SupplementaryConditions struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
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
