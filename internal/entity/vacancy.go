package entity

import (
	"fmt"
	"time"
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

type VacancyChatInfo struct {
	VacancyID  int
	ResumeID   int
	EmployerID int
}

func (v *Vacancy) Validate() error {
	if v.Title == "" || len([]rune(v.Title)) < 3 || len([]rune(v.Title)) > 50 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("название вакансии должно быть от 3 до 50 символов"),
		)
	}

	if v.EmployerID <= 0 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректный ID работодателя"),
		)
	}

	if v.SpecializationID <= 0 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректный ID специализации"),
		)
	}

	validWorkFormats := map[string]bool{
		"office":    true,
		"hybrid":    true,
		"remote":    true,
		"traveling": true,
	}
	if !validWorkFormats[v.WorkFormat] {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректный формат работы"),
		)
	}

	validEmployment := map[string]bool{
		"full_time":  true,
		"part_time":  true,
		"contract":   true,
		"internship": true,
		"freelance":  true,
		"watch":      true,
	}
	if !validEmployment[v.Employment] {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректный тип занятости"),
		)
	}

	validSchedules := map[string]bool{
		"5/2":          true,
		"2/2":          true,
		"6/1":          true,
		"3/3":          true,
		"on_weekend":   true,
		"by_agreement": true,
	}
	if !validSchedules[v.Schedule] {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректный график работы"),
		)
	}

	if v.WorkingHours < 1 || v.WorkingHours > 96 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректное количество рабочих часов"),
		)
	}

	if v.SalaryFrom < 15000 || v.SalaryFrom > 1000000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректная минимальная зарплата"),
		)
	}

	if v.SalaryTo < 0 || v.SalaryTo > 1000000 || v.SalaryTo < v.SalaryFrom {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректная максимальная зарплата"),
		)
	}

	validExperience := map[string]bool{
		"no_experience": true,
		"1_3_years":     true,
		"3_6_years":     true,
		"6_plus_years":  true,
	}
	if !validExperience[v.Experience] {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректный требуемый опыт"),
		)
	}

	if v.Description == "" || len([]rune(v.Description)) < 10 || len([]rune(v.Description)) > 500 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("описание вакансии должно быть от 10 до 500 символов"),
		)
	}

	if v.Tasks != "" && (len([]rune(v.Tasks)) < 10 || len([]rune(v.Tasks)) > 500) {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("описание задач должно быть от 10 до 500 символов %d", len([]rune(v.Tasks))),
		)
	}

	if v.Requirements != "" && (len([]rune(v.Requirements)) < 10 || len([]rune(v.Requirements)) > 500) {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("описание требований должно быть от 10 до 500 символов"),
		)
	}

	if v.OptionalRequirements != "" && (len([]rune(v.OptionalRequirements)) < 10 || len([]rune(v.OptionalRequirements)) > 500) {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("описание дополнительных требований должно быть от 10 до 500 символов"),
		)
	}

	if v.City == "" || len([]rune(v.City)) < 3 || len([]rune(v.City)) > 50 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("город должен быть от 3 до 50 символов"),
		)
	}

	return nil
}

// Остальные структуры остаются без изменений
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
	Liked          bool      `json:"liked"`
}

type VacancyCreate struct {
	Title                string   `form:"title"`
	Specialization       string   `form:"specialization"`
	City                 string   `form:"city"`
	Employment           string   `form:"employment"`
	Schedule             string   `form:"schedule"`
	WorkingHours         int      `form:"working_hours"`
	WorkFormat           string   `form:"work_format"`
	SalaryFrom           int      `form:"salary_from"`
	SalaryTo             int      `form:"salary_to"`
	TaxesIncluded        string   `form:"taxes_included"`
	Experience           string   `form:"experience"`
	Description          string   `form:"description"`
	Tasks                string   `form:"tasks"`
	Requirements         string   `form:"requirements"`
	OptionalRequirements string   `form:"optional_requirements"`
	Skills               []string `form:"skills"`
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
	ResumeID    int       `json:"resume_id"`
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
