package postgres

import (
	"ResuMatch/internal/config"
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

type CityRepository struct {
	DB *sql.DB
}

func NewCityRepository(cfg config.PostgresConfig) (repository.CityRepository, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось установить соединение с PostgreSQL из CityRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить соединение PostgreSQL из CityRepository: %w", err),
		)
	}

	if err := db.Ping(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось выполнить ping PostgreSQL из CityRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось выполнить ping PostgreSQL из CityRepository: %w", err),
		)
	}
	return &CityRepository{DB: db}, nil
}

func (r *CityRepository) GetByID(ctx context.Context, id int) (*entity.City, error) {
	requestID := utils.GetRequestID(ctx)

	query := `SELECT id, name FROM city WHERE id = $1`
	var city entity.City
	err := r.DB.QueryRowContext(ctx, query, id).Scan(&city.ID, &city.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("город с id=%d не найден", id),
			)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        id,
			"error":     err,
		}).Error("не удалось найти город по id")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить город по id=%d", id),
		)
	}

	return &city, nil
}

func (r *CityRepository) GetAll(ctx context.Context) ([]*entity.City, error) {
	requestID := utils.GetRequestID(ctx)

	query := `SELECT id, name FROM city ORDER BY name ASC`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("список городов пустой"),
			)
		}
	}
	defer rows.Close()

	var cities []*entity.City
	for rows.Next() {
		var city entity.City
		if err := rows.Scan(&city.ID, &city.Name); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("не удалось получить список городов")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить список городов"),
			)
		}
		cities = append(cities, &city)
	}
	return cities, nil

}

func (r *CityRepository) GetByName(ctx context.Context, name string) (*entity.City, error) {
	requestID := utils.GetRequestID(ctx)

	query := `SELECT id, name FROM city WHERE name = $1`
	var city entity.City
	err := r.DB.QueryRowContext(ctx, query, name).Scan(&city.ID, &city.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("город с name=%s не найден", name),
			)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"name":      name,
			"error":     err,
		}).Error("не удалось найти город по названию")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось найти город с name=%s", name),
		)
	}
	return &city, nil
}
