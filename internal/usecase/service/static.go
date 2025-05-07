package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/metrics"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/google/uuid"
)

type StaticService struct {
	staticRepository repository.StaticRepository
}

var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
}

func NewStaticService(staticRepository repository.StaticRepository) usecase.Static {
	return &StaticService{
		staticRepository: staticRepository,
	}
}

func (s *StaticService) UploadStatic(ctx context.Context, data []byte) (*dto.UploadStaticResponse, error) {
	const maxFileSize = 5 << 20
	if len(data) > maxFileSize {
		metrics.LayerErrorCounter.WithLabelValues("Static Service", "UploadStatic").Inc()
		return nil, entity.NewError(entity.ErrBadRequest, fmt.Errorf("размер файла превышает 5MB"))
	}

	contentType := http.DetectContentType(data)
	ext, allowed := allowedTypes[contentType]
	if !allowed {
		metrics.LayerErrorCounter.WithLabelValues("Static Service", "UploadStatic").Inc()
		return nil, entity.NewError(entity.ErrBadRequest, fmt.Errorf("недопустимый формат файла"))
	}

	if err := s.validateImageContent(data, contentType); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Static Service", "UploadStatic").Inc()
		return nil, err
	}

	fileName := uuid.New().String() + ext
	id, path, err := s.staticRepository.UploadStatic(ctx, fileName, contentType, data)
	if err != nil {
		return nil, err
	}
	return &dto.UploadStaticResponse{
		ID:   id,
		Path: path,
	}, nil
}

func (s *StaticService) validateImageContent(data []byte, contentType string) error {
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

func (s *StaticService) GetStatic(ctx context.Context, id int) (string, error) {
	return s.staticRepository.GetStatic(ctx, id)
}

func (s *StaticService) DeleteStatic(ctx context.Context, id int) error {
	return s.staticRepository.DeleteStatic(ctx, id)
}
