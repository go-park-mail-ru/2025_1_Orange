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
			// defer db.Close()
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

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
			// defer db.Close()
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

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

func TestSpecializationRepository_GetSpecializationSalaries(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		SELECT 
			s.id, s.name,
			MIN(v.salary_from) AS min_salary,
			MAX(v.salary_to) AS max_salary,
			ROUND(AVG((v.salary_from + v.salary_to) / 2)) AS avg_salary
		FROM 
			specialization s
		JOIN 
			vacancy v ON s.id = v.specialization_id 
					AND v.is_active = TRUE 
					AND v.salary_from IS NOT NULL 
					AND v.salary_to IS NOT NULL
		GROUP BY 
			s.id, s.name
		ORDER BY 
			s.name;
	`)

	columns := []string{"id", "name", "min_salary", "max_salary", "avg_salary"}

	testCases := []struct {
		name           string
		expectedResult []entity.SpecializationSalaryRange
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock)
	}{
		{
			name: "Успешное получение диапазонов зарплат специализаций",
			expectedResult: []entity.SpecializationSalaryRange{
				{
					ID:        1,
					Name:      "Backend Developer",
					MinSalary: 80000,
					MaxSalary: 200000,
					AvgSalary: 140000,
				},
				{
					ID:        2,
					Name:      "Frontend Developer",
					MinSalary: 70000,
					MaxSalary: 180000,
					AvgSalary: 125000,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Backend Developer", 80000, 200000, 140000).
					AddRow(2, "Frontend Developer", 70000, 180000, 125000)
				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Пустой список диапазонов зарплат",
			expectedResult: []entity.SpecializationSalaryRange{},
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
				fmt.Errorf("ошибка при получении списка вилок специализаций: %w", errors.New("database error")),
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
					AddRow(1, "Backend Developer", 80000, 200000, 140000).
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
					AddRow(1, "Backend Developer", 80000, 200000, 140000)
				rows.CloseError(errors.New("rows error"))
				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
		},
		{
			name: "Один результат с корректными данными",
			expectedResult: []entity.SpecializationSalaryRange{
				{
					ID:        3,
					Name:      "DevOps Engineer",
					MinSalary: 90000,
					MaxSalary: 220000,
					AvgSalary: 155000,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns).
					AddRow(3, "DevOps Engineer", 90000, 220000, 155000)
				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
		},
		{
			name: "Несколько специализаций с разными диапазонами",
			expectedResult: []entity.SpecializationSalaryRange{
				{
					ID:        1,
					Name:      "Backend Developer",
					MinSalary: 50000,
					MaxSalary: 300000,
					AvgSalary: 175000,
				},
				{
					ID:        4,
					Name:      "Data Scientist",
					MinSalary: 100000,
					MaxSalary: 250000,
					AvgSalary: 175000,
				},
				{
					ID:        2,
					Name:      "Frontend Developer",
					MinSalary: 45000,
					MaxSalary: 200000,
					AvgSalary: 122500,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Backend Developer", 50000, 300000, 175000).
					AddRow(4, "Data Scientist", 100000, 250000, 175000).
					AddRow(2, "Frontend Developer", 45000, 200000, 122500)
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
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock)

			repo := &SpecializationRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetSpecializationSalaries(ctx)

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
					require.Equal(t, expected.MinSalary, result[i].MinSalary)
					require.Equal(t, expected.MaxSalary, result[i].MaxSalary)
					require.Equal(t, expected.AvgSalary, result[i].AvgSalary)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
