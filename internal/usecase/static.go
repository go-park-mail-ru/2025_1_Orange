package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Static interface {
	UploadStatic(ctx context.Context, data []byte) (*dto.UploadStaticResponse, error)
	GetStatic(ctx context.Context, id int) (string, error)
	DeleteStatic(ctx context.Context, id int) error
}
