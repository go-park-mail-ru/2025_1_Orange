package usecase

import "context"

type Static interface {
	UploadStatic(ctx context.Context, data []byte) (int, error)
}
