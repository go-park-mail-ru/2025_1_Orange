package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Static interface {
	UploadStatic(ctx context.Context, data []byte) (*dto.UploadStaticResponse, error)
}
