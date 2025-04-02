package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type ApplicantRepository interface {
	Create(ctx context.Context, applicant *entity.Applicant) (*entity.Applicant, error)
	GetByID(ctx context.Context, id int) (*entity.Applicant, error)
	GetByEmail(ctx context.Context, email string) (*entity.Applicant, error)
	Update(ctx context.Context, applicant *entity.Applicant) error
}
