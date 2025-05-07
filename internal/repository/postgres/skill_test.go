package postgres

import (
	"ResuMatch/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestSkillRepository_GetByIDs(t *testing.T) {
	t.Parallel()

	columns := []string{"id", "name"}

	testCases := []struct {
		name           string
		ids            []int
		expectedResult []entity.Skill
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, ids []int)
	}{
		{
			name:           "Пустой список ID",
			ids:            []int{},
			expectedResult: []entity.Skill{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				// No database queries expected for empty input
			},
		},
		{
			name: "Успешное получение списка навыков",
			ids:  []int{1, 2},
			expectedResult: []entity.Skill{
				{ID: 1, Name: "Go"},
				{ID: 2, Name: "SQL"},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT id, name
					FROM skill
					WHERE id IN (%s)
				`, strings.Join([]string{"$1", "$2"}, ", ")))
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Go").
					AddRow(2, "SQL")
				mock.ExpectQuery(query).
					WithArgs(1, 2).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Пустой список навыков",
			ids:            []int{3, 4},
			expectedResult: []entity.Skill{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT id, name
					FROM skill
					WHERE id IN (%s)
				`, strings.Join([]string{"$1", "$2"}, ", ")))
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(3, 4).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка базы данных при выполнении запроса",
			ids:            []int{1, 2},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении навыков по ID: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT id, name
					FROM skill
					WHERE id IN (%s)
				`, strings.Join([]string{"$1", "$2"}, ", ")))
				mock.ExpectQuery(query).
					WithArgs(1, 2).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка сканирования строк",
			ids:            []int{1, 2},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по навыкам: %w", errors.New("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT id, name
					FROM skill
					WHERE id IN (%s)
				`, strings.Join([]string{"$1", "$2"}, ", ")))
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Go").
					RowError(0, errors.New("scan error"))
				mock.ExpectQuery(query).
					WithArgs(1, 2).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка итерации по строкам",
			ids:            []int{1, 2},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по навыкам: %w", errors.New("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT id, name
					FROM skill
					WHERE id IN (%s)
				`, strings.Join([]string{"$1", "$2"}, ", ")))
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Go")
				rows.CloseError(errors.New("rows error"))
				mock.ExpectQuery(query).
					WithArgs(1, 2).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка закрытия строк",
			ids:            []int{1, 2},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по навыкам: %w", errors.New("close error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT id, name
					FROM skill
					WHERE id IN (%s)
				`, strings.Join([]string{"$1", "$2"}, ", ")))
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Go")
				rows.CloseError(errors.New("close error"))
				mock.ExpectQuery(query).
					WithArgs(1, 2).
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
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.ids)

			repo := &SkillRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetByIDs(ctx, tc.ids)

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
