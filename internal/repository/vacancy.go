package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type VacancyRepository interface {
	Create(ctx context.Context, vacancy *entity.Vacancy) (int, error)
	GetByID(ctx context.Context, id int) (*entity.Vacancy, error)
	Update(ctx context.Context, vacancy *entity.Vacancy) error
	GetAll(ctx context.Context) ([]*entity.Vacancy, error)
	Delete(ctx context.Context, employerID, vacancyID int) error
	GetVacanciesByEmpID(ctx context.Context, employerID int) ([]*entity.Vacancy, error)
	//Subscribe(ctx context.Context, vacancyID uint64, applicantID uint64) error
	//Unsubscribe(ctx context.Context, vacancyID uint64, applicantID uint64) error
}
