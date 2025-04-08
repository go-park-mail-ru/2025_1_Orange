package usecase

import (
<<<<<<< HEAD
=======
	"ResuMatch/internal/entity"
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	"ResuMatch/internal/entity/dto"
	"context"
)

type Vacancy interface {
<<<<<<< HEAD
	CreateVacancy(ctx context.Context, createReq *dto.VacancyCreate) (*dto.VacancyResponse, error)
	GetVacancy(ctx context.Context, id int) (*dto.VacancyResponse, error)
	UpdateVacancy(ctx context.Context, id int, request *dto.VacancyUpdate) (*dto.VacancyResponse, error)
	DeleteVacancy(ctx context.Context, id int, employerID int) (*dto.DeleteVacancy, error)
	GetAll(ctx context.Context) ([]dto.VacancyShortResponse, error)
	//CreateResponse(ctx context.Context, vacancyID, applicantID int, resumeID *int) error
	ApplyToVacancy(ctx context.Context, vacancyID, applicantID, resumeID int) error
=======
	SearchVacancies(ctx context.Context) ([]*entity.Vacancy, error)
	GetVacanciesByEmpID(ctx context.Context, employerID int) ([]*dto.Vacancy, error)
	CreateVacancy(ctx context.Context, user *dto.UserFromSession) (*entity.Vacancy, error)
	GetVacancy(ctx context.Context, id int) (int, error)
	UpdateVacancy(ctx context.Context, id int, update *entity.Vacancy, user *dto.UserFromSession) (*entity.Vacancy, error)
	DeleteVacancy(ctx context.Context, id int, user *dto.UserFromSession) error
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
}
