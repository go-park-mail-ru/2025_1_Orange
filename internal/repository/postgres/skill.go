package postgres

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/middleware"
	"ResuMatch/internal/repository"
	l "ResuMatch/pkg/logger"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type SkillRepository struct {
	DB *sql.DB
}

// Замечание 10 - Добавление коннектора
// Изменен конструктор для использования готового соединения с БД
func NewSkillRepository(db *sql.DB) (repository.SkillRepository, error) {
	return &SkillRepository{DB: db}, nil
}

func (r *SkillRepository) GetByIDs(ctx context.Context, ids []int) ([]entity.Skill, error) {
	requestID := middleware.GetRequestID(ctx)

	if len(ids) == 0 {
		return []entity.Skill{}, nil
	}

	// Создаем параметры для запроса
	params := make([]interface{}, len(ids))
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		params[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT id, name
		FROM skill
		WHERE id IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := r.DB.QueryContext(ctx, query, params...)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"ids":       ids,
			"error":     err,
		}).Error("ошибка при получении навыков по ID")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении навыков по ID: %w", err),
		)
	}
	defer rows.Close()

	var skills []entity.Skill
	for rows.Next() {
		var skill entity.Skill
		if err := rows.Scan(&skill.ID, &skill.Name); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("ошибка при сканировании навыка")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании навыка: %w", err),
			)
		}
		skills = append(skills, skill)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при итерации по навыкам")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по навыкам: %w", err),
		)
	}

	return skills, nil
}
