package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type VacancyRepository interface {
	Create(ctx context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error)
	AddSkills(ctx context.Context, vacancyID int, skillIDs []int) error
	AddApplicant(ctx context.Context, vacancyID, applicantID int) error
	AddCity(ctx context.Context, vacancyID int, cityIDs []int) error
	GetByID(ctx context.Context, id int) (*entity.Vacancy, error)
	Update(ctx context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error)
	GetAll(ctx context.Context) ([]*entity.Vacancy, error)
	Delete(ctx context.Context, vacancyID int) error
	GetSkillsByVacancyID(ctx context.Context, vacancyID int) ([]entity.Skill, error)
	GetCityByVacancyID(ctx context.Context, vacancyID int) ([]entity.City, error)
	// GetVacancyResponsesByVacancyID(ctx context.Context, vacancyID int) ([]entity.Applicant, error)
	// GetVacancyLikesByVacancyID(ctx context.Context, vacancyID int) ([]entity.Applicant, error)
	DeleteSkills(ctx context.Context, vacancyID int) error
	DeleteCity(ctx context.Context, vacancyID int) error
	FindSkillIDsByNames(ctx context.Context, skillNames []string) ([]int, error)
	FindCityIDsByNames(ctx context.Context, cityNames []string) ([]int, error)
}
