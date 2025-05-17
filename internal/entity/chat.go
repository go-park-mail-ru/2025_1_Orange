package entity

import (
	"time"
)

type Chat struct {
	ID          int       `json:"id"`
	VacancyID   int       `json:"vacancy_id"`
	ApplicantID int       `json:"applicant_id"`
	EmployerID  int       `json:"employer_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
