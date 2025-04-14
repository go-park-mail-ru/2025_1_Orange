package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type CityRepository interface {
	GetCityByID(context.Context, int) (*entity.City, error)
<<<<<<< HEAD
=======
	// GetAllCities(context.Context) ([]*entity.City, error)
>>>>>>> a6396a4 (Fix mistakes)
	GetCityByName(context.Context, string) (*entity.City, error)
}
