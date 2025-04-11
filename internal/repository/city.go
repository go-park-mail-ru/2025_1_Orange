package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type CityRepository interface {
	GetByID(context.Context, int) (*entity.City, error)
	GetAll(context.Context) ([]*entity.City, error)
	GetByName(context.Context, string) (*entity.City, error)
}
