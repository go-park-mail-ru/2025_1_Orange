package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Resume interface {
	Create(ctx context.Context, request *dto.CreateResumeRequest) (*dto.ResumeResponse, error)
	GetByID(ctx context.Context, id int) (*dto.ResumeResponse, error)
	// Updated methods
	Update(ctx context.Context, id int, request *dto.UpdateResumeRequest) (*dto.ResumeResponse, error)
	Delete(ctx context.Context, id int, applicantID int) (*dto.DeleteResumeResponse, error)
	// New method for employers
	GetAll(ctx context.Context) ([]dto.ResumeShortResponse, error)
}
