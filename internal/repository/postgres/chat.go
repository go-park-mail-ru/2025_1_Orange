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

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) repository.ChatRepository {
	return &ChatRepository{
		db: db,
	}
}

func (r *ChatRepository) CreateChat(ctx context.Context, vacancyID, resumeID, employerID, applicantID int) (*entity.Chat, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("выполнение sql-запроса создания чата CreateChat")

	query := `
		INSERT INTO chat (vacancy_id, resume_id, employer_id, applicant_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, vacancy_id, resume_id, employer_id, applicant_id, created_at, updated_at
	`

	var chat entity.Chat
	err := r.db.QueryRowContext(
		ctx,
		query,
		vacancyID,
		resumeID,
		employerID,
		applicantID,
	).Scan(
		&chat.ID,
		&chat.VacancyID,
		&chat.ResumeID,
		&chat.EmployerID,
		&chat.ApplicantID,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation:
				return nil, entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("нарушение условия уникальности при создании чата: %w", pqErr),
				)
			case entity.PSQLNotNullViolation:
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле таблицы чата отсутствует: %w", pqErr),
				)
			default:
				return nil, entity.NewError(
					entity.ErrInternal,
					fmt.Errorf("неизвестная ошибка при создании чата: %w", pqErr),
				)
			}
		}
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при создании чата: %w", err),
		)
	}
	return &chat, nil
}

func (r *ChatRepository) GetChatByID(ctx context.Context, chatID int) (*entity.Chat, error) {
	requestID := utils.GetRequestID(ctx)
	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("Выполнение sql-запроса получения чата по id GetChatByID")

	query := `
	SELECT id, vacancy_id, resume_id, applicant_id, employer_id, created_at, updated_at
	FROM chat WHERE id=$1
	`

	var chat entity.Chat
	err := r.db.QueryRowContext(ctx, query, chatID).Scan(
		&chat.ID,
		&chat.VacancyID,
		&chat.ResumeID,
		&chat.ApplicantID,
		&chat.EmployerID,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("чат с id=%d не найден", chatID),
			)
		}

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить чат по id=%d", chatID),
		)
	}
	return &chat, nil
}

func (r *ChatRepository) GetForUser(ctx context.Context, userID int, isApplicant bool) ([]*entity.Chat, error) {
	requestID := utils.GetRequestID(ctx)
	l.Log.WithFields(logrus.Fields{
		"requestID":   requestID,
		"userID":      userID,
		"isApplicant": isApplicant,
	}).Info("Выполнение sql-запроса получения чатов пользователя")

	query := `
        SELECT id, vacancy_id, resume_id, applicant_id, employer_id, created_at, updated_at
        FROM chat
        WHERE
            CASE
                WHEN $2 THEN applicant_id = $1
                ELSE employer_id = $1
            END
        ORDER BY updated_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID, isApplicant)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"userID":      userID,
			"isApplicant": isApplicant,
		}).Error("Не удалось получить чаты пользователя")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении чатов пользователя: %w", err),
		)
	}

	var chats []*entity.Chat

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":   requestID,
				"userID":      userID,
				"isApplicant": isApplicant,
				"error":       closeErr,
			}).Error("Ошибка при закрытии строк после ошибки сканирования сообщений чата")
		}
	}()

	for rows.Next() {
		var chat entity.Chat
		err = rows.Scan(
			&chat.ID,
			&chat.VacancyID,
			&chat.ResumeID,
			&chat.ApplicantID,
			&chat.EmployerID,
			&chat.CreatedAt,
			&chat.UpdatedAt,
		)

		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID":   requestID,
				"userID":      userID,
				"isApplicant": isApplicant,
				"error":       err,
			}).Error("Ошибка при сканировании строки результата списка чатов")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка чатов"),
			)
		}
		chats = append(chats, &chat)
	}
	if err = rows.Err(); err != nil {
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка итерации по строкам при получении списка чатов"),
		)
	}

	return chats, nil
}
