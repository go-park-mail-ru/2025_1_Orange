package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type EmployerRepository interface {
	CreateEmployer(ctx context.Context, email, companyName, legalAddress string, passwordHash, passwordSalt []byte) (*entity.Employer, error)
	GetEmployerByID(ctx context.Context, id int) (*entity.Employer, error)
	GetEmployerByEmail(ctx context.Context, email string) (*entity.Employer, error)
	UpdateEmployer(ctx context.Context, userID int, fields map[string]interface{}) error
}
