package postgres

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/metrics"
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
		metrics.LayerErrorCounter.WithLabelValues("Specialization Repository", "GetByID").Inc()
		return nil, entity.NewError(
			entity.ErrNotFound,
			fmt.Errorf("специализация с id=%d не найдена", id),
		)
	} else if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Specialization Repository", "GetByID").Inc()
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
		metrics.LayerErrorCounter.WithLabelValues("Specialization Repository", "GetAll").Inc()
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
			metrics.LayerErrorCounter.WithLabelValues("Specialization Repository", "GetAll").Inc()
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

	var specializations []entity.Specialization
	for rows.Next() {
		var specialization entity.Specialization
		if err := rows.Scan(&specialization.ID, &specialization.Name); err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Specialization Repository", "GetAll").Inc()
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
		metrics.LayerErrorCounter.WithLabelValues("Specialization Repository", "GetAll").Inc()
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

func (r *SpecializationRepository) GetSpecializationSalaries(ctx context.Context) ([]entity.SpecializationSalaryRange, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
		SELECT 
			s.id, s.name,
			MIN(v.salary_from) AS min_salary,
			MAX(v.salary_to) AS max_salary,
			ROUND(AVG((v.salary_from + v.salary_to) / 2)) AS avg_salary
		FROM 
			specialization s
		JOIN 
			vacancy v ON s.id = v.specialization_id 
					AND v.is_active = TRUE 
					AND v.salary_from IS NOT NULL 
					AND v.salary_to IS NOT NULL
		GROUP BY 
			s.id, s.name
		ORDER BY 
			s.name;
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Specialization Repository", "getSpecializationSalaries").Inc()
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при получении списка вилок специализаций")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении списка вилок специализаций: %w", err),
		)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Specialization Repository", "getSpecializationSalaries").Inc()
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

	var specializations []entity.SpecializationSalaryRange
	for rows.Next() {
		var specialization entity.SpecializationSalaryRange
		if err := rows.Scan(
			&specialization.ID,
			&specialization.Name,
			&specialization.MinSalary,
			&specialization.MaxSalary,
			&specialization.AvgSalary,
		); err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Specialization Repository", "getSpecializationSalaries").Inc()
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
		metrics.LayerErrorCounter.WithLabelValues("Specialization Repository", "getSpecializationSalaries").Inc()
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
