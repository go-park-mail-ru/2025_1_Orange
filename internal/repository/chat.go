package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type ChatRepository interface {
	CreateChat(ctx context.Context, vacancyID, resumeID, employerID, applicantID int) (*entity.Chat, error)
	GetChatByID(ctx context.Context, chatID int) (*entity.Chat, error)
	GetForUser(ctx context.Context, userID int, isApplicant bool) ([]*entity.Chat, error)
}
