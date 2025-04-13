package connector

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	l "ResuMatch/pkg/logger"
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
)

func NewPostgresConnection(cfg config.PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось установить соединение с PostgreSQL")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить соединение PostgreSQL: %w", err),
		)
	}

	if err := db.Ping(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось выполнить ping PostgreSQL")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось выполнить ping PostgreSQL: %w", err),
		)
	}
	return db, nil
}
