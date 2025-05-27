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

type CityRepository struct {
	DB *sql.DB
}

func NewCityRepository(db *sql.DB) (repository.CityRepository, error) {
	return &CityRepository{DB: db}, nil
}

func (r *CityRepository) GetCityByID(ctx context.Context, id int) (*entity.City, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"id":        id,
	}).Info("выполнение sql-запроса получения города по ID GetCityByID")

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

//func (r *CityRepository) GetAllCities(ctx context.Context) ([]*entity.City, error) {
//	requestID := utils.GetRequestID(ctx)
//
//	l.Log.WithFields(logrus.Fields{
//		"requestID": requestID,
//	}).Info("выполнение sql-запроса получения всех городов GetAllCities")
//
//	query := `SELECT id, name FROM city ORDER BY name ASC`
//	rows, err := r.DB.QueryContext(ctx, query)
//	if err != nil {
//		if errors.Is(err, sql.ErrNoRows) {
//			return nil, entity.NewError(
//				entity.ErrNotFound,
//				fmt.Errorf("список городов пустой"),
//			)
//		}
//	}
//	defer rows.Close()
//
//	var cities []*entity.City
//	for rows.Next() {
//		var city entity.City
//		if err := rows.Scan(&city.ID, &city.Name); err != nil {
//			l.Log.WithFields(logrus.Fields{
//				"requestID": requestID,
//				"error":     err,
//			}).Error("не удалось получить список городов")
//
//			return nil, entity.NewError(
//				entity.ErrInternal,
//				fmt.Errorf("не удалось получить список городов"),
//			)
//		}
//		cities = append(cities, &city)
//	}
//	return cities, nil
//
//}

func (r *CityRepository) GetCityByName(ctx context.Context, name string) (*entity.City, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"name":      name,
	}).Info("выполнение sql-запроса получения города по названию GetCityByName")

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
