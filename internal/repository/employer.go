package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type EmployerRepository interface {
	Create(ctx context.Context, employer *entity.Employer) (*entity.Employer, error)
	GetByID(ctx context.Context, id int) (*entity.Employer, error)
	GetByEmail(ctx context.Context, email string) (*entity.Employer, error)
	Update(ctx context.Context, employer *entity.Employer) error
}
