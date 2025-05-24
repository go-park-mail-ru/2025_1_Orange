package usecase

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"context"
)

type Vacancy interface {
	CreateVacancy(ctx context.Context, employerID int, createReq *dto.VacancyCreate) (*dto.VacancyResponse, error)
	GetVacancy(ctx context.Context, id, currentUserID int, userRole string) (*dto.VacancyResponse, error)
	UpdateVacancy(ctx context.Context, id int, employerID int, request *dto.VacancyUpdate) (*dto.VacancyResponse, error)
	DeleteVacancy(ctx context.Context, id int, employerID int) (*dto.DeleteVacancy, error)
	GetAll(ctx context.Context, currentUserID int, userRole string, limit int, offset int) ([]dto.VacancyShortResponse, error)
	ApplyToVacancy(ctx context.Context, vacancyID, applicantID, resumeID int) (*entity.Notification, error)
	GetVacanciesByApplicantID(ctx context.Context, applicantID int, limit int, offset int) ([]dto.VacancyShortResponse, error)
	GetActiveVacanciesByEmployerID(ctx context.Context, employerID, userID int, userRole string, limit int, offset int) ([]dto.VacancyShortResponse, error)
	SearchVacancies(ctx context.Context, userID int, userRole string, searchQuery string, limit int, offset int) ([]dto.VacancyShortResponse, error)
	SearchVacanciesBySpecializations(ctx context.Context, userID int, userRole string, specializations []string, limit int, offset int) ([]dto.VacancyShortResponse, error)
	SearchVacanciesByQueryAndSpecializations(ctx context.Context, userID int, userRole string, searchQuery string, specializations []string, limit int, offset int) ([]dto.VacancyShortResponse, error)
	LikeVacancy(ctx context.Context, vacancyID, applicantID int) error
	GetLikedVacancies(ctx context.Context, applicantID int, limit, offset int) ([]dto.VacancyShortResponse, error)
	GetRespondedResumeOnVacancy(ctx context.Context, vacancyID int, limit, offset int) ([]dto.ResumeShortResponse, error)
}
