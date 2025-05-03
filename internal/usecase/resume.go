package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type ResumeUsecase interface {
	Create(ctx context.Context, applicantID int, request *dto.CreateResumeRequest) (*dto.ResumeResponse, error)
	GetByID(ctx context.Context, id int) (*dto.ResumeResponse, error)
	Update(ctx context.Context, id int, applicantID int, request *dto.UpdateResumeRequest) (*dto.ResumeResponse, error)
	Delete(ctx context.Context, id int, applicantID int) (*dto.DeleteResumeResponse, error)
	GetAll(ctx context.Context, limit int, offset int) ([]dto.ResumeShortResponse, error)                                      // Дополнение - добавлена пагинация
	GetAllResumesByApplicantID(ctx context.Context, applicantID int, limit int, offset int) ([]dto.ResumeShortResponse, error) // Дополнение - добавлена пагинация
}
