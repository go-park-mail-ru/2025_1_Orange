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

func TestStaticRepository_GetStatic(t *testing.T) {
	t.Parallel()

	query := `SELECT file_path, file_name FROM static WHERE id = $1`

	columns := []string{"file_path", "file_name"}
	filePath := "assets/img"
	fileName := "avatar.png"

	testCases := []struct {
		name           string
		id             int
		expectedResult string
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name:           "Успешное получение пути до файла по ID",
			id:             1,
			expectedResult: fmt.Sprintf("/%s/%s", filePath, fileName),
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows(columns).AddRow(
						filePath,
						fileName,
					))
			},
		},
		{
			name:           "Статический файл не найден по ID",
			id:             777,
			expectedResult: "",
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("файл с id=777 не найден"),
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
			expectedResult: "",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении запроса GetStatic: %w", errors.New("postgresql error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnError(errors.New("postgresql error"))
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
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}()

			tc.setupMock(mock, tc.id)

			repo := &StaticRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetStatic(ctx, tc.id)

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
