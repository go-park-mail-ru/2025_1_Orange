package postgres

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) repository.MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) CreateMessage(ctx context.Context, chatID, senderID int, fromApplicant bool, payload string) (*entity.Message, error) {
	requestID := utils.GetRequestID(ctx)
	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("Выполнение sql-запроса создания сообщения CreateMessage")

	query := `
	INSERT INTO message (chat_id, sender_id, from_applicant, payload)
	VALUES ($1, $2, $3, $4)
	RETURNING id, chat_id, sender_id, from_applicant, payload, sent_at
	`

	var message entity.Message
	err := r.db.QueryRowContext(
		ctx,
		query,
		chatID,
		senderID,
		fromApplicant,
		payload,
	).Scan(
		&message.ID,
		&message.ChatID,
		&message.SenderID,
		&message.FromApplicant,
		&message.Payload,
		&message.SentAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLNotNullViolation:
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле таблицы message отсутствует: %w", pqErr),
				)
			case entity.PSQLCheckViolation:
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("указаны неправильные данные при создании сообщения: %w", pqErr),
				)
			}
		}
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при создании сообщения: %w", err),
		)
	}
	return &message, nil
}

func (r *MessageRepository) GetMessagesForChat(ctx context.Context, chatID int) ([]*entity.Message, error) {
	requestID := utils.GetRequestID(ctx)
	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("Выполнение sql-запроса получения всех сообщений чата")

	query := `
	SELECT id, chat_id, sender_id, from_applicant, payload, sent_at
	FROM message WHERE chat_id = $1
	ORDER BY sent_at ASC
    `
	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"chatID": chatID,
		}).Error("Не удалось получить сообщения чата")
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении сообщений чата: %w", err),
		)
	}

	var messages []*entity.Message

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"chatID":    chatID,
				"error":     closeErr,
			}).Error("Ошибка при закрытии строк после ошибки сканирования сообщений чата")
		}
	}()

	for rows.Next() {
		var message entity.Message
		err = rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.SenderID,
			&message.FromApplicant,
			&message.Payload,
			&message.SentAt,
		)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"chatID":    chatID,
				"error":     err,
			}).Error("Ошибка при сканировании строки результата сообщений чата")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении сообщений чата с id %d", chatID),
			)
		}
		messages = append(messages, &message)
	}

	if err = rows.Err(); err != nil {
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка итерации по строкам при получении сообщений чата с id %d", chatID),
		)
	}

	return messages, nil
}
