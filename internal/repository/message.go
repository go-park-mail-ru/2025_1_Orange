package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type MessageRepository interface {
	CreateMessage(ctx context.Context, chatID, senderID int, fromApplicant bool, payload string) (*entity.Message, error)
	GetMessagesForChat(ctx context.Context, chatID int) ([]*entity.Message, error)
}
