package postgres

import (
	"ResuMatch/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVacancyRepository_Create(t *testing.T) {
	t.Parallel()

	columns := []string{
		"id", "title", "is_active", "employer_id", "specialization_id",
		"work_format", "employment", "schedule", "working_hours",
		"salary_from", "salary_to", "taxes_included", "experience",
		"description", "tasks", "requirements", "optional_requirements",
		"city", "created_at", "updated_at",
	}

	now := time.Now()

	testCases := []struct {
		name           string
		inputVacancy   *entity.Vacancy
		expectedResult *entity.Vacancy
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy)
	}{
		{
			name: "Успешное создание вакансии",
			inputVacancy: &entity.Vacancy{
				Title:                "Backend Developer",
				IsActive:             true,
				EmployerID:           1,
				SpecializationID:     2,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "flexible",
				WorkingHours:         8,
				SalaryFrom:           100000,
				SalaryTo:             150000,
				TaxesIncluded:        true,
				Experience:           "3+ years",
				Description:          "Backend development",
				Tasks:                "Develop APIs",
				Requirements:         "Go experience",
				OptionalRequirements: "Kubernetes knowledge",
				City:                 "Moscow",
			},
			expectedResult: &entity.Vacancy{
				ID:                   1,
				Title:                "Backend Developer",
				IsActive:             true,
				EmployerID:           1,
				SpecializationID:     2,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "flexible",
				WorkingHours:         8,
				SalaryFrom:           100000,
				SalaryTo:             150000,
				TaxesIncluded:        true,
				Experience:           "3+ years",
				Description:          "Backend development",
				Tasks:                "Develop APIs",
				Requirements:         "Go experience",
				OptionalRequirements: "Kubernetes knowledge",
				City:                 "Moscow",
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				query := regexp.QuoteMeta(`
                    INSERT INTO vacancy (
                        employer_id, title, specialization_id, work_format, 
                        employment, schedule, working_hours, salary_from, 
                        salary_to, taxes_included, experience, description, 
                        tasks, requirements, optional_requirements, city,
                        created_at, updated_at
                    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
                    RETURNING id, employer_id, title, is_active, specialization_id, 
                              work_format, employment, schedule, working_hours, 
                              salary_from, salary_to, taxes_included, experience, 
                              description, tasks, requirements, optional_requirements, 
                              city, created_at, updated_at
                `)
				mock.ExpectQuery(query).
					WithArgs(
						vacancy.EmployerID,
						vacancy.Title,
						vacancy.SpecializationID,
						vacancy.WorkFormat,
						vacancy.Employment,
						vacancy.Schedule,
						vacancy.WorkingHours,
						vacancy.SalaryFrom,
						vacancy.SalaryTo,
						vacancy.TaxesIncluded,
						vacancy.Experience,
						vacancy.Description,
						vacancy.Tasks,
						vacancy.Requirements,
						vacancy.OptionalRequirements,
						vacancy.City,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								1,
								vacancy.EmployerID,
								vacancy.Title,
								true,
								vacancy.SpecializationID,
								vacancy.WorkFormat,
								vacancy.Employment,
								vacancy.Schedule,
								vacancy.WorkingHours,
								vacancy.SalaryFrom,
								vacancy.SalaryTo,
								vacancy.TaxesIncluded,
								vacancy.Experience,
								vacancy.Description,
								vacancy.Tasks,
								vacancy.Requirements,
								vacancy.OptionalRequirements,
								vacancy.City,
								now,
								now,
							),
					)
			},
		},
		{
			name: "Ошибка - нарушение уникальности",
			inputVacancy: &entity.Vacancy{
				Title:                "Backend Developer",
				EmployerID:           1,
				SpecializationID:     2,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "flexible",
				WorkingHours:         8,
				SalaryFrom:           100000,
				SalaryTo:             150000,
				TaxesIncluded:        true,
				Experience:           "3+ years",
				Description:          "Backend development",
				Tasks:                "Develop APIs",
				Requirements:         "Go experience",
				OptionalRequirements: "Kubernetes knowledge",
				City:                 "Moscow",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("вакансия с такими параметрами уже существует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				query := regexp.QuoteMeta(`
                    INSERT INTO vacancy (
                        employer_id, title, specialization_id, work_format, 
                        employment, schedule, working_hours, salary_from, 
                        salary_to, taxes_included, experience, description, 
                        tasks, requirements, optional_requirements, city,
                        created_at, updated_at
                    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
                    RETURNING id, employer_id, title, is_active, specialization_id, 
                              work_format, employment, schedule, working_hours, 
                              salary_from, salary_to, taxes_included, experience, 
                              description, tasks, requirements, optional_requirements, 
                              city, created_at, updated_at
                `)
				pqErr := &pq.Error{
					Code: entity.PSQLUniqueViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						vacancy.EmployerID,
						vacancy.Title,
						vacancy.SpecializationID,
						vacancy.WorkFormat,
						vacancy.Employment,
						vacancy.Schedule,
						vacancy.WorkingHours,
						vacancy.SalaryFrom,
						vacancy.SalaryTo,
						vacancy.TaxesIncluded,
						vacancy.Experience,
						vacancy.Description,
						vacancy.Tasks,
						vacancy.Requirements,
						vacancy.OptionalRequirements,
						vacancy.City,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - обязательное поле отсутствует",
			inputVacancy: &entity.Vacancy{
				Title:                "Backend Developer",
				EmployerID:           1,
				SpecializationID:     2,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "flexible",
				WorkingHours:         8,
				SalaryFrom:           100000,
				SalaryTo:             150000,
				TaxesIncluded:        true,
				Experience:           "3+ years",
				Description:          "Backend development",
				Tasks:                "Develop APIs",
				Requirements:         "Go experience",
				OptionalRequirements: "Kubernetes knowledge",
				City:                 "Moscow",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				query := regexp.QuoteMeta(`
                    INSERT INTO vacancy (
                        employer_id, title, specialization_id, work_format, 
                        employment, schedule, working_hours, salary_from, 
                        salary_to, taxes_included, experience, description, 
                        tasks, requirements, optional_requirements, city,
                        created_at, updated_at
                    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
                    RETURNING id, employer_id, title, is_active, specialization_id, 
                              work_format, employment, schedule, working_hours, 
                              salary_from, salary_to, taxes_included, experience, 
                              description, tasks, requirements, optional_requirements, 
                              city, created_at, updated_at
                `)
				pqErr := &pq.Error{
					Code: entity.PSQLNotNullViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						vacancy.EmployerID,
						vacancy.Title,
						vacancy.SpecializationID,
						vacancy.WorkFormat,
						vacancy.Employment,
						vacancy.Schedule,
						vacancy.WorkingHours,
						vacancy.SalaryFrom,
						vacancy.SalaryTo,
						vacancy.TaxesIncluded,
						vacancy.Experience,
						vacancy.Description,
						vacancy.Tasks,
						vacancy.Requirements,
						vacancy.OptionalRequirements,
						vacancy.City,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - внутренняя ошибка базы данных",
			inputVacancy: &entity.Vacancy{
				Title:                "Backend Developer",
				EmployerID:           1,
				SpecializationID:     2,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "flexible",
				WorkingHours:         8,
				SalaryFrom:           100000,
				SalaryTo:             150000,
				TaxesIncluded:        true,
				Experience:           "3+ years",
				Description:          "Backend development",
				Tasks:                "Develop APIs",
				Requirements:         "Go experience",
				OptionalRequirements: "Kubernetes knowledge",
				City:                 "Moscow",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании вакансии: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				query := regexp.QuoteMeta(`
                    INSERT INTO vacancy (
                        employer_id, title, specialization_id, work_format, 
                        employment, schedule, working_hours, salary_from, 
                        salary_to, taxes_included, experience, description, 
                        tasks, requirements, optional_requirements, city,
                        created_at, updated_at
                    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
                    RETURNING id, employer_id, title, is_active, specialization_id, 
                              work_format, employment, schedule, working_hours, 
                              salary_from, salary_to, taxes_included, experience, 
                              description, tasks, requirements, optional_requirements, 
                              city, created_at, updated_at
                `)
				mock.ExpectQuery(query).
					WithArgs(
						vacancy.EmployerID,
						vacancy.Title,
						vacancy.SpecializationID,
						vacancy.WorkFormat,
						vacancy.Employment,
						vacancy.Schedule,
						vacancy.WorkingHours,
						vacancy.SalaryFrom,
						vacancy.SalaryTo,
						vacancy.TaxesIncluded,
						vacancy.Experience,
						vacancy.Description,
						vacancy.Tasks,
						vacancy.Requirements,
						vacancy.OptionalRequirements,
						vacancy.City,
					).
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

			tc.setupMock(mock, tc.inputVacancy)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.Create(ctx, tc.inputVacancy)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.Title, result.Title)
				require.Equal(t, tc.expectedResult.IsActive, result.IsActive)
				require.Equal(t, tc.expectedResult.EmployerID, result.EmployerID)
				require.Equal(t, tc.expectedResult.SpecializationID, result.SpecializationID)
				require.Equal(t, tc.expectedResult.WorkFormat, result.WorkFormat)
				require.Equal(t, tc.expectedResult.Employment, result.Employment)
				require.Equal(t, tc.expectedResult.Schedule, result.Schedule)
				require.Equal(t, tc.expectedResult.WorkingHours, result.WorkingHours)
				require.Equal(t, tc.expectedResult.SalaryFrom, result.SalaryFrom)
				require.Equal(t, tc.expectedResult.SalaryTo, result.SalaryTo)
				require.Equal(t, tc.expectedResult.TaxesIncluded, result.TaxesIncluded)
				require.Equal(t, tc.expectedResult.Experience, result.Experience)
				require.Equal(t, tc.expectedResult.Description, result.Description)
				require.Equal(t, tc.expectedResult.Tasks, result.Tasks)
				require.Equal(t, tc.expectedResult.Requirements, result.Requirements)
				require.Equal(t, tc.expectedResult.OptionalRequirements, result.OptionalRequirements)
				require.Equal(t, tc.expectedResult.City, result.City)
				require.False(t, result.CreatedAt.IsZero())
				require.False(t, result.UpdatedAt.IsZero())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestVacancyRepository_GetByID(t *testing.T) {
	t.Parallel()

	columns := []string{
		"id", "title", "employer_id", "specialization_id", "work_format",
		"employment", "schedule", "working_hours", "salary_from", "salary_to",
		"taxes_included", "experience", "description", "tasks", "requirements",
		"optional_requirements", "city", "created_at", "updated_at",
	}

	now := time.Now()

	testCases := []struct {
		name           string
		id             int
		expectedResult *entity.VacancyResponse
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name: "Успешное получение вакансии",
			id:   1,
			expectedResult: &entity.VacancyResponse{
				ID:                   1,
				Title:                "Разработчик Go",
				EmployerID:           101,
				Specialization:       "Backend Development",
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "flexible",
				WorkingHours:         40,
				SalaryFrom:           100000,
				SalaryTo:             150000,
				TaxesIncluded:        true,
				Experience:           "no_experience",
				Description:          "Работа с высоконагруженными системами.",
				Tasks:                "Разработка новых сервисов.",
				Requirements:         "Go, PostgreSQL.",
				OptionalRequirements: "Docker, Kubernetes.",
				City:                 "Москва",
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				queryVacancy := regexp.QuoteMeta(`
					SELECT id, title, employer_id, specialization_id, work_format, employment, 
					schedule, working_hours, salary_from, salary_to, taxes_included, experience, 
					description, tasks, requirements, optional_requirements, city, created_at, updated_at
					FROM vacancy WHERE id = $1
				`)
				mock.ExpectQuery(queryVacancy).
					WithArgs(id).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								1, "Разработчик Go", 101, "Backend Development", "remote", "full_time", "flexible",
								40, 100000, 150000, true, "no_experience", "Работа с высоконагруженными системами.",
								"Разработка новых сервисов.", "Go, PostgreSQL.", "Docker, Kubernetes.",
								"Москва", time.Now(), time.Now(),
							),
					)

				querySkills := regexp.QuoteMeta(`
					SELECT s.id, s.name FROM skill s 
					JOIN vacancy_skill vs ON s.id = vs.skill_id 
					WHERE vs.vacancy_id = $1
				`)
				mock.ExpectQuery(querySkills).
					WithArgs(id).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name"}).
							AddRow(1, "Go").
							AddRow(2, "PostgreSQL"),
					)
			},
		},
		{
			name:           "Вакансия не найдена",
			id:             999,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("вакансия с id=999 не найдена"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
                    SELECT id, title, employer_id, specialization_id, work_format, employment, 
                    schedule, working_hours, salary_from, salary_to, taxes_included, experience, 
                    description, tasks, requirements, optional_requirements, city, created_at, updated_at 
                    FROM vacancy 
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
				fmt.Errorf("ошибка при получении вакансии: database failure"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
                    SELECT id, title, employer_id, specialization_id, work_format, employment, 
                    schedule, working_hours, salary_from, salary_to, taxes_included, experience, 
                    description, tasks, requirements, optional_requirements, city, created_at, updated_at 
                    FROM vacancy 
                    WHERE id = $1
                `)
				mock.ExpectQuery(query).
					WithArgs(id).
					WillReturnError(errors.New("database failure"))
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

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetByID(ctx, tc.id)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.Title, result.Title)
				require.Equal(t, tc.expectedResult.EmployerID, result.EmployerID)
				require.Equal(t, tc.expectedResult.Specialization, result.Specialization)
				require.Equal(t, tc.expectedResult.WorkFormat, result.WorkFormat)
				require.Equal(t, tc.expectedResult.Employment, result.Employment)
				require.Equal(t, tc.expectedResult.Schedule, result.Schedule)
				require.Equal(t, tc.expectedResult.WorkingHours, result.WorkingHours)
				require.Equal(t, tc.expectedResult.SalaryFrom, result.SalaryFrom)
				require.Equal(t, tc.expectedResult.SalaryTo, result.SalaryTo)
				require.Equal(t, tc.expectedResult.TaxesIncluded, result.TaxesIncluded)
				require.Equal(t, tc.expectedResult.Experience, result.Experience)
				require.Equal(t, tc.expectedResult.Description, result.Description)
				require.Equal(t, tc.expectedResult.Tasks, result.Tasks)
				require.Equal(t, tc.expectedResult.Requirements, result.Requirements)
				require.Equal(t, tc.expectedResult.OptionalRequirements, result.OptionalRequirements)
				require.Equal(t, tc.expectedResult.City, result.City)
				require.Equal(t, tc.expectedResult.CreatedAt.Unix(), result.CreatedAt.Unix())
				require.Equal(t, tc.expectedResult.UpdatedAt.Unix(), result.UpdatedAt.Unix())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &VacancyRepository{DB: db}

	tests := []struct {
		name        string
		mock        func()
		vacancy     *entity.Vacancy
		expected    *entity.Vacancy
		expectedErr error
	}{
		{
			name: "Success",
			mock: func() {
				mock.ExpectQuery("UPDATE vacancy").
					WithArgs(
						"Title", true, 1, "format", "employment",
						"schedule", 8, 1000, 2000, true,
						"experience", "desc", "tasks", "req", "opt",
						"city", 1, 1,
					).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "employer_id", "title", "is_active", "specialization_id",
						"work_format", "employment", "schedule", "working_hours",
						"salary_from", "salary_to", "taxes_included", "experience",
						"description", "tasks", "requirements", "optional_requirements",
						"city", "created_at", "updated_at",
					}).AddRow(
						1, 1, "Title", true, 1,
						"format", "employment", "schedule", 8,
						1000, 2000, true, "experience",
						"desc", "tasks", "req", "opt",
						"city", time.Now(), time.Now(),
					))
			},
			vacancy: &entity.Vacancy{
				ID:                   1,
				Title:                "Title",
				IsActive:             true,
				SpecializationID:     1,
				WorkFormat:           "format",
				Employment:           "employment",
				Schedule:             "schedule",
				WorkingHours:         8,
				SalaryFrom:           1000,
				SalaryTo:             2000,
				TaxesIncluded:        true,
				Experience:           "experience",
				Description:          "desc",
				Tasks:                "tasks",
				Requirements:         "req",
				OptionalRequirements: "opt",
				City:                 "city",
				EmployerID:           1,
			},
			expected: &entity.Vacancy{
				ID:                   1,
				EmployerID:           1,
				Title:                "Title",
				IsActive:             true,
				SpecializationID:     1,
				WorkFormat:           "format",
				Employment:           "employment",
				Schedule:             "schedule",
				WorkingHours:         8,
				SalaryFrom:           1000,
				SalaryTo:             2000,
				TaxesIncluded:        true,
				Experience:           "experience",
				Description:          "desc",
				Tasks:                "tasks",
				Requirements:         "req",
				OptionalRequirements: "opt",
				City:                 "city",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.Update(context.Background(), tt.vacancy)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_Delete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		id          int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name:        "Успешное удаление вакансии",
			id:          1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM vacancy
                    WHERE id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(0, 1)) // ✅ 1 строка удалена
			},
		},
		{
			name: "Вакансия не найдена",
			id:   999,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("вакансия с id=999 не найдена"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM vacancy
                    WHERE id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(0, 0)) // ✅ 0 строк удалено (не найдено)
			},
		},
		{
			name: "Ошибка при выполнении запроса",
			id:   2,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении вакансии: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM vacancy
                    WHERE id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(id).
					WillReturnError(errors.New("database error")) // ❌ Ошибка БД
			},
		},
		{
			name: "Ошибка при получении количества затронутых строк",
			id:   3,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении количества затронутых строк: %w", errors.New("rows affected error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM vacancy
                    WHERE id = $1
                `)
				result := sqlmock.NewErrorResult(errors.New("rows affected error"))
				mock.ExpectExec(query).
					WithArgs(id).
					WillReturnResult(result) // ❌ Ошибка получения `RowsAffected`
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

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			err = repo.Delete(ctx, tc.id)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_GetSkillsByVacancyID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &VacancyRepository{DB: db}

	tests := []struct {
		name        string
		mock        func()
		id          int
		expected    []string
		expectedErr error
	}{
		{
			name: "Success",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Skill1").
					AddRow(2, "Skill2")
				mock.ExpectQuery("SELECT s.id, s.name").
					WithArgs(1).
					WillReturnRows(rows)
			},
			id:       1,
			expected: []string{"Skill1", "Skill2"},
		},
		{
			name: "No Skills",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery("SELECT s.id, s.name").
					WithArgs(1).
					WillReturnRows(rows)
			},
			id:       1,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.GetSkillsByVacancyID(context.Background(), tt.id)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_ResponseExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &VacancyRepository{DB: db}

	tests := []struct {
		name        string
		mock        func()
		vacancyID   int
		applicantID int
		expected    bool
		expectedErr error
	}{
		{
			name: "Exists",
			mock: func() {
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			vacancyID:   1,
			applicantID: 1,
			expected:    true,
		},
		{
			name: "Not Exists",
			mock: func() {
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			vacancyID:   1,
			applicantID: 1,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.ResponseExists(context.Background(), tt.vacancyID, tt.applicantID)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_CreateResponse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &VacancyRepository{DB: db}

	tests := []struct {
		name             string
		mock             func()
		vacancyID        int
		applicantID      int
		expectedResumeID int
		expectedErr      error
	}{
		{
			name: "Success",
			mock: func() {
				// Mock resume query
				mock.ExpectQuery("SELECT id FROM resume").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Mock insert response
				mock.ExpectExec("INSERT INTO vacancy_response").
					WithArgs(1, 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			vacancyID:        1,
			applicantID:      1,
			expectedResumeID: 1,
		},
		{
			name: "No Resume",
			mock: func() {
				mock.ExpectQuery("SELECT id FROM resume").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			vacancyID:   1,
			applicantID: 1,
			expectedErr: entity.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.CreateResponse(context.Background(), tt.vacancyID, tt.applicantID)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_AddSkills(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &VacancyRepository{DB: db}

	tests := []struct {
		name        string
		mock        func()
		vacancyID   int
		skillIDs    []int
		expectedErr error
	}{
		{
			name: "Success",
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO vacancy_skill").
					ExpectExec().
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			vacancyID: 1,
			skillIDs:  []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.AddSkills(context.Background(), tt.vacancyID, tt.skillIDs)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_CreateSkillIfNotExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &VacancyRepository{DB: db}

	tests := []struct {
		name        string
		mock        func()
		skillName   string
		expectedID  int
		expectedErr error
	}{
		{
			name: "Skill Exists",
			mock: func() {
				mock.ExpectQuery("SELECT id FROM skill").
					WithArgs("Skill1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			skillName:  "Skill1",
			expectedID: 1,
		},
		{
			name: "Create New Skill",
			mock: func() {
				// First check - not found
				mock.ExpectQuery("SELECT id FROM skill").
					WithArgs("NewSkill").
					WillReturnError(sql.ErrNoRows)

				// Then create
				mock.ExpectQuery("INSERT INTO skill").
					WithArgs("NewSkill").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
			skillName:  "NewSkill",
			expectedID: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.CreateSkillIfNotExists(context.Background(), tt.skillName)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_GetVacancyResponsesByVacancyID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &VacancyRepository{DB: db}

	appliedAt := time.Now()

	tests := []struct {
		name        string
		mock        func()
		vacancyID   int
		expected    []entity.VacancyResponses
		expectedErr error
	}{
		{
			name: "Success",
			mock: func() {
				mock.ExpectQuery(`SELECT r.id, r.vacancy_id, r.applicant_id, r.applied_at FROM vacancy_response WHERE vacancy_id = \$1 ORDER BY applied_at DESC`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "vacancy_id", "applicant_id", "applied_at"}).
						AddRow(1, 1, 101, appliedAt).
						AddRow(2, 1, 102, appliedAt.Add(-time.Hour)))
			},
			vacancyID: 1,
			expected: []entity.VacancyResponses{
				{
					ID:          1,
					VacancyID:   1,
					ApplicantID: 101,
					AppliedAt:   appliedAt,
				},
				{
					ID:          2,
					VacancyID:   1,
					ApplicantID: 102,
					AppliedAt:   appliedAt.Add(-time.Hour),
				},
			},
		},
		{
			name: "No Responses",
			mock: func() {
				mock.ExpectQuery(`SELECT r.id, r.vacancy_id, r.applicant_id, r.applied_at FROM vacancy_response WHERE vacancy_id = \$1 ORDER BY applied_at DESC`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "vacancy_id", "applicant_id", "applied_at"}))
			},
			vacancyID: 1,
			expected:  []entity.VacancyResponses{},
		},
		{
			name: "Database Error",
			mock: func() {
				mock.ExpectQuery(`SELECT r.id, r.vacancy_id, r.applicant_id, r.applied_at FROM vacancy_response WHERE vacancy_id = \$1 ORDER BY applied_at DESC`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			vacancyID:   1,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.GetVacancyResponsesByVacancyID(context.Background(), tt.vacancyID)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &VacancyRepository{DB: db}

	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name        string
		mock        func()
		expected    []*entity.Vacancy
		expectedErr error
	}{
		{
			name: "Success",
			mock: func() {
				mock.ExpectQuery(`SELECT id, title, is_active, employer_id, specialization_id, work_format, employment, schedule, working_hours, salary_from, salary_to, taxes_included, experience, description, tasks, requirements, optional_requirements, city, created_at, updated_at FROM vacancy ORDER BY updated_at DESC LIMIT 100`).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "title", "is_active", "employer_id", "specialization_id",
						"work_format", "employment", "schedule", "working_hours",
						"salary_from", "salary_to", "taxes_included", "experience",
						"description", "tasks", "requirements", "optional_requirements",
						"city", "created_at", "updated_at",
					}).AddRow(
						1, "Backend Developer", true, 1, 1,
						"remote", "full_time", "flexible", 8,
						100000, 150000, true, "3+ years",
						"Backend development", "Develop APIs", "Go experience",
						"Kubernetes knowledge", "Moscow", createdAt, updatedAt,
					).AddRow(
						2, "Frontend Developer", true, 2, 2,
						"hybrid", "full_time", "9-to-5", 8,
						90000, 120000, false, "2+ years",
						"Frontend development", "Develop UI", "React experience",
						"TypeScript knowledge", "Saint Petersburg", createdAt, updatedAt,
					))
			},
			expected: []*entity.Vacancy{
				{
					ID:                   1,
					Title:                "Backend Developer",
					IsActive:             true,
					EmployerID:           1,
					SpecializationID:     1,
					WorkFormat:           "remote",
					Employment:           "full_time",
					Schedule:             "flexible",
					WorkingHours:         8,
					SalaryFrom:           100000,
					SalaryTo:             150000,
					TaxesIncluded:        true,
					Experience:           "3+ years",
					Description:          "Backend development",
					Tasks:                "Develop APIs",
					Requirements:         "Go experience",
					OptionalRequirements: "Kubernetes knowledge",
					City:                 "Moscow",
					CreatedAt:            createdAt,
					UpdatedAt:            updatedAt,
				},
				{
					ID:                   2,
					Title:                "Frontend Developer",
					IsActive:             true,
					EmployerID:           2,
					SpecializationID:     2,
					WorkFormat:           "hybrid",
					Employment:           "full_time",
					Schedule:             "9-to-5",
					WorkingHours:         8,
					SalaryFrom:           90000,
					SalaryTo:             120000,
					TaxesIncluded:        false,
					Experience:           "2+ years",
					Description:          "Frontend development",
					Tasks:                "Develop UI",
					Requirements:         "React experience",
					OptionalRequirements: "TypeScript knowledge",
					City:                 "Saint Petersburg",
					CreatedAt:            createdAt,
					UpdatedAt:            updatedAt,
				},
			},
		},
		{
			name: "No Vacancies",
			mock: func() {
				mock.ExpectQuery(`SELECT id, title, is_active, employer_id, specialization_id, work_format, employment, schedule, working_hours, salary_from, salary_to, taxes_included, experience, description, tasks, requirements, optional_requirements, city, created_at, updated_at FROM vacancy ORDER BY updated_at DESC LIMIT 100`).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "title", "is_active", "employer_id", "specialization_id",
						"work_format", "employment", "schedule", "working_hours",
						"salary_from", "salary_to", "taxes_included", "experience",
						"description", "tasks", "requirements", "optional_requirements",
						"city", "created_at", "updated_at",
					}))
			},
			expected:    []*entity.Vacancy{},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.GetAll(context.Background())
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
