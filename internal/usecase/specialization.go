package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type SpecializationUsecase interface {
	GetAllSpecializationNames(ctx context.Context) (*dto.SpecializationNamesResponse, error)
}
