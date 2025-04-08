package usecase

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"context"
)

type Vacancy interface {
	SearchVacancies(ctx context.Context) ([]*entity.Vacancy, error)
	GetVacanciesByEmpID(ctx context.Context, employerID int) ([]*dto.Vacancy, error)
	CreateVacancy(ctx context.Context, user *dto.UserFromSession) (*entity.Vacancy, error)
	GetVacancy(ctx context.Context, id int) (int, error)
	UpdateVacancy(ctx context.Context, id int, update *entity.Vacancy, user *dto.UserFromSession) (*entity.Vacancy, error)
	DeleteVacancy(ctx context.Context, id int, user *dto.UserFromSession) error
}
