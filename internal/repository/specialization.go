package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type SpecializationRepository interface {
	GetByID(ctx context.Context, id int) (*entity.Specialization, error)
	GetAll(ctx context.Context) ([]entity.Specialization, error)
}
