package postgres

import (
	"ResuMatch/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestCityRepository_GetCityByID(t *testing.T) {
	t.Parallel()

	query := `SELECT id, name FROM city WHERE id = $1`

	columns := []string{"id", "name"}

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
				Name: "Москва",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows(columns).AddRow(
						1,
						"Москва",
					))
			},
		},
		{
			name:           "Город не найден по ID",
			id:             777,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("город с id=777 не найден"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка PostgreSQL",
			id:             5,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить город по id=5"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnError(errors.New("database connection error"))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tc.setupMock(mock, tc.id)

			repo := &CityRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetCityByID(ctx, tc.id)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCityRepository_GetCityByName(t *testing.T) {
	t.Parallel()

	query := `SELECT id, name FROM city WHERE name = $1`

	columns := []string{"id", "name"}

	testCases := []struct {
		name           string
		city           string
		expectedResult *entity.City
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, city string)
	}{
		{
			name: "Успешное получение города по названию",
			city: "Москва",
			expectedResult: &entity.City{
				ID:   1,
				Name: "Москва",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, city string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(city).
					WillReturnRows(sqlmock.NewRows(columns).AddRow(
						1,
						"Москва",
					))
			},
		},
		{
			name:           "Город не найден по названию",
			city:           "Подземелье",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("город с name=Подземелье не найден"),
			),
			setupMock: func(mock sqlmock.Sqlmock, city string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(city).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка PostgreSQL",
			city:           "Москва",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось найти город с name=Москва"),
			),
			setupMock: func(mock sqlmock.Sqlmock, city string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(city).
					WillReturnError(errors.New("database connection error"))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tc.setupMock(mock, tc.city)

			repo := &CityRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetCityByName(ctx, tc.city)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
