package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type VacancyRepository interface {
	Create(ctx context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error)
	AddSkills(ctx context.Context, vacancyID int, skillIDs []int) error
	AddCity(ctx context.Context, vacancyID int, cityIDs []int) error
	GetByID(ctx context.Context, id int) (*entity.Vacancy, error)
	Update(ctx context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error)
	GetAll(ctx context.Context, limit int, offset int) ([]*entity.Vacancy, error)
	Delete(ctx context.Context, vacancyID int) error
	GetSkillsByVacancyID(ctx context.Context, vacancyID int) ([]entity.Skill, error)
	GetCityByVacancyID(ctx context.Context, vacancyID int) ([]entity.City, error)
	DeleteSkills(ctx context.Context, vacancyID int) error
	DeleteCity(ctx context.Context, vacancyID int) error
	FindSkillIDsByNames(ctx context.Context, skillNames []string) ([]int, error)
	FindCityIDsByNames(ctx context.Context, cityNames []string) ([]int, error)
	ResponseExists(ctx context.Context, vacancyID, applicantID int) (bool, error)
	CreateResponse(ctx context.Context, vacancyID, applicantID int) error
	FindSpecializationIDByName(ctx context.Context, specializationName string) (int, error)
	CreateSkillIfNotExists(ctx context.Context, skillName string) (int, error)
	CreateSpecializationIfNotExists(ctx context.Context, specializationName string) (int, error)
	GetActiveVacanciesByEmployerID(ctx context.Context, employerID int, limit int, offset int) ([]*entity.Vacancy, error)
	GetVacanciesByApplicantID(ctx context.Context, applicantID int, limit int, offset int) ([]*entity.Vacancy, error)
	SearchVacancies(ctx context.Context, searchQuery string, limit int, offset int) ([]*entity.Vacancy, error)
	SearchVacanciesByEmployerID(ctx context.Context, employerID int, searchQuery string, limit int, offset int) ([]*entity.Vacancy, error)
}
