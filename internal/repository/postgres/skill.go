package postgres

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> 71cf6a4 (Made vacansies usecases and handlers)
=======
>>>>>>> d7704b3 (Fix mistakes)

	"ResuMatch/internal/repository"
	"ResuMatch/internal/utils"

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d7704b3 (Fix mistakes)
=======
	"ResuMatch/internal/repository"
	"ResuMatch/internal/utils"
>>>>>>> a6396a4 (Fix mistakes)
<<<<<<< HEAD
=======
>>>>>>> 71cf6a4 (Made vacansies usecases and handlers)
=======
>>>>>>> d7704b3 (Fix mistakes)
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

func NewSkillRepository(cfg config.PostgresConfig) (repository.SkillRepository, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось установить соединение с PostgreSQL из SkillRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить соединение PostgreSQL из SkillRepository: %w", err),
		)
	}

	if err := db.Ping(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось выполнить ping PostgreSQL из SkillRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось выполнить ping PostgreSQL из SkillRepository: %w", err),
		)
	}
	return &SkillRepository{DB: db}, nil
}

func (r *SkillRepository) GetByIDs(ctx context.Context, ids []int) ([]entity.Skill, error) {
	requestID := utils.GetRequestID(ctx)

	if len(ids) == 0 {
		return []entity.Skill{}, nil
	}

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
