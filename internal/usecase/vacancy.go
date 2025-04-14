package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Vacancy interface {
	CreateVacancy(ctx context.Context, createReq *dto.VacancyCreate) (*dto.VacancyResponse, error)
	GetVacancy(ctx context.Context, id int) (*dto.VacancyResponse, error)
	UpdateVacancy(ctx context.Context, id int, request *dto.VacancyUpdate) (*dto.VacancyResponse, error)
	DeleteVacancy(ctx context.Context, id int, employerID int) (*dto.DeleteVacancy, error)
	GetAll(ctx context.Context) ([]dto.VacancyShortResponse, error)
	//CreateResponse(ctx context.Context, vacancyID, applicantID int, resumeID *int) error
	ApplyToVacancy(ctx context.Context, vacancyID, applicantID, resumeID int) error
}
