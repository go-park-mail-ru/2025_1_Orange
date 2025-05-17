package postgres

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository"
	"context"
	"database/sql"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) repository.MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) Create(ctx context.Context, chatID, senderID int, fromApplicant bool, payload string) (int, error) {
	panic("implement me")
}

func (r *MessageRepository) GetForChat(ctx context.Context, chatID int) ([]*entity.Message, error) {
	panic("implement me")
}
