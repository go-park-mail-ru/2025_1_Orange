package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Chat interface {
	StartChat(ctx context.Context, vacancyID, resumeID, applicantID, employerID int) (int, error)
	GetChat(ctx context.Context, chatID int, userID int, role string) (*dto.ChatResponse, error)
	SendMessage(ctx context.Context, chatID, senderID int, role string, payload string) (*dto.MessageResponse, error)
	GetUserChats(ctx context.Context, userID int, role string) (dto.ChatResponseList, error)
	GetChatMessages(ctx context.Context, chatID int) (dto.MessagesResponseList, error)
	GetVacancyChat(ctx context.Context, vacancyID, applicantID int, role string) (*dto.ChatResponse, error)
}
