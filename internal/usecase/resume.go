package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Resume interface {
	Create(ctx context.Context, request *dto.CreateResumeRequest) (*dto.ResumeResponse, error)
	GetByID(ctx context.Context, id int) (*dto.ResumeResponse, error)
}
