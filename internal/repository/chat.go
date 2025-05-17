package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type ChatRepository interface {
	Create(ctx context.Context, vacancyID, employerID, applicantID int) (int, error)
	GetByID(ctx context.Context, chatID int) (*entity.Chat, error)
}
