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

	"github.com/sirupsen/logrus"
)

type SpecializationRepository struct {
	DB *sql.DB
}

func NewSpecializationRepository(db *sql.DB) (repository.SpecializationRepository, error) {
	return &SpecializationRepository{DB: db}, nil
}

func (r *SpecializationRepository) GetByID(ctx context.Context, id int) (*entity.Specialization, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
		SELECT id, name
		FROM specialization
		WHERE id = $1
	`

	var specialization entity.Specialization
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&specialization.ID,
		&specialization.Name,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.NewError(
			entity.ErrNotFound,
			fmt.Errorf("специализация с id=%d не найдена", id),
		)
	} else if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        id,
			"error":     err,
		}).Error("не удалось найти специализацию по id")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить специализацию по id=%d", id),
		)
	}

	return &specialization, nil
}

func (r *SpecializationRepository) GetAll(ctx context.Context) ([]entity.Specialization, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
		SELECT id, name
		FROM specialization
		ORDER BY name
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при получении списка специализаций")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении списка специализаций: %w", err),
		)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

	var specializations []entity.Specialization
	for rows.Next() {
		var specialization entity.Specialization
		if err := rows.Scan(&specialization.ID, &specialization.Name); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("ошибка при сканировании специализации")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании специализации: %w", err),
			)
		}
		specializations = append(specializations, specialization)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при итерации по специализациям")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по специализациям: %w", err),
		)
	}

	return specializations, nil
}
