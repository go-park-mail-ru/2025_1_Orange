package repository

import (
	"context"
)

type StaticRepository interface {
	UploadStatic(ctx context.Context, fileName string, contentType string, data []byte) (int, string, error)
	GetStatic(ctx context.Context, id int) (string, error)
	DeleteStatic(ctx context.Context, id int) error
}
