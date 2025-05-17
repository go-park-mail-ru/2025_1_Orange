package postgres

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository"
	"context"
	"database/sql"
)

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) repository.ChatRepository {
	return &ChatRepository{
		db: db,
	}
}

func (r *ChatRepository) Create(ctx context.Context, vacancyID, employerID, applicantID int) (int, error) {
	panic("implement me")
}

func (r *ChatRepository) GetByID(ctx context.Context, chatID int) (*entity.Chat, error) {
	panic("implement me")
}
