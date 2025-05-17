package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type MessageRepository interface {
	Create(ctx context.Context, chatID, senderID int, fromApplicant bool, payload string) (int, error)
	GetForChat(ctx context.Context, chatID int) ([]*entity.Message, error)
}
