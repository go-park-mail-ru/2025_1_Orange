package postgres

import (
	"ResuMatch/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestCityRepository_GetCityByID(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`SELECT id, name FROM city WHERE id = $1`)

	testCases := []struct {
		name           string
		id             int
		expectedResult *entity.City
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name: "Успешное получение города по ID",
			id:   1,
			expectedResult: &entity.City{
				ID:   1,
				Name: "Moscow",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Moscow")
				mock.ExpectQuery(query).
					WithArgs(id).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Город не найден",
			id:             999,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("город с id=%d не найден", 999),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(query).
					WithArgs(id).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка базы данных",
			id:             1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить город по id=%d", 1),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(query).
					WithArgs(id).
					WillReturnError(errors.New("database error"))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close() //nolint:errcheck

			tc.setupMock(mock, tc.id)

			repo := &CityRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetCityByID(ctx, tc.id)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Equal(t, tc.expectedResult, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.Name, result.Name)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCityRepository_GetCityByName(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`SELECT id, name FROM city WHERE name = $1`)

	testCases := []struct {
		name           string
		cityName       string
		expectedResult *entity.City
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, cityName string)
	}{
		{
			name:     "Успешное получение города по названию",
			cityName: "Moscow",
			expectedResult: &entity.City{
				ID:   1,
				Name: "Moscow",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, cityName string) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Moscow")
				mock.ExpectQuery(query).
					WithArgs(cityName).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Город не найден",
			cityName:       "Nonexistent",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("город с name=%s не найден", "Nonexistent"),
			),
			setupMock: func(mock sqlmock.Sqlmock, cityName string) {
				mock.ExpectQuery(query).
					WithArgs(cityName).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка базы данных",
			cityName:       "Moscow",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось найти город с name=%s", "Moscow"),
			),
			setupMock: func(mock sqlmock.Sqlmock, cityName string) {
				mock.ExpectQuery(query).
					WithArgs(cityName).
					WillReturnError(errors.New("database error"))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close() //nolint:errcheck

			tc.setupMock(mock, tc.cityName)

			repo := &CityRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetCityByName(ctx, tc.cityName)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Equal(t, tc.expectedResult, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.Name, result.Name)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
