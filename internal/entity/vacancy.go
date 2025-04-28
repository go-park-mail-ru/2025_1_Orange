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
	TaxesIncluded           bool                      `json:"taxes_included"`
	Experience              string                    `json:"experience"`
	Description             string                    `json:"description"`
	Tasks                   string                    `json:"tasks"`
	Requirements            string                    `json:"requirements"`
	OptionalRequirements    string                    `json:"optional_requirements"`
	CreatedAt               time.Time                 `json:"created_at"`
	UpdatedAt               time.Time                 `json:"updated_at"`
	Skills                  []Skill                   `json:"-"`
	City                    string                    `json:"city"`
	SupplementaryConditions []SupplementaryConditions `json:"-"`
	Responded               bool                      `json:"responded"`
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
	Responded      bool      `json:"responded"`
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
	City                 string   `json:"city"`
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
	WorkingHours         int       `json:"working_hours"`
	SalaryFrom           int       `json:"salary_from"`
	SalaryTo             int       `json:"salary_to"`
	TaxesIncluded        bool      `json:"taxes_included"`
	Experience           string    `json:"experience"`
	City                 string    `json:"city"`
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
	ApplicantID int       `json:"applicant_id"`
	AppliedAt   time.Time `json:"applied_at"`
}

type VacancyLike struct {
	ID          int       `json:"id"`
	VacancyID   int       `json:"vacancy_id"`
	ApplicantID int       `json:"applicant_id"`
	LikedAt     time.Time `json:"liked_at"`
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
	if utf8.RuneCountInString(v.Title) > 128 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("название вакансии не может быть длиннее 128 символов"),
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

	if v.OptionalRequirements != "" && utf8.RuneCountInString(v.OptionalRequirements) > 2000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("дополнительные требования вакансии не могут быть длиннее 2000 символов"),
		)
	}

	// Проверка зарплатного диапазона
	if v.SalaryFrom < 0 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("минимальная зарплата не может быть отрицательной"),
		)
	}

	if v.SalaryTo < 0 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("максимальная зарплата не может быть отрицательной"),
		)
	}

	if v.SalaryFrom > 0 && v.SalaryTo > 0 && v.SalaryFrom > v.SalaryTo {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("минимальная зарплата не может быть больше максимальной"),
		)
	}

	// Проверка формата работы
	validWorkFormats := map[string]bool{
		"office":    true,
		"remote":    true,
		"hybrid":    true,
		"traveling": true,
	}
	if v.WorkFormat != "" && !validWorkFormats[v.WorkFormat] {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("недопустимый формат работы: %s", v.WorkFormat),
		)
	}

	// Проверка типа занятости
	validEmployment := map[string]bool{
		"full_time":  true,
		"part_time":  true,
		"contract":   true,
		"internship": true,
		"freelance":  true,
		"watch":      true,
	}
	if v.Employment != "" && !validEmployment[v.Employment] {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("недопустимый тип занятости: %s", v.Employment),
		)
	}

	// Проверка графика работы
	validSchedules := map[string]bool{
		"5/2":          true,
		"2/2":          true,
		"6/1":          true,
		"3/3":          true,
		"on_weekend":   true,
		"by_agreement": true,
	}
	if v.Schedule != "" && !validSchedules[v.Schedule] {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("недопустимый график работы: %s", v.Schedule),
		)
	}

	// Проверка количества рабочих часов
	if v.WorkingHours <= 0 || v.WorkingHours > 168 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("количество рабочих часов должно быть больше 0 и не более 168"),
		)
	}

	// Проверка опыта работы
	validExperience := map[string]bool{
		"no_matter":     true,
		"no_experience": true,
		"1_3_years":     true,
		"3_6_years":     true,
		"6_plus_years":  true,
	}
	if v.Experience != "" && !validExperience[v.Experience] {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("недопустимый уровень опыта: %s", v.Experience),
		)
	}

	// Проверка работодателя
	if v.EmployerID <= 0 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("необходимо указать существующего работодателя"),
		)
	}

	// Проверка города
	if v.City == "" {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("необходимо указать город"),
		)
	}

	return nil
}
