package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type CityRepository interface {
	GetCityByID(context.Context, int) (*entity.City, error)
	// GetAllCities(context.Context) ([]*entity.City, error)
	GetCityByName(context.Context, string) (*entity.City, error)
}
