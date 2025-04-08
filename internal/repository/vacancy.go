package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type VacancyRepository interface {
<<<<<<< HEAD
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
	DeleteSkills(ctx context.Context, vacancyID int) error
	DeleteCity(ctx context.Context, vacancyID int) error
	FindSkillIDsByNames(ctx context.Context, skillNames []string) ([]int, error)
	FindCityIDsByNames(ctx context.Context, cityNames []string) ([]int, error)
	ResponseExists(ctx context.Context, vacancyID, applicantID int) (bool, error)
	CreateResponse(ctx context.Context, vacancyID, applicantID, resumeID int) error
=======
	Create(ctx context.Context, vacancy *entity.Vacancy) (int, error)
	GetByID(ctx context.Context, id int) (*entity.Vacancy, error)
	Update(ctx context.Context, vacancy *entity.Vacancy) error
	GetAll(ctx context.Context) ([]*entity.Vacancy, error)
	Delete(ctx context.Context, employerID, vacancyID int) error
	GetVacanciesByEmpID(ctx context.Context, employerID int) ([]*entity.Vacancy, error)
	//Subscribe(ctx context.Context, vacancyID uint64, applicantID uint64) error
	//Unsubscribe(ctx context.Context, vacancyID uint64, applicantID uint64) error
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
}
