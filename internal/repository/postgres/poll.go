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

type PollRepository struct {
	DB *sql.DB
}

func NewPollRepository(db *sql.DB) (repository.PollRepository, error) {
	return &PollRepository{DB: db}, nil
}

func (r *PollRepository) CreateVote(ctx context.Context, voteEntity *entity.Vote) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("выполнение sql-запроса создания голоса CreateVote")

	query := `
		INSERT INTO vote (poll_id, user_id, role, answer)
		VALUES ($1, $2, $3, $4)
	`

	result, err := r.DB.ExecContext(ctx, query,
		voteEntity.PollID,
		voteEntity.UserID,
		voteEntity.Role,
		voteEntity.Answer,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation: // Уникальное ограничение
				return entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("ошибка уникальности"),
				)
			case entity.PSQLNotNullViolation: // NOT NULL ограничение
				return entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле отсутствует"),
				)
			default:
				return entity.NewError(
					entity.ErrInternal,
					fmt.Errorf("неизвестная ошибка при создании голоса err=%w", err),
				)
			}
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("не удалось создать голос")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось создать голос"),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("не удалось получить обновленные строки при создании голоса")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить обновленные строки при создании голоса"),
		)
	}

	if rowsAffected == 0 {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("отсутствуют обновленные строки при создании голоса")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("отсутствуют обновленные строки при создании голоса"),
		)
	}

	return nil
}
