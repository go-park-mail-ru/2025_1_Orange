package entity

import (
	"fmt"
	"time"
	"unicode/utf8"
)

// EducationType represents the education level enum
type EducationType string

const (
	SecondarySchool  EducationType = "secondary_school"
	IncompleteHigher EducationType = "incomplete_higher"
	Higher           EducationType = "higher"
	Bachelor         EducationType = "bachelor"
	Master           EducationType = "master"
	PhD              EducationType = "phd"
)

type Resume struct {
	ID                        int              `json:"id"`
	ApplicantID               int              `json:"applicant_id"`
	AboutMe                   string           `json:"about_me,omitempty"`
	SpecializationID          int              `json:"specialization_id,omitempty"`
	Education                 EducationType    `json:"education,omitempty"`
	EducationalInstitution    string           `json:"educational_institution,omitempty"`
	GraduationYear            time.Time        `json:"graduation_year,omitempty"`
	CreatedAt                 time.Time        `json:"created_at"`
	UpdatedAt                 time.Time        `json:"updated_at"`
	Skills                    []int            `json:"-"`
	AdditionalSpecializations []int            `json:"-"`
	WorkExperiences           []WorkExperience `json:"-"`
}

type WorkExperience struct {
	ID           int       `json:"id"`
	ResumeID     int       `json:"resume_id"`
	EmployerName string    `json:"employer_name"`
	Position     string    `json:"position"`
	Duties       string    `json:"duties,omitempty"`
	Achievements string    `json:"achievements,omitempty"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date,omitempty"`
	UntilNow     bool      `json:"until_now"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (r *Resume) Validate() error {
	if r.ApplicantID <= 0 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректный ID соискателя"),
		)
	}

	if r.AboutMe != "" && utf8.RuneCountInString(r.AboutMe) > 2000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("информация о себе не может быть длиннее 2000 символов"),
		)
	}

	if r.SpecializationID < 0 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректный ID специализации"),
		)
	}

	return nil
}

func (w *WorkExperience) Validate() error {
	if utf8.RuneCountInString(w.EmployerName) > 64 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("название работодателя не может быть длиннее 64 символов"),
		)
	}

	if utf8.RuneCountInString(w.Position) > 64 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("должность не может быть длиннее 64 символов"),
		)
	}

	if w.Duties != "" && utf8.RuneCountInString(w.Duties) > 1000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("обязанности не могут быть длиннее 1000 символов"),
		)
	}

	if w.Achievements != "" && utf8.RuneCountInString(w.Achievements) > 1000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("достижения не могут быть длиннее 1000 символов"),
		)
	}

	if !w.UntilNow && !w.EndDate.IsZero() && w.EndDate.Before(w.StartDate) {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("дата окончания работы не может быть раньше даты начала"),
		)
	}

	return nil
}
