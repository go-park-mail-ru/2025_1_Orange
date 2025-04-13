package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type StaticRepository interface {
	UploadStatic(ctx context.Context, filePath, fileName string, data []byte) (*entity.Static, error)
	GetStatic(ctx context.Context, id int) (string, error)
}
