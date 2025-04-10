package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
)

type StaticService struct {
	staticRepository repository.StaticRepository
}

var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	// "image/webp": ".webp",
}

func NewStaticService(staticRepository repository.StaticRepository) usecase.Static {
	return &StaticService{
		staticRepository: staticRepository,
	}
}

func (s *StaticService) UploadStatic(ctx context.Context, data []byte) (*dto.UploadStaticResponse, error) {
	const maxFileSize = 5 << 20
	if len(data) > maxFileSize {
		return nil, entity.NewError(entity.ErrBadRequest, fmt.Errorf("размер файла превышает 5MB"))
	}

	contentType := http.DetectContentType(data)
	ext, allowed := allowedTypes[contentType]
	if !allowed {
		return nil, entity.NewError(entity.ErrBadRequest, fmt.Errorf("недопустимый формат файла"))
	}

	if err := s.validateImageContent(data, contentType); err != nil {
		return nil, err
	}

	fileName := uuid.New().String() + ext
	filePath := "assets/img"
	static, err := s.staticRepository.UploadStatic(ctx, filePath, fileName, data)
	if err != nil {
		return nil, err
	}
	return &dto.UploadStaticResponse{
		ID:        static.ID,
		Path:      fmt.Sprintf("%s/%s", static.FilePath, static.FileName),
		CreatedAt: static.CreatedAt,
		UpdatedAt: static.UpdatedAt,
	}, nil
}

func (s *StaticService) validateImageContent(data []byte, contentType string) error {
	fmt.Println(contentType)
	switch contentType {
	case "image/jpeg", "image/png":
		if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
			return entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("невалидное изображение: %w", err),
			)
		}
	default:
		return entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("невалидное изображение"),
		)
	}
	return nil
}
