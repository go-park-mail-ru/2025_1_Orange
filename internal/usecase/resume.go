package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

// Исправление 5: Переименовано из Resume в ResumeUsecase
type ResumeUsecase interface {
	// Исправление 6: Добавлен applicantID как параметр вместо использования контекста
	Create(ctx context.Context, applicantID int, request *dto.CreateResumeRequest) (*dto.ResumeResponse, error)
	GetByID(ctx context.Context, id int) (*dto.ResumeResponse, error)
	// Исправление 6: Добавлен applicantID как параметр вместо использования контекста
	Update(ctx context.Context, id int, applicantID int, request *dto.UpdateResumeRequest) (*dto.ResumeResponse, error)
	Delete(ctx context.Context, id int, applicantID int) (*dto.DeleteResumeResponse, error)
	// Исправление 8: Добавлен лимит для безопасной работы с большим количеством резюме
	GetAll(ctx context.Context) ([]dto.ResumeShortResponse, error)
}
