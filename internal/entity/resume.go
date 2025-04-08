package entity

import (
	"fmt"
	"time"
	"unicode/utf8"
)

type Resume struct {
	ID                        int              `json:"id"`
	ApplicantID               int              `json:"applicant_id"`
	AboutMe                   string           `json:"about_me"`
	SpecializationID          int              `json:"specialization_id"`
	Education                 int              `json:"education"`
	EducationalInstitution    string           `json:"educational_institution"`
	GraduationYear            time.Time        `json:"graduation_year"`
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
	Duties       string    `json:"duties"`
	Achievements string    `json:"achievements"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	UntilNow     bool      `json:"until_now"`
}

func (r *Resume) Validate() error {
	if r.ApplicantID <= 0 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректный ID соискателя"),
		)
	}

	if utf8.RuneCountInString(r.AboutMe) > 2000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("информация о себе не может быть длиннее 2000 символов"),
		)
	}

	if r.SpecializationID <= 0 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректный ID специализации"),
		)
	}

	if r.Education < 1 || r.Education > 5 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("некорректный тип образования"),
		)
	}

	if utf8.RuneCountInString(r.EducationalInstitution) > 255 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("название учебного заведения не может быть длиннее 255 символов"),
		)
	}

	return nil
}

func (w *WorkExperience) Validate() error {
	if utf8.RuneCountInString(w.EmployerName) > 100 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("название работодателя не может быть длиннее 100 символов"),
		)
	}

	if utf8.RuneCountInString(w.Position) > 100 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("должность не может быть длиннее 100 символов"),
		)
	}

	if utf8.RuneCountInString(w.Duties) > 1000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("обязанности не могут быть длиннее 1000 символов"),
		)
	}

	if utf8.RuneCountInString(w.Achievements) > 1000 {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("достижения не могут быть длиннее 1000 символов"),
		)
	}

	if !w.UntilNow && w.EndDate.Before(w.StartDate) {
		return NewError(
			ErrBadRequest,
			fmt.Errorf("дата окончания работы не может быть раньше даты начала"),
		)
	}

	return nil
}
