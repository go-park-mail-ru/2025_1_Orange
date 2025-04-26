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

func (r *PollRepository) GetAll(ctx context.Context) ([]*entity.Poll, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("выполнение sql-запроса поолучения всех опросов")

	query := `SELECT id, name FROM poll`

	rows, err := r.DB.QueryContext(ctx, query)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("список опросов пустой"),
			)
		}
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении всех опросов err=%w", err),
		)
	}

	defer rows.Close()

	var polls []*entity.Poll
	for rows.Next() {
		p := new(entity.Poll)
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		polls = append(polls, p)
	}

	return polls, nil
}

func (r *PollRepository) GetVotesByPoll(ctx context.Context, pollID int) ([]*entity.VoteStats, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"pollID":    pollID,
	}).Info("выполнение sql-запроса получения голосов по опросу")

	query := `
        SELECT answer, COUNT(*) as count 
        FROM vote 
        WHERE poll_id = $1
        GROUP BY answer
        ORDER BY answer
    `

	rows, err := r.DB.QueryContext(ctx, query, pollID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("список опросов пустой"),
			)
		}
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении всех опросов err=%w", err),
		)
	}
	defer rows.Close()

	var stats []*entity.VoteStats
	for rows.Next() {
		vs := new(entity.VoteStats)
		if err := rows.Scan(&vs.Answer, &vs.Count); err != nil {
			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при маппинге статистики голосов err=%w", err),
			)
		}
		stats = append(stats, vs)
	}

	return stats, nil
}
