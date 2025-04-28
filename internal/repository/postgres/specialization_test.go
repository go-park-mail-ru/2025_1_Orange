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

func TestSpecializationRepository_GetByID(t *testing.T) {
	t.Parallel()

	columns := []string{"id", "name"}

	testCases := []struct {
		name           string
		id             int
		expectedResult *entity.Specialization
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name: "Успешное получение специализации по ID",
			id:   1,
			expectedResult: &entity.Specialization{
				ID:   1,
				Name: "Backend разработка",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					SELECT id, name
					FROM specialization
					WHERE id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(id).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(1, "Backend разработка"),
					)
			},
		},
		{
			name:           "Специализация не найдена",
			id:             999,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("специализация с id=999 не найдена"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					SELECT id, name
					FROM specialization
					WHERE id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(id).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка базы данных",
			id:             2,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить специализацию по id=2"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					SELECT id, name
					FROM specialization
					WHERE id = $1
				`)
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
			defer func() {
				err := db.Close()
				require.NoError(t, err)
			}()

			tc.setupMock(mock, tc.id)

			repo := &SpecializationRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetByID(ctx, tc.id)

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

func TestSpecializationRepository_GetAll(t *testing.T) {
	t.Parallel()

	columns := []string{"id", "name"}

	testCases := []struct {
		name           string
		expectedResult []entity.Specialization
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock)
	}{
		{
			name: "Успешное получение всех специализаций",
			expectedResult: []entity.Specialization{
				{ID: 1, Name: "Backend разработка"},
				{ID: 2, Name: "Frontend разработка"},
				{ID: 3, Name: "DevOps"},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta(`
					SELECT id, name
					FROM specialization
					ORDER BY name
				`)
				mock.ExpectQuery(query).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(1, "Backend разработка").
							AddRow(2, "Frontend разработка").
							AddRow(3, "DevOps"),
					)
			},
		},
		{
			name:           "Пустой список специализаций",
			expectedResult: nil,
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta(`
					SELECT id, name
					FROM specialization
					ORDER BY name
				`)
				mock.ExpectQuery(query).
					WillReturnRows(sqlmock.NewRows(columns))
			},
		},
		{
			name:           "Ошибка при выполнении запроса",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка специализаций: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta(`
					SELECT id, name
					FROM specialization
					ORDER BY name
				`)
				mock.ExpectQuery(query).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка при сканировании строк",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании специализации: %w",
					fmt.Errorf("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid_id\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta(`
					SELECT id, name
					FROM specialization
					ORDER BY name
				`)
				mock.ExpectQuery(query).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow("invalid_id", "Backend разработка"), // Неправильный тип для id
					)
			},
		},
		{
			name:           "Ошибка при итерации по строкам",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по специализациям: %w", errors.New("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta(`
					SELECT id, name
					FROM specialization
					ORDER BY name
				`)
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Backend разработка").
					AddRow(2, "Frontend разработка").
					CloseError(errors.New("rows error"))
				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() {
				err := db.Close()
				require.NoError(t, err)
			}()

			tc.setupMock(mock)

			repo := &SpecializationRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetAll(ctx)

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
