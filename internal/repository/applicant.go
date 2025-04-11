package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type ApplicantRepository interface {
	CreateApplicant(ctx context.Context, email, firstName, lastName string, passwordHash, passwordSalt []byte) (*entity.Applicant, error)
	GetApplicantByID(ctx context.Context, id int) (*entity.Applicant, error)
	GetApplicantByEmail(ctx context.Context, email string) (*entity.Applicant, error)
	UpdateApplicant(ctx context.Context, userID int, fields map[string]interface{}) error
}
