package dto

import (
	"time"

	"github.com/asaskevich/govalidator"
)

func init() {
	// Регистрация кастомных валидаторов для вакансий
	govalidator.CustomTypeTagMap.Set("vacancyWorkFormat", func(i interface{}, _ interface{}) bool {
		val, ok := i.(string)
		if !ok {
			return false
		}
		formats := map[string]bool{
			"office":    true,
			"hybrid":    true,
			"remote":    true,
			"traveling": true,
		}
		return formats[val]
	})

	govalidator.CustomTypeTagMap.Set("vacancyEmployment", func(i interface{}, _ interface{}) bool {
		val, ok := i.(string)
		if !ok {
			return false
		}
		types := map[string]bool{
			"full_time":  true,
			"part_time":  true,
			"contract":   true,
			"internship": true,
			"freelance":  true,
			"watch":      true,
		}
		return types[val]
	})

	govalidator.CustomTypeTagMap.Set("vacancySchedule", func(i interface{}, _ interface{}) bool {
		val, ok := i.(string)
		if !ok {
			return false
		}
		schedules := map[string]bool{
			"5/2":          true,
			"2/2":          true,
			"6/1":          true,
			"3/3":          true,
			"on_weekend":   true,
			"by_agreement": true,
		}
		return schedules[val]
	})

	govalidator.CustomTypeTagMap.Set("vacancyExperience", func(i interface{}, _ interface{}) bool {
		val, ok := i.(string)
		if !ok {
			return false
		}
		levels := map[string]bool{
			"no_experience": true,
			"1_3_years":     true,
			"3_6_years":     true,
			"6_plus_years":  true,
		}
		return levels[val]
	})
}

// VacancyShortResponse представляет сокращенную информацию о вакансии
// easyjson:json
type VacancyShortResponse struct {
	ID             int                      `json:"id"`
	Title          string                   `json:"title" valid:"required,stringlength(3|100)"`
	Employer       *EmployerProfileResponse `json:"employer" valid:"required"`
	Specialization string                   `json:"specialization" valid:"required,stringlength(3|50)"`
	WorkFormat     string                   `json:"work_format" valid:"required,vacancyWorkFormat"`
	Employment     string                   `json:"employment" valid:"required,vacancyEmployment"`
	WorkingHours   int                      `json:"working_hours" valid:"required,range(1|96)"`
	SalaryFrom     int                      `json:"salary_from" valid:"required,range(15000|1000000)"`
	SalaryTo       int                      `json:"salary_to" valid:"required,range(0|1000000),gtefield=SalaryFrom"`
	TaxesIncluded  bool                     `json:"taxes_included"`
	CreatedAt      string                   `json:"created_at" valid:"required"`
	UpdatedAt      string                   `json:"updated_at" valid:"required"`
	City           string                   `json:"city" valid:"required,stringlength(2|50)"`
	Responded      bool                     `json:"responded"`
	Liked          bool                     `json:"liked"`
}

// easyjson:json
type VacancyCreate struct {
	Title                string   `json:"title" valid:"required,stringlength(3|100)"`
	Specialization       string   `json:"specialization" valid:"required,stringlength(3|50)"`
	WorkFormat           string   `json:"work_format" valid:"required,vacancyWorkFormat"`
	Employment           string   `json:"employment" valid:"required,vacancyEmployment"`
	Schedule             string   `json:"schedule" valid:"required,vacancySchedule"`
	WorkingHours         int      `json:"working_hours" valid:"required,range(1|96)"`
	SalaryFrom           int      `json:"salary_from" valid:"required,range(15000|1000000)"`
	SalaryTo             int      `json:"salary_to" valid:"required,range(0|1000000),gtefield=SalaryFrom"`
	TaxesIncluded        bool     `json:"taxes_included"`
	Experience           string   `json:"experience" valid:"required,vacancyExperience"`
	City                 string   `json:"city" valid:"required,stringlength(2|50)"`
	Skills               []string `json:"skills" valid:"required,min=1,max=20,dive,stringlength(2|30)"`
	Description          string   `json:"description" valid:"required,stringlength(10|5000)"`
	Tasks                string   `json:"tasks" valid:"required,stringlength(10|2000)"`
	Requirements         string   `json:"requirements" valid:"required,stringlength(10|2000)"`
	OptionalRequirements string   `json:"optional_requirements" valid:"stringlength(0|2000)"`
}

// easyjson:json
type VacancyUpdate struct {
	Title                string   `json:"title" valid:"required,stringlength(3|100)"`
	Specialization       string   `json:"specialization" valid:"required,stringlength(3|50)"`
	WorkFormat           string   `json:"work_format" valid:"required,vacancyWorkFormat"`
	Employment           string   `json:"employment" valid:"required,vacancyEmployment"`
	Schedule             string   `json:"schedule" valid:"required,vacancySchedule"`
	WorkingHours         int      `json:"working_hours" valid:"required,range(1|96)"`
	SalaryFrom           int      `json:"salary_from" valid:"required,range(15000|1000000)"`
	SalaryTo             int      `json:"salary_to" valid:"required,range(0|1000000),gtefield=SalaryFrom"`
	TaxesIncluded        bool     `json:"taxes_included"`
	Experience           string   `json:"experience" valid:"required,vacancyExperience"`
	City                 string   `json:"city" valid:"required,stringlength(2|50)"`
	Skills               []string `json:"skills" valid:"required,min=1,max=20,dive,stringlength(2|30)"`
	Description          string   `json:"description" valid:"required,stringlength(10|5000)"`
	Tasks                string   `json:"tasks" valid:"required,stringlength(10|2000)"`
	Requirements         string   `json:"requirements" valid:"required,stringlength(10|2000)"`
	OptionalRequirements string   `json:"optional_requirements" valid:"stringlength(0|2000)"`
}

// easyjson:json
type VacancyResponse struct {
	ID                   int      `json:"id" valid:"required"`
	EmployerID           int      `json:"employer_id" valid:"required"`
	Title                string   `json:"title" valid:"required,stringlength(3|100)"`
	Specialization       string   `json:"specialization" valid:"required,stringlength(3|50)"`
	WorkFormat           string   `json:"work_format" valid:"required,vacancyWorkFormat"`
	Employment           string   `json:"employment" valid:"required,vacancyEmployment"`
	Schedule             string   `json:"schedule" valid:"required,vacancySchedule"`
	WorkingHours         int      `json:"working_hours" valid:"required,range(1|96)"`
	SalaryFrom           int      `json:"salary_from" valid:"required,range(15000|1000000)"`
	SalaryTo             int      `json:"salary_to" valid:"required,range(0|1000000),gtefield=SalaryFrom"`
	TaxesIncluded        bool     `json:"taxes_included"`
	Experience           string   `json:"experience" valid:"required,vacancyExperience"`
	City                 string   `json:"city" valid:"required,stringlength(2|50)"`
	Skills               []string `json:"skills" valid:"required,min=1,max=20,dive,stringlength(2|30)"`
	Description          string   `json:"description" valid:"required,stringlength(10|5000)"`
	Tasks                string   `json:"tasks" valid:"required,stringlength(10|2000)"`
	Requirements         string   `json:"requirements" valid:"required,stringlength(10|2000)"`
	OptionalRequirements string   `json:"optional_requirements" valid:"stringlength(0|2000)"`
	CreatedAt            string   `json:"created_at" valid:"required"`
	UpdatedAt            string   `json:"updated_at" valid:"required"`
	Responded            bool     `json:"responded"`
	Liked                bool     `json:"liked"`
}

// easyjson:json
type VacancyResponsed struct {
	ID          int       `json:"id" valid:"required"`
	VacancyID   int       `json:"vacancy_id" valid:"required"`
	ApplicantID int       `json:"applicant_id" valid:"required"`
	ResumeID    []int     `json:"resume_id,omitempty" valid:"optional"`
	AppliedAt   time.Time `json:"applied_at" valid:"required"`
}

// easyjson:json
type DeleteVacancy struct {
	Success bool   `json:"success" valid:"required"`
	Message string `json:"message" valid:"required,stringlength(1|100)"`
}

// easyjson:json
type ApplyToVacancyRequest struct {
	ResumeID int `json:"resume_id,omitempty" valid:"optional"`
}

// SearchBySpecializationsRequest для поиска вакансий по специализациям
// easyjson:json
type SearchBySpecializationsRequest struct {
	Specializations []string `json:"specializations" valid:"required,min=1,max=10,dive,stringlength(3|50)"`
}

// easyjson:json
type SearchByQueryAndSpecializationsRequest struct {
	Specializations []string `json:"specializations" valid:"required,min=1,max=10,dive,stringlength(3|50)"`
}

// easyjson:json
type VacancyShortResponseList []VacancyShortResponse
