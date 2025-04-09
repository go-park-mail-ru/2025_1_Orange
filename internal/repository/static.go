package repository

import "context"

type StaticRepository interface {
	UploadStatic(ctx context.Context, filePath, fileName string, data []byte) (int, error)
	GetStatic(ctx context.Context, id int) (string, error)
}
