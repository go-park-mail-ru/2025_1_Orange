package entity

import (
	"time"
	_ "unicode/utf8"

	"github.com/asaskevich/govalidator"
)

type Vacancy struct {
	ID                      int                       `json:"id"`
	Title                   string                    `json:"title" valid:"required,length(3|50),matches(^[а-яА-Яa-zA-Z0-9\\s\\.,#+\\-]+$)"`
	IsActive                bool                      `json:"is_active"`
	EmployerID              int                       `json:"employer_id" valid:"required"`
	SpecializationID        int                       `json:"specialization_id" valid:"required"`
	WorkFormat              string                    `json:"work_format" valid:"required,in(office|hybrid|remote|traveling)"`
	Employment              string                    `json:"employment" valid:"required,in(full_time|part_time|contract|internship|freelance|watch)"`
	Schedule                string                    `json:"schedule" valid:"required,in(5/2|2/2|6/1|3/3|on_weekend|by_agreement)"`
	WorkingHours            int                       `json:"working_hours" valid:"required,range(1|96)"`
	SalaryFrom              int                       `json:"salary_from" valid:"required,range(15000|1000000)"`
	SalaryTo                int                       `json:"salary_to" valid:"required,range(0|1000000)"`
	TaxesIncluded           bool                      `json:"taxes_included"`
	Experience              string                    `json:"experience" valid:"required,in(no_experience|1_3_years|3_6_years|6_plus_years)"`
	Description             string                    `json:"description" valid:"required,length(10|500),matches(^[а-яА-Яa-zA-Z0-9\\s\\.,#+\\-]+$)"`
	Tasks                   string                    `json:"tasks" valid:"length(10|500),matches(^[а-яА-Яa-zA-Z0-9\\s\\.,#+\\-]+$)"`
	Requirements            string                    `json:"requirements" valid:"length(10|500),matches(^[а-яА-Яa-zA-Z0-9\\s\\.,]+$)"`
	OptionalRequirements    string                    `json:"optional_requirements" valid:"length(10|500),matches(^[а-яА-Яa-zA-Z0-9\\s\\.,#+\\-]+$)"`
	CreatedAt               time.Time                 `json:"created_at"`
	UpdatedAt               time.Time                 `json:"updated_at"`
	Skills                  []Skill                   `json:"-" valid:"dive"`
	City                    string                    `json:"city" valid:"required,length(3|50),matches(^[а-яА-Яa-zA-Z0-9\\s\\.,]+$)"`
	SupplementaryConditions []SupplementaryConditions `json:"-"`
	Responded               bool                      `json:"responded"`
}

func (v *Vacancy) Validate() (bool, error) {
	return govalidator.ValidateStruct(v)
}

type VacancyShort struct {
	ID             int32     `json:"id" validate:"required,min=1"`
	Title          string    `json:"title" validate:"required,min=10,max=100"`
	Employer       Employer  `json:"employer" validate:"required"`
	Specialization string    `json:"specialization" validate:"required,min=3,max=50"`
	City           string    `json:"city" validate:"required,min=2,max=50"`
	WorkFormat     string    `json:"work_format" validate:"required,oneof=remote office hybrid traveling"`
	Employment     string    `json:"employment" validate:"required,oneof=full_time part_time contract freelance internship watch"`
	WorkingHours   int32     `json:"working_hours" validate:"required,min=1,max=168"`
	SalaryFrom     int32     `json:"salary_from" validate:"required,min=0"`
	SalaryTo       int32     `json:"salary_to" validate:"required,min=0,gtfield=SalaryFrom"`
	TaxesIncluded  bool      `json:"taxes_included"`
	CreatedAt      time.Time `json:"created_at" validate:"required"`
	UpdatedAt      time.Time `json:"updated_at" validate:"required"`
	Responded      bool      `json:"responded"`
	Liked          bool      `json:"liked"`
}

type VacancyCreate struct {
	Title                string   `form:"title" validate:"required,min=3,max=50,validTitle"`
	Specialization       string   `form:"specialization" validate:"required,min=3,max=50,validText"`
	City                 string   `form:"city" validate:"required,min=3,max=50,validCity"`
	Employment           string   `form:"employment" validate:"required,oneof=full_time part_time contract internship freelance watch"`
	Schedule             string   `form:"schedule" validate:"required,oneof=5/2 2/2 6/1 3/3 on_weekend by_agreement"`
	WorkingHours         int      `form:"working_hours" validate:"required,min=1,max=96"`
	WorkFormat           string   `form:"work_format" validate:"required,oneof=office hybrid remote traveling"`
	SalaryFrom           int      `form:"salary_from" validate:"required,min=15000,max=1000000"`
	SalaryTo             int      `form:"salary_to" validate:"required,min=0,max=1000000,gtefield=SalaryFrom"`
	TaxesIncluded        string   `form:"taxes_included" validate:"required,oneof=true false"`
	Experience           string   `form:"experience" validate:"required,oneof=no_experience 1_3_years 3_6_years 6_plus_years"`
	Description          string   `form:"description" validate:"required,min=10,max=500,validText"`
	Tasks                string   `form:"tasks" validate:"min=10,max=500,validText"`
	Requirements         string   `form:"requirements" validate:"min=10,max=500,validText"`
	OptionalRequirements string   `form:"optional_requirements" validate:"min=10,max=500,validText"`
	Skills               []string `validate:"dive,min=2,max=30,validText"`
}

type VacancyResponse struct {
	ID                   int       `json:"id" validate:"required,min=1"`
	EmployerID           int       `json:"employer_id" validate:"required,min=1"`
	Title                string    `json:"title" validate:"required,min=10,max=100"`
	Specialization       string    `json:"specialization" validate:"required,min=3,max=50"`
	WorkFormat           string    `json:"work_format" validate:"required,oneof=remote office hybrid traveling"`
	Employment           string    `json:"employment" validate:"required,oneof=full_time part_time contract freelance internship watch"`
	Schedule             string    `json:"schedule"`
	WorkingHours         int       `json:"working_hours" validate:"required,min=1,max=168"`
	SalaryFrom           int       `json:"salary_from" validate:"required,min=0"`
	SalaryTo             int       `json:"salary_to" validate:"required,min=0,gtfield=SalaryFrom"`
	TaxesIncluded        bool      `json:"taxes_included"`
	Experience           string    `json:"experience"`
	City                 string    `json:"city" validate:"required,min=2,max=50"`
	Skills               []string  `json:"skills" validate:"required,min=1,max=10,dive,min=2,max=30"`
	Description          string    `json:"description" validate:"required,min=10,max=5000"`
	Tasks                string    `json:"tasks" validate:"required,min=20,max=2000"`
	Requirements         string    `json:"requirements" validate:"required,min=10,max=2000"`
	OptionalRequirements string    `json:"optional_requirements" validate:"max=2000"`
	CreatedAt            time.Time `json:"created_at" validate:"required"`
	UpdatedAt            time.Time `json:"updated_at" validate:"required"`
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
<<<<<<< HEAD
=======

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
>>>>>>> 327c6813add79596443fff2ebd31e7419339cd7b
