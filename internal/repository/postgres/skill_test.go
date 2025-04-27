package postgres

import (
	"ResuMatch/internal/entity"
	"context"
	"errors"
	"fmt"
	"regexp"
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
			name: "Успешное получение навыков по ID",
			ids:  []int{1, 2, 3},
			expectedResult: []entity.Skill{
				{ID: 1, Name: "Go"},
				{ID: 2, Name: "SQL"},
				{ID: 3, Name: "Docker"},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				query := regexp.QuoteMeta(fmt.Sprintf(
					`SELECT id, name
					FROM skill
					WHERE id IN (%s)`,
					"$1, $2, $3",
				))
				mock.ExpectQuery(query).
					WithArgs(ids[0], ids[1], ids[2]).
					WillReturnRows(sqlmock.NewRows(columns).
						AddRow(1, "Go").
						AddRow(2, "SQL").
						AddRow(3, "Docker"))
			},
		},
		{
			name:           "Пустой список ID",
			ids:            []int{},
			expectedResult: []entity.Skill{},
			expectedErr:    nil,
			setupMock:      func(mock sqlmock.Sqlmock, ids []int) {},
		},
		{
			name: "Частичное нахождение навыков",
			ids:  []int{1, 2, 999},
			expectedResult: []entity.Skill{
				{ID: 1, Name: "Go"},
				{ID: 2, Name: "SQL"},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				query := regexp.QuoteMeta(fmt.Sprintf(
					`SELECT id, name
					FROM skill
					WHERE id IN (%s)`,
					"$1, $2, $3",
				))
				mock.ExpectQuery(query).
					WithArgs(ids[0], ids[1], ids[2]).
					WillReturnRows(sqlmock.NewRows(columns).
						AddRow(1, "Go").
						AddRow(2, "SQL"))
			},
		},
		{
			name:           "Ошибка при выполнении запроса",
			ids:            []int{1, 2},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении навыков по ID: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				query := regexp.QuoteMeta(fmt.Sprintf(
					`SELECT id, name
					FROM skill
					WHERE id IN (%s)`,
					"$1, $2",
				))
				mock.ExpectQuery(query).
					WithArgs(ids[0], ids[1]).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка при сканировании строк",
			ids:            []int{1},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по навыкам: %w", errors.New("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				query := regexp.QuoteMeta(fmt.Sprintf(
					`SELECT id, name
					FROM skill
					WHERE id IN (%s)`,
					"$1",
				))
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Go").
					AddRow(2, "SQL").
					RowError(1, errors.New("scan error"))
				mock.ExpectQuery(query).
					WithArgs(ids[0]).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка при итерации по строкам",
			ids:            []int{1, 2},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по навыкам: %w", errors.New("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, ids []int) {
				query := regexp.QuoteMeta(fmt.Sprintf(
					`SELECT id, name
					FROM skill
					WHERE id IN (%s)`,
					"$1, $2",
				))
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Go").
					AddRow(2, "SQL").
					CloseError(errors.New("rows error"))
				mock.ExpectQuery(query).
					WithArgs(ids[0], ids[1]).
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

			tc.setupMock(mock, tc.ids)

			repo := &SkillRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetByIDs(ctx, tc.ids)

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
