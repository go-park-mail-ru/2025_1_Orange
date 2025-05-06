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

	query := regexp.QuoteMeta(`
		SELECT id, name
		FROM specialization
		WHERE id = $1
	`)

	testCases := []struct {
		name           string
		id             int
		expectedResult *entity.Specialization
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name: "Успешное получение специализации",
			id:   1,
			expectedResult: &entity.Specialization{
				ID:   1,
				Name: "Backend Developer",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Backend Developer")
				mock.ExpectQuery(query).
					WithArgs(id).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Специализация не найдена",
			id:             2,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("специализация с id=%d не найдена", 2),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(query).
					WithArgs(id).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка базы данных",
			id:             3,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить специализацию по id=%d", 3),
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
			defer db.Close()

			tc.setupMock(mock, tc.id)

			repo := &SpecializationRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetByID(ctx, tc.id)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.Name, result.Name)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestSpecializationRepository_GetAll(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		SELECT id, name
		FROM specialization
		ORDER BY name
	`)

	columns := []string{"id", "name"}

	testCases := []struct {
		name           string
		expectedResult []entity.Specialization
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock)
	}{
		{
			name: "Успешное получение списка специализаций",
			expectedResult: []entity.Specialization{
				{ID: 1, Name: "Backend Developer"},
				{ID: 2, Name: "Frontend Developer"},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Backend Developer").
					AddRow(2, "Frontend Developer")
				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Пустой список специализаций",
			expectedResult: []entity.Specialization{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка базы данных при выполнении запроса",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка специализаций: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка сканирования строк",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по специализациям: %w", errors.New("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Backend Developer").
					RowError(0, errors.New("scan error"))
				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка итерации по строкам",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по специализациям: %w", errors.New("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Backend Developer")
				rows.CloseError(errors.New("rows error"))
				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка закрытия строк",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по специализациям: %w", errors.New("close error")),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Backend Developer")
				rows.CloseError(errors.New("close error"))
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
			defer db.Close()

			tc.setupMock(mock)

			repo := &SpecializationRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetAll(ctx)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))
				for i, expected := range tc.expectedResult {
					require.Equal(t, expected.ID, result[i].ID)
					require.Equal(t, expected.Name, result[i].Name)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
