package postgres

import (
	"ResuMatch/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	// "reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"

	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVacancyRepository_Create(t *testing.T) {
	t.Parallel()

	now := time.Now()

	columns := []string{
		"id", "employer_id", "title", "is_active", "specialization_id",
		"work_format", "employment", "schedule", "working_hours",
		"salary_from", "salary_to", "taxes_included", "experience",
		"description", "tasks", "requirements", "optional_requirements",
		"city", "created_at", "updated_at",
	}

	query := regexp.QuoteMeta(`
		INSERT INTO vacancy (
			employer_id, title, specialization_id, work_format, employment,
			schedule, working_hours, salary_from, salary_to, taxes_included,
			experience, description, tasks, requirements, optional_requirements,
			city, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
		RETURNING id, employer_id, title, is_active, specialization_id, work_format,
			employment, schedule, working_hours, salary_from, salary_to,
			taxes_included, experience, description, tasks,
			requirements, optional_requirements, city, created_at, updated_at
	`)

	testCases := []struct {
		name           string
		inputVacancy   *entity.Vacancy
		expectedResult *entity.Vacancy
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy)
	}{
		{
			name: "Успешное создание вакансии с полными данными",
			inputVacancy: &entity.Vacancy{
				EmployerID:           1,
				Title:                "Senior Go Developer",
				SpecializationID:     2,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
			},
			expectedResult: &entity.Vacancy{
				ID:                   1,
				EmployerID:           1,
				Title:                "Senior Go Developer",
				IsActive:             false,
				SpecializationID:     2,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
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
								false,
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
			name: "Успешное создание вакансии с минимальными данными",
			inputVacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Junior Developer",
				SpecializationID: 3,
				WorkFormat:       "office",
				Employment:       "full_time",
				Schedule:         "5/2",
				WorkingHours:     40,
				SalaryFrom:       100000,
				SalaryTo:         150000,
				Experience:       "no_experience",
				Description:      "Разработка веб-приложений",
				City:             "Санкт-Петербург",
			},
			expectedResult: &entity.Vacancy{
				ID:               2,
				EmployerID:       1,
				Title:            "Junior Developer",
				IsActive:         false,
				SpecializationID: 3,
				WorkFormat:       "office",
				Employment:       "full_time",
				Schedule:         "5/2",
				WorkingHours:     40,
				SalaryFrom:       100000,
				SalaryTo:         150000,
				Experience:       "no_experience",
				Description:      "Разработка веб-приложений",
				City:             "Санкт-Петербург",
				CreatedAt:        now,
				UpdatedAt:        now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
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
								2,
								vacancy.EmployerID,
								vacancy.Title,
								false,
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
			name: "Ошибка - пустое название вакансии",
			inputVacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "",
				SpecializationID: 2,
				WorkFormat:       "remote",
				Employment:       "full_time",
				Schedule:         "5/2",
				WorkingHours:     40,
				SalaryFrom:       200000,
				SalaryTo:         300000,
				Experience:       "3_6_years",
				Description:      "Разработка систем",
				City:             "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				// No DB query expected since validation fails
			},
		},
		{
			name: "Ошибка - SpecializationID равен нулю",
			inputVacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Senior Developer",
				SpecializationID: 0,
				WorkFormat:       "remote",
				Employment:       "full_time",
				Schedule:         "5/2",
				WorkingHours:     40,
				SalaryFrom:       200000,
				SalaryTo:         300000,
				Experience:       "3_6_years",
				Description:      "Разработка систем",
				City:             "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				// No DB query expected since validation fails
			},
		},
		{
			name: "Ошибка - нарушение уникальности",
			inputVacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Senior Go Developer",
				SpecializationID: 2,
				WorkFormat:       "remote",
				Employment:       "full_time",
				Schedule:         "5/2",
				WorkingHours:     40,
				SalaryFrom:       200000,
				SalaryTo:         300000,
				Experience:       "3_6_years",
				Description:      "Разработка систем",
				City:             "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("вакансия с такими параметрами уже существует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
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
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
			},
		},
		{
			name: "Ошибка - обязательное поле отсутствует",
			inputVacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Senior Go Developer",
				SpecializationID: 2,
				WorkFormat:       "remote",
				Employment:       "full_time",
				Schedule:         "5/2",
				WorkingHours:     40,
				SalaryFrom:       200000,
				SalaryTo:         300000,
				Experience:       "3_6_years",
				Description:      "Разработка систем",
				City:             "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
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
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
			},
		},
		{
			name: "Ошибка - неверный тип данных",
			inputVacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Senior Go Developer",
				SpecializationID: 2,
				WorkFormat:       "remote",
				Employment:       "full_time",
				Schedule:         "5/2",
				WorkingHours:     40,
				SalaryFrom:       200000,
				SalaryTo:         300000,
				Experience:       "3_6_years",
				Description:      "Разработка систем",
				City:             "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
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
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
			},
		},
		{
			name: "Ошибка - нарушение проверки данных",
			inputVacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Senior Go Developer",
				SpecializationID: 2,
				WorkFormat:       "remote",
				Employment:       "full_time",
				Schedule:         "5/2",
				WorkingHours:     40,
				SalaryFrom:       200000,
				SalaryTo:         300000,
				Experience:       "3_6_years",
				Description:      "Разработка систем",
				City:             "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
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
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
			},
		},
		{
			name: "Ошибка - внутренняя ошибка базы данных",
			inputVacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Senior Go Developer",
				SpecializationID: 2,
				WorkFormat:       "remote",
				Employment:       "full_time",
				Schedule:         "5/2",
				WorkingHours:     40,
				SalaryFrom:       200000,
				SalaryTo:         300000,
				Experience:       "3_6_years",
				Description:      "Разработка систем",
				City:             "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании вакансии: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
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
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.inputVacancy)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.Create(ctx, tc.inputVacancy)

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
				require.Equal(t, tc.expectedResult.EmployerID, result.EmployerID)
				require.Equal(t, tc.expectedResult.Title, result.Title)
				require.Equal(t, tc.expectedResult.IsActive, result.IsActive)
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

func TestVacancyRepository_AddSkills(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		INSERT INTO vacancy_skill (vacancy_id, skill_id)
		VALUES ($1, $2)
	`)

	testCases := []struct {
		name        string
		vacancyID   int
		skillIDs    []int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int)
	}{
		{
			name:        "Успешное добавление нескольких навыков",
			vacancyID:   1,
			skillIDs:    []int{1, 2, 3},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(query)
				for _, skillID := range skillIDs {
					stmt.ExpectExec().
						WithArgs(vacancyID, skillID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
				mock.ExpectCommit()
			},
		},
		{
			name:        "Успешное добав | дубликатами",
			vacancyID:   1,
			skillIDs:    []int{1, 1, 2},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(query)
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[0]).
					WillReturnResult(sqlmock.NewResult(1, 1))
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[1]).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[2]).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		// {
		// 	name:        "Пустой список навыков",
		// 	vacancyID:   1,
		// 	skillIDs:    []int{},
		// 	expectedErr: nil,
		// 	setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
		// 		mock.ExpectBegin()
		// 		mock.ExpectCommit()
		// 	},
		// },
		{
			name:      "Ошибка при начале транзакции",
			vacancyID: 1,
			skillIDs:  []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при начале транзакции для добавления навыков: %w", errors.New("transaction begin error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin().WillReturnError(errors.New("transaction begin error"))
			},
		},
		{
			name:      "Ошибка при подготовке запроса",
			vacancyID: 1,
			skillIDs:  []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при подготовке запроса для добавления навыков: %w", errors.New("prepare error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).WillReturnError(errors.New("prepare error"))
				mock.ExpectRollback()
			},
		},
		{
			name:      "Ошибка - обязательное поле отсутствует",
			vacancyID: 1,
			skillIDs:  []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(query)
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[0]).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:      "Ошибка - неверный тип данных",
			vacancyID: 1,
			skillIDs:  []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(query)
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[0]).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:      "Ошибка - нарушение проверки данных",
			vacancyID: 1,
			skillIDs:  []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(query)
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[0]).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:      "Ошибка при выполнении запроса",
			vacancyID: 1,
			skillIDs:  []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении навыка к вакансии: %w", errors.New("execution error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(query)
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[0]).
					WillReturnError(errors.New("execution error"))
				mock.ExpectRollback()
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

			tc.setupMock(mock, tc.vacancyID, tc.skillIDs)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			err = repo.AddSkills(ctx, tc.vacancyID, tc.skillIDs)

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

func TestVacancyRepository_AddCity(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		INSERT INTO vacancy_city (vacancy_id, city_id)
		VALUES ($1, $2)
	`)

	testCases := []struct {
		name        string
		vacancyID   int
		cityIDs     []int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, vacancyID int, cityIDs []int)
	}{
		{
			name:        "Успешное добавление одного города",
			vacancyID:   1,
			cityIDs:     []int{10},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, cityIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).
					WillReturnCloseError(nil)
				mock.ExpectExec(query).
					WithArgs(vacancyID, cityIDs[0]).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:        "Успешное добавление нескольких городов",
			vacancyID:   1,
			cityIDs:     []int{10, 20, 30},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, cityIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).
					WillReturnCloseError(nil)
				for _, cityID := range cityIDs {
					mock.ExpectExec(query).
						WithArgs(vacancyID, cityID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
				mock.ExpectCommit()
			},
		},
		{
			name:        "Успешное добавление с игнорированием дубликатов",
			vacancyID:   1,
			cityIDs:     []int{10, 10, 20},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, cityIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).
					WillReturnCloseError(nil)
				mock.ExpectExec(query).
					WithArgs(vacancyID, 10).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(query).
					WithArgs(vacancyID, 10).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectExec(query).
					WithArgs(vacancyID, 20).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:        "Пустой список городов",
			vacancyID:   1,
			cityIDs:     []int{},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, cityIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).
					WillReturnCloseError(nil)
				mock.ExpectCommit()
			},
		},
		{
			name:      "Ошибка при начале транзакции",
			vacancyID: 1,
			cityIDs:   []int{10},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при начале транзакции для добавления городов: %w", errors.New("transaction begin error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, cityIDs []int) {
				mock.ExpectBegin().
					WillReturnError(errors.New("transaction begin error"))
			},
		},
		{
			name:      "Ошибка при подготовке запроса",
			vacancyID: 1,
			cityIDs:   []int{10},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при подготовке запроса для добавления городов: %w", errors.New("prepare error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, cityIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).
					WillReturnError(errors.New("prepare error"))
				mock.ExpectRollback()
			},
		},
		{
			name:      "Ошибка - обязательное поле отсутствует",
			vacancyID: 1,
			cityIDs:   []int{10},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, cityIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).
					WillReturnCloseError(nil)
				mock.ExpectExec(query).
					WithArgs(vacancyID, cityIDs[0]).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:      "Ошибка - неверный тип данных",
			vacancyID: 1,
			cityIDs:   []int{10},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, cityIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).
					WillReturnCloseError(nil)
				mock.ExpectExec(query).
					WithArgs(vacancyID, cityIDs[0]).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:      "Ошибка - нарушение проверки данных",
			vacancyID: 1,
			cityIDs:   []int{10},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, cityIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).
					WillReturnCloseError(nil)
				mock.ExpectExec(query).
					WithArgs(vacancyID, cityIDs[0]).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:      "Ошибка - внутренняя ошибка при выполнении запроса",
			vacancyID: 1,
			cityIDs:   []int{10},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении города к вакансии: %w", errors.New("exec error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, cityIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).
					WillReturnCloseError(nil)
				mock.ExpectExec(query).
					WithArgs(vacancyID, cityIDs[0]).
					WillReturnError(errors.New("exec error"))
				mock.ExpectRollback()
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

			tc.setupMock(mock, tc.vacancyID, tc.cityIDs)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			err = repo.AddCity(ctx, tc.vacancyID, tc.cityIDs)

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

func TestVacancyRepository_CreateSkillIfNotExists(t *testing.T) {
	t.Parallel()
	selectQuery := regexp.QuoteMeta(`
        SELECT id
        FROM skill
        WHERE name = $1
    `)

	insertQuery := regexp.QuoteMeta(`
        INSERT INTO skill (name)
        VALUES ($1)
        RETURNING id
    `)

	testCases := []struct {
		name        string
		skillName   string
		expectedID  int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, skillName string)
	}{
		{
			name:        "Успешное получение существующего навыка",
			skillName:   "Go",
			expectedID:  1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name:        "Успешное создание нового навыка",
			skillName:   "Python",
			expectedID:  2,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(skillName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
		},
		{
			name:        "Успешное создание с обработкой конфликта уникальности",
			skillName:   "Java",
			expectedID:  3,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(skillName).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
			},
		},
		{
			name:       "Ошибка при проверке существования навыка",
			skillName:  "Rust",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования навыка: %w", errors.New("select error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(errors.New("select error"))
			},
		},
		{
			name:       "Ошибка при создании навыка",
			skillName:  "C++",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании навыка: %w", errors.New("insert error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(skillName).
					WillReturnError(errors.New("insert error"))
			},
		},
		{
			name:       "Ошибка при получении ID после конфликта уникальности",
			skillName:  "JavaScript",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении ID навыка после конфликта: %w", errors.New("retry select error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(skillName).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(errors.New("retry select error"))
			},
		},
		{
			name:       "Пустое имя навыка",
			skillName:  "",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании навыка: %w", errors.New("empty skill name error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(skillName).
					WillReturnError(errors.New("empty skill name error"))
			},
		},
		{
			name:       "Ошибка - нарушение ограничения not-null при создании",
			skillName:  "Kotlin",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании навыка: %w", &pq.Error{Code: entity.PSQLNotNullViolation}),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(skillName).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
			},
		},
		{
			name:       "Ошибка - нарушение ограничения данных при создании",
			skillName:  "SQL",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании навыка: %w", &pq.Error{Code: entity.PSQLCheckViolation}),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(skillName).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
			},
		},
		{
			name:        "Успешное создание нового навыка",
			skillName:   "Go Programming",
			expectedID:  1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				queryCheck := regexp.QuoteMeta(`SELECT id FROM skill WHERE name = $1`)
				mock.ExpectQuery(queryCheck).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)

				queryInsert := regexp.QuoteMeta(`INSERT INTO skill (name) VALUES ($1) RETURNING id`)
				mock.ExpectQuery(queryInsert).
					WithArgs(skillName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name:        "Навык уже существует",
			skillName:   "Python Programming",
			expectedID:  2,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				queryCheck := regexp.QuoteMeta(`SELECT id FROM skill WHERE name = $1`)
				mock.ExpectQuery(queryCheck).
					WithArgs(skillName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
		},
		{
			name:        "Ошибка при проверке существования навыка",
			skillName:   "JavaScript",
			expectedID:  0,
			expectedErr: entity.NewError(entity.ErrInternal, fmt.Errorf("ошибка при проверке существования навыка: database error")),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				queryCheck := regexp.QuoteMeta(`SELECT id FROM skill WHERE name = $1`)
				mock.ExpectQuery(queryCheck).
					WithArgs(skillName).
					WillReturnError(fmt.Errorf("database error"))
			},
		},
		// {
		// 	name:        "Ошибка при создании навыка - уникальное нарушение",
		// 	skillName:   "Ruby Programming",
		// 	expectedID:  3,
		// 	expectedErr: nil,
		// 	setupMock: func(mock sqlmock.Sqlmock, skillName string) {
		// 		queryCheck := regexp.QuoteMeta(`SELECT id FROM skill WHERE name = $1`)
		// 		mock.ExpectQuery(queryCheck).
		// 			WithArgs(skillName).
		// 			WillReturnError(sql.ErrNoRows)

		// 		queryInsert := regexp.QuoteMeta(`INSERT INTO skill (name) VALUES ($1) RETURNING id`)
		// 		mock.ExpectQuery(queryInsert).
		// 			WithArgs(skillName).
		// 			WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
		// 		mock.ExpectQuery(selectQuery).
		// 			WithArgs(skillName).
		// 			WillReturnError(errors.New("retry select error"))
		// 	},
		// },
		{
			name:       "Пустое имя навыка",
			skillName:  "",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании навыка: %w", errors.New("empty skill name error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(skillName).
					WillReturnError(errors.New("empty skill name error"))
			},
		},
		{
			name:       "Ошибка - нарушение ограничения not-null при создании",
			skillName:  "Kotlin",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании навыка: %w", &pq.Error{Code: entity.PSQLNotNullViolation}),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(skillName).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
			},
		},
		{
			name:       "Ошибка - нарушение ограничения данных при создании",
			skillName:  "SQL",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании навыка: %w", &pq.Error{Code: entity.PSQLCheckViolation}),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(skillName).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
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

			tc.setupMock(mock, tc.skillName)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			id, err := repo.CreateSkillIfNotExists(ctx, tc.skillName)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Equal(t, tc.expectedID, id)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, id)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_GetByID(t *testing.T) {
	t.Parallel()

	now := time.Now()

	columns := []string{
		"id", "title", "employer_id", "specialization_id",
		"work_format", "employment", "schedule", "working_hours",
		"salary_from", "salary_to", "taxes_included", "experience",
		"description", "tasks", "requirements", "optional_requirements",
		"city", "created_at", "updated_at",
	}

	query := regexp.QuoteMeta(`
        SELECT
            id,
            title,
            employer_id,
            specialization_id,
            work_format,
            employment,
            schedule,
            working_hours,
            salary_from,
            salary_to,
            taxes_included,
            experience,
            description,
            tasks,
            requirements,
            optional_requirements,
			city,
			created_at,
			updated_at
        FROM vacancy
        WHERE id = $1
    `)

	testCases := []struct {
		name           string
		vacancyID      int
		expectedResult *entity.Vacancy
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, vacancyID int)
	}{
		{
			name:      "Успешное получение вакансии",
			vacancyID: 1,
			expectedResult: &entity.Vacancy{
				ID:                   1,
				Title:                "Senior Go Developer",
				EmployerID:           2,
				SpecializationID:     3,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								1,
								"Senior Go Developer",
								2,
								3,
								"remote",
								"full_time",
								"5/2",
								40,
								200000,
								300000,
								true,
								"3_6_years",
								"Разработка высоконагруженных систем",
								"Писать код, проводить код-ревью",
								"Знание Go, PostgreSQL",
								"Опыт с Kubernetes",
								"Москва",
								now,
								now,
							),
					)
			},
		},
		{
			name:           "Вакансия не найдена",
			vacancyID:      999,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("вакансия с id=%d не найдена", 999),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка базы данных",
			vacancyID:      1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении вакансии: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
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
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.vacancyID)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetByID(ctx, tc.vacancyID)

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
				require.Equal(t, tc.expectedResult.Title, result.Title)
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
				require.Equal(t, tc.expectedResult.CreatedAt, result.CreatedAt)
				require.Equal(t, tc.expectedResult.UpdatedAt, result.UpdatedAt)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_Update(t *testing.T) {
	t.Parallel()

	now := time.Now()

	columns := []string{
		"id", "employer_id", "title", "specialization_id",
		"work_format", "employment", "schedule", "working_hours",
		"salary_from", "salary_to", "taxes_included", "experience",
		"description", "tasks", "requirements", "optional_requirements",
		"city", "created_at", "updated_at",
	}

	query := regexp.QuoteMeta(`
        UPDATE vacancy
        SET
            title = $1,
            specialization_id = $2,
            work_format = $3,
            employment = $4,
            schedule = $5,
            working_hours = $6,
            salary_from = $7,
            salary_to = $8,
            taxes_included = $9,
            experience = $10,
            description = $11,
            tasks = $12,
            requirements = $13,
            optional_requirements = $14,
            city = $15,
            updated_at = NOW()
        WHERE id = $16 AND employer_id = $17
        RETURNING id, employer_id, title, specialization_id, work_format,
         employment, schedule, working_hours, salary_from, salary_to, taxes_included,
         experience, description, tasks, requirements, optional_requirements, city, created_at, updated_at
    `)

	testCases := []struct {
		name           string
		inputVacancy   *entity.Vacancy
		expectedResult *entity.Vacancy
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy)
	}{
		{
			name: "Успешное обновление вакансии",
			inputVacancy: &entity.Vacancy{
				ID:                   1,
				EmployerID:           2,
				Title:                "Senior Go Developer",
				SpecializationID:     3,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
			},
			expectedResult: &entity.Vacancy{
				ID:                   1,
				EmployerID:           2,
				Title:                "Senior Go Developer",
				SpecializationID:     3,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				mock.ExpectQuery(query).
					WithArgs(
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
						vacancy.ID,
						vacancy.EmployerID,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								vacancy.ID,
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
								now,
								now,
							),
					)
			},
		},
		{
			name: "Ошибка - нарушение уникального ограничения",
			inputVacancy: &entity.Vacancy{
				ID:                   1,
				EmployerID:           2,
				Title:                "Senior Go Developer",
				SpecializationID:     3,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("конфликт уникальных данных вакансии"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				mock.ExpectQuery(query).
					WithArgs(
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
						vacancy.ID,
						vacancy.EmployerID,
					).
					WillReturnError(&pq.Error{Code: "23505"})
			},
		},
		{
			name: "Ошибка - нарушение внешнего ключа",
			inputVacancy: &entity.Vacancy{
				ID:                   1,
				EmployerID:           2,
				Title:                "Senior Go Developer",
				SpecializationID:     3,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("работодатель или специализация с указанным ID не существует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				mock.ExpectQuery(query).
					WithArgs(
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
						vacancy.ID,
						vacancy.EmployerID,
					).
					WillReturnError(&pq.Error{Code: "23503"})
			},
		},
		{
			name: "Ошибка - нарушение ограничения NOT NULL",
			inputVacancy: &entity.Vacancy{
				ID:                   1,
				EmployerID:           2,
				Title:                "Senior Go Developer",
				SpecializationID:     3,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				mock.ExpectQuery(query).
					WithArgs(
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
						vacancy.ID,
						vacancy.EmployerID,
					).
					WillReturnError(&pq.Error{Code: "23502"})
			},
		},
		{
			name: "Ошибка - нарушение формата данных",
			inputVacancy: &entity.Vacancy{
				ID:                   1,
				EmployerID:           2,
				Title:                "Senior Go Developer",
				SpecializationID:     3,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				mock.ExpectQuery(query).
					WithArgs(
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
						vacancy.ID,
						vacancy.EmployerID,
					).
					WillReturnError(&pq.Error{Code: "22P02"})
			},
		},
		{
			name: "Ошибка - нарушение ограничения проверки",
			inputVacancy: &entity.Vacancy{
				ID:                   1,
				EmployerID:           2,
				Title:                "Senior Go Developer",
				SpecializationID:     3,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные (например, salary_from > salary_to)"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				mock.ExpectQuery(query).
					WithArgs(
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
						vacancy.ID,
						vacancy.EmployerID,
					).
					WillReturnError(&pq.Error{Code: "23514"})
			},
		},
		{
			name: "Ошибка - общая ошибка базы данных",
			inputVacancy: &entity.Vacancy{
				ID:                   1,
				EmployerID:           2,
				Title:                "Senior Go Developer",
				SpecializationID:     3,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось обновить вакансию с id=%d", 1),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				mock.ExpectQuery(query).
					WithArgs(
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
						vacancy.ID,
						vacancy.EmployerID,
					).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name: "Ошибка - вакансия не найдена или не принадлежит работодателю",
			inputVacancy: &entity.Vacancy{
				ID:                   1,
				EmployerID:           2,
				Title:                "Senior Go Developer",
				SpecializationID:     3,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         40,
				SalaryFrom:           200000,
				SalaryTo:             300000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Разработка высоконагруженных систем",
				Tasks:                "Писать код, проводить код-ревью",
				Requirements:         "Знание Go, PostgreSQL",
				OptionalRequirements: "Опыт с Kubernetes",
				City:                 "Москва",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось обновить вакансию с id=%d", 1),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancy *entity.Vacancy) {
				mock.ExpectQuery(query).
					WithArgs(
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
						vacancy.ID,
						vacancy.EmployerID,
					).
					WillReturnError(sql.ErrNoRows)
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

			result, err := repo.Update(ctx, tc.inputVacancy)

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
				require.Equal(t, tc.expectedResult.EmployerID, result.EmployerID)
				require.Equal(t, tc.expectedResult.Title, result.Title)
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

func TestVacancyRepository_GetAll(t *testing.T) {
	t.Parallel()

	now := time.Now()

	columns := []string{
		"id", "title", "is_active", "employer_id", "specialization_id",
		"work_format", "employment", "schedule", "working_hours",
		"salary_from", "salary_to", "taxes_included", "experience",
		"description", "tasks", "requirements", "optional_requirements",
		"city", "created_at", "updated_at",
	}

	query := regexp.QuoteMeta(`
        SELECT
            id,
            title,
            is_active,
            employer_id,
            specialization_id,
            work_format,
            employment,
            schedule,
            working_hours,
            salary_from,
            salary_to,
            taxes_included,
            experience,
            description,
            tasks,
            requirements,
            optional_requirements,
			city,
			created_at,
			updated_at
        FROM vacancy
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`)

	testCases := []struct {
		name           string
		limit          int
		offset         int
		expectedResult []*entity.Vacancy
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, limit, offset int)
	}{
		{
			name:   "Успешное получение нескольких вакансий",
			limit:  2,
			offset: 0,
			expectedResult: []*entity.Vacancy{
				{
					ID:                   1,
					Title:                "Senior Go Developer",
					IsActive:             true,
					EmployerID:           2,
					SpecializationID:     3,
					WorkFormat:           "remote",
					Employment:           "full_time",
					Schedule:             "5/2",
					WorkingHours:         40,
					SalaryFrom:           200000,
					SalaryTo:             300000,
					TaxesIncluded:        true,
					Experience:           "3_6_years",
					Description:          "Разработка высоконагруженных систем",
					Tasks:                "Писать код, проводить код-ревью",
					Requirements:         "Знание Go, PostgreSQL",
					OptionalRequirements: "Опыт с Kubernetes",
					City:                 "Москва",
					CreatedAt:            now,
					UpdatedAt:            now,
				},
				{
					ID:                   2,
					Title:                "Junior Python Developer",
					IsActive:             false,
					EmployerID:           3,
					SpecializationID:     4,
					WorkFormat:           "office",
					Employment:           "full_time",
					Schedule:             "5/2",
					WorkingHours:         40,
					SalaryFrom:           100000,
					SalaryTo:             150000,
					TaxesIncluded:        false,
					Experience:           "no_experience",
					Description:          "Разработка веб-приложений",
					Tasks:                "Писать код, тестировать",
					Requirements:         "Знание Python",
					OptionalRequirements: "Опыт с Django",
					City:                 "Санкт-Петербург",
					CreatedAt:            now,
					UpdatedAt:            now,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						"Senior Go Developer",
						true,
						2,
						3,
						"remote",
						"full_time",
						"5/2",
						40,
						200000,
						300000,
						true,
						"3_6_years",
						"Разработка высоконагруженных систем",
						"Писать код, проводить код-ревью",
						"Знание Go, PostgreSQL",
						"Опыт с Kubernetes",
						"Москва",
						now,
						now,
					).
					AddRow(
						2,
						"Junior Python Developer",
						false,
						3,
						4,
						"office",
						"full_time",
						"5/2",
						40,
						100000,
						150000,
						false,
						"no_experience",
						"Разработка веб-приложений",
						"Писать код, тестировать",
						"Знание Python",
						"Опыт с Django",
						"Санкт-Петербург",
						now,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:   "Успешное получение одной вакансии",
			limit:  1,
			offset: 0,
			expectedResult: []*entity.Vacancy{
				{
					ID:                   1,
					Title:                "Senior Go Developer",
					IsActive:             true,
					EmployerID:           2,
					SpecializationID:     3,
					WorkFormat:           "remote",
					Employment:           "full_time",
					Schedule:             "5/2",
					WorkingHours:         40,
					SalaryFrom:           200000,
					SalaryTo:             300000,
					TaxesIncluded:        true,
					Experience:           "3_6_years",
					Description:          "Разработка высоконагруженных систем",
					Tasks:                "Писать код, проводить код-ревью",
					Requirements:         "Знание Go, PostgreSQL",
					OptionalRequirements: "Опыт с Kubernetes",
					City:                 "Москва",
					CreatedAt:            now,
					UpdatedAt:            now,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						"Senior Go Developer",
						true,
						2,
						3,
						"remote",
						"full_time",
						"5/2",
						40,
						200000,
						300000,
						true,
						"3_6_years",
						"Разработка высоконагруженных систем",
						"Писать код, проводить код-ревью",
						"Знание Go, PostgreSQL",
						"Опыт с Kubernetes",
						"Москва",
						now,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение пустого списка",
			limit:          10,
			offset:         0,
			expectedResult: []*entity.Vacancy{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnRows(sqlmock.NewRows(columns))
			},
		},
		{
			name:           "Ошибка при выполнении запроса",
			limit:          10,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить список вакансий: %w", errors.New("query error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnError(errors.New("query error"))
			},
		},
		{
			name:           "Ошибка при сканировании строки",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обработке результатов запроса вакансий: %w", errors.New("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						"Senior Go Developer",
						true,
						2,
						3,
						"remote",
						"full_time",
						"5/2",
						40,
						200000,
						300000,
						true,
						"3_6_years",
						"Разработка высоконагруженных систем",
						"Писать код, проводить код-ревью",
						"Знание Go, PostgreSQL",
						"Опыт с Kubernetes",
						"Москва",
						now,
						now,
					).
					RowError(0, errors.New("scan error"))
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка при обработке результатов",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обработке результатов запроса вакансий: %w", errors.New("iteration error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						"Senior Go Developer",
						true,
						2,
						3,
						"remote",
						"full_time",
						"5/2",
						40,
						200000,
						300000,
						true,
						"3_6_years",
						"Разработка высоконагруженных систем",
						"Писать код, проводить код-ревью",
						"Знание Go, PostgreSQL",
						"Опыт с Kubernetes",
						"Москва",
						now,
						now,
					).
					AddRow(
						2,
						"Junior Python Developer",
						false,
						3,
						4,
						"office",
						"full_time",
						"5/2",
						40,
						100000,
						150000,
						false,
						"no_experience",
						"Разработка веб-приложений",
						"Писать код, тестировать",
						"Знание Python",
						"Опыт с Django",
						"Санкт-Петербург",
						now,
						now,
					).
					RowError(1, errors.New("iteration error"))
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:   "Пагинация с ненулевым offset",
			limit:  1,
			offset: 1,
			expectedResult: []*entity.Vacancy{
				{
					ID:                   2,
					Title:                "Junior Python Developer",
					IsActive:             false,
					EmployerID:           3,
					SpecializationID:     4,
					WorkFormat:           "office",
					Employment:           "full_time",
					Schedule:             "5/2",
					WorkingHours:         40,
					SalaryFrom:           100000,
					SalaryTo:             150000,
					TaxesIncluded:        false,
					Experience:           "no_experience",
					Description:          "Разработка веб-приложений",
					Tasks:                "Писать код, тестировать",
					Requirements:         "Знание Python",
					OptionalRequirements: "Опыт с Django",
					City:                 "Санкт-Петербург",
					CreatedAt:            now,
					UpdatedAt:            now,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						2,
						"Junior Python Developer",
						false,
						3,
						4,
						"office",
						"full_time",
						"5/2",
						40,
						100000,
						150000,
						false,
						"no_experience",
						"Разработка веб-приложений",
						"Писать код, тестировать",
						"Знание Python",
						"Опыт с Django",
						"Санкт-Петербург",
						now,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
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

			tc.setupMock(mock, tc.limit, tc.offset)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetAll(ctx, tc.limit, tc.offset)

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
					require.NotNil(t, result[i])
					require.Equal(t, expected.ID, result[i].ID)
					require.Equal(t, expected.Title, result[i].Title)
					require.Equal(t, expected.IsActive, result[i].IsActive)
					require.Equal(t, expected.EmployerID, result[i].EmployerID)
					require.Equal(t, expected.SpecializationID, result[i].SpecializationID)
					require.Equal(t, expected.WorkFormat, result[i].WorkFormat)
					require.Equal(t, expected.Employment, result[i].Employment)
					require.Equal(t, expected.Schedule, result[i].Schedule)
					require.Equal(t, expected.WorkingHours, result[i].WorkingHours)
					require.Equal(t, expected.SalaryFrom, result[i].SalaryFrom)
					require.Equal(t, expected.SalaryTo, result[i].SalaryTo)
					require.Equal(t, expected.TaxesIncluded, result[i].TaxesIncluded)
					require.Equal(t, expected.Experience, result[i].Experience)
					require.Equal(t, expected.Description, result[i].Description)
					require.Equal(t, expected.Tasks, result[i].Tasks)
					require.Equal(t, expected.Requirements, result[i].Requirements)
					require.Equal(t, expected.OptionalRequirements, result[i].OptionalRequirements)
					require.Equal(t, expected.City, result[i].City)
					require.Equal(t, expected.CreatedAt, result[i].CreatedAt)
					require.Equal(t, expected.UpdatedAt, result[i].UpdatedAt)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_Delete(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
        DELETE FROM vacancy
        WHERE id = $1
    `)

	testCases := []struct {
		name        string
		vacancyID   int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, vacancyID int)
	}{
		{
			name:        "Успешное удаление вакансии",
			vacancyID:   1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:      "Вакансия не найдена",
			vacancyID: 999,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("вакансия с id=%d не найдена", 999),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:      "Ошибка при выполнении запроса",
			vacancyID: 1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось удалить вакансию с id=%d", 1),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnError(errors.New("query error"))
			},
		},
		{
			name:      "Ошибка при получении количества затронутых строк",
			vacancyID: 1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении количества затронутых строк: %w", errors.New("rows affected error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))
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

			tc.setupMock(mock, tc.vacancyID)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			err = repo.Delete(ctx, tc.vacancyID)

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
	t.Parallel()

	query := regexp.QuoteMeta(`
		SELECT s.id, s.name
		FROM skill s
		JOIN vacancy_skill vs ON s.id = vs.skill_id
		WHERE vs.vacancy_id = $1
	`)

	columns := []string{"id", "name"}

	testCases := []struct {
		name           string
		vacancyID      int
		expectedResult []entity.Skill
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, vacancyID int)
	}{
		{
			name:      "Успешное получение навыков",
			vacancyID: 1,
			expectedResult: []entity.Skill{
				{ID: 1, Name: "Go"},
				{ID: 2, Name: "SQL"},
				{ID: 3, Name: "Docker"},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Go").
					AddRow(2, "SQL").
					AddRow(3, "Docker")
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение - нет навыков",
			vacancyID:      2,
			expectedResult: []entity.Skill{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - внутренняя ошибка при выполнении запроса",
			vacancyID:      1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении навыков резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка - ошибка при сканировании",
			vacancyID:      1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании навыка: %w", errors.New("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				rows := sqlmock.NewRows(columns).
					AddRow("invalid", "Go") // Некорректное значение для id
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при итерации",
			vacancyID:      1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по навыкам: %w", errors.New("iteration error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Go").
					AddRow(2, "SQL")
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnRows(rows)
				rows.CloseError(errors.New("iteration error"))
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

			tc.setupMock(mock, tc.vacancyID)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetSkillsByVacancyID(ctx, tc.vacancyID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))
				for i, expectedSkill := range tc.expectedResult {
					require.Equal(t, expectedSkill.ID, result[i].ID)
					require.Equal(t, expectedSkill.Name, result[i].Name)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_GetCityByVacancyID(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		SELECT c.id, c.name
		FROM city c
		JOIN vacancy_city vc ON c.id = vc.city_id
		WHERE vc.vacancy_id = $1
	`)

	columns := []string{"id", "name"}

	testCases := []struct {
		name           string
		vacancyID      int
		expectedResult []entity.City
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, vacancyID int)
	}{
		{
			name:      "Успешное получение городов",
			vacancyID: 1,
			expectedResult: []entity.City{
				{ID: 1, Name: "Москва"},
				{ID: 2, Name: "Санкт-Петербург"},
				{ID: 3, Name: "Екатеринбург"},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Москва").
					AddRow(2, "Санкт-Петербург").
					AddRow(3, "Екатеринбург")
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение - нет городов",
			vacancyID:      2,
			expectedResult: []entity.City{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - внутренняя ошибка при выполнении запроса",
			vacancyID:      1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении городов вакансии: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка - ошибка при сканировании",
			vacancyID:      1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании города: %w", errors.New("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				rows := sqlmock.NewRows(columns).
					AddRow("invalid", "Москва") // Некорректное значение для id
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при итерации",
			vacancyID:      1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по городам: %w", errors.New("iteration error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Москва").
					AddRow(2, "Санкт-Петербург")
				mock.ExpectQuery(query).
					WithArgs(vacancyID).
					WillReturnRows(rows)
				rows.CloseError(errors.New("iteration error"))
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

			tc.setupMock(mock, tc.vacancyID)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetCityByVacancyID(ctx, tc.vacancyID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))
				for i, expectedCity := range tc.expectedResult {
					require.Equal(t, expectedCity.ID, result[i].ID)
					require.Equal(t, expectedCity.Name, result[i].Name)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_DeleteSkills(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		DELETE FROM vacancy_skill
		WHERE vacancy_id = $1
	`)

	testCases := []struct {
		name        string
		vacancyID   int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, vacancyID int)
	}{
		{
			name:      "Успешное удаление навыков",
			vacancyID: 1,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:      "Успешное удаление - навыки отсутствуют",
			vacancyID: 2,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:      "Ошибка - внутренняя ошибка при выполнении запроса",
			vacancyID: 1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении навыков вакансии: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID).
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
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.vacancyID)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			err = repo.DeleteSkills(ctx, tc.vacancyID)

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
func TestVacancyRepository_DeleteCity(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		DELETE FROM vacancy_city
		WHERE vacancy_id = $1
	`)

	testCases := []struct {
		name        string
		vacancyID   int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, vacancyID int)
	}{
		{
			name:      "Успешное удаление городов",
			vacancyID: 1,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:      "Успешное удаление - города отсутствуют",
			vacancyID: 2,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:      "Ошибка - внутренняя ошибка при выполнении запроса",
			vacancyID: 1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении городов вакансии: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID).
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
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.vacancyID)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			err = repo.DeleteCity(ctx, tc.vacancyID)

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

func TestVacancyRepository_FindSkillIDsByNames(t *testing.T) {
	t.Parallel()

	selectQuery := regexp.QuoteMeta(`
		SELECT id
		FROM skill
		WHERE name = $1
	`)

	insertQuery := regexp.QuoteMeta(`
		INSERT INTO skill (name)
		VALUES ($1)
		RETURNING id
	`)

	testCases := []struct {
		name           string
		skillNames     []string
		expectedResult []int
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, skillNames []string)
	}{
		{
			name:           "Успешное получение ID существующих навыков",
			skillNames:     []string{"Go", "SQL"},
			expectedResult: []int{1, 2},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				for i, name := range skillNames {
					mock.ExpectQuery(selectQuery).
						WithArgs(name).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 1))
				}
			},
		},
		{
			name:           "Успешное создание новых навыков",
			skillNames:     []string{"Docker", "Kubernetes"},
			expectedResult: []int{3, 4},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				for i, name := range skillNames {
					// Симулируем, что навык не существует
					mock.ExpectQuery(selectQuery).
						WithArgs(name).
						WillReturnRows(sqlmock.NewRows([]string{"id"}))
					// Симулируем создание нового навыка
					mock.ExpectQuery(insertQuery).
						WithArgs(name).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 3))
				}
			},
		},
		{
			name:           "Успешное получение - пустой список",
			skillNames:     []string{},
			expectedResult: []int{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				// Нет запросов, так как список пустой
			},
		},
		{
			name:           "Ошибка - ошибка при поиске навыка",
			skillNames:     []string{"Go"},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования навыка: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(skillNames[0]).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка - ошибка при создании нового навыка",
			skillNames:     []string{"Docker"},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании навыка: %w", errors.New("insert error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				// Симулируем, что навык не существует
				mock.ExpectQuery(selectQuery).
					WithArgs(skillNames[0]).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
				// Симулируем ошибку при создании
				mock.ExpectQuery(insertQuery).
					WithArgs(skillNames[0]).
					WillReturnError(errors.New("insert error"))
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

			tc.setupMock(mock, tc.skillNames)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.FindSkillIDsByNames(ctx, tc.skillNames)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_FindCityIDsByNames(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
        SELECT id
        FROM city
        WHERE name IN
	`)

	testCases := []struct {
		name           string
		cityNames      []string
		expectedResult []int
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, cityNames []string)
	}{
		{
			name:           "Успешное получение ID городов",
			cityNames:      []string{"Москва", "Санкт-Петербург"},
			expectedResult: []int{1, 2},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, cityNames []string) {
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1).
					AddRow(2)
				mock.ExpectQuery(query).
					WithArgs(
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение - пустой список",
			cityNames:      []string{},
			expectedResult: []int{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, cityNames []string) {
				// Нет запросов, так как список пустой
			},
		},
		{
			name:           "Успешное получение - некоторые города не найдены",
			cityNames:      []string{"Москва", "Неизвестный"},
			expectedResult: []int{1},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, cityNames []string) {
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1)
				mock.ExpectQuery(query).
					WithArgs(
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при выполнении запроса",
			cityNames:      []string{"Москва"},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске ID городов по названиям: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, cityNames []string) {
				mock.ExpectQuery(query).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка - ошибка при сканировании",
			cityNames:      []string{"Москва"},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании ID города: %w", errors.New("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock, cityNames []string) {
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow("invalid") // Некорректное значение для id
				mock.ExpectQuery(query).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при итерации",
			cityNames:      []string{"Москва", "Санкт-Петербург"},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по ID городов: %w", errors.New("iteration error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, cityNames []string) {
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1).
					AddRow(2)
				mock.ExpectQuery(query).
					WithArgs(
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnRows(rows)
				rows.CloseError(errors.New("iteration error"))
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

			tc.setupMock(mock, tc.cityNames)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.FindCityIDsByNames(ctx, tc.cityNames)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_ResponseExists(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM vacancy_response WHERE vacancy_id = $1 AND applicant_id = $2)`)

	testCases := []struct {
		name           string
		vacancyID      int
		applicantID    int
		expectedResult bool
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, vacancyID, applicantID int)
	}{
		{
			name:           "Успешное получение - отклик существует",
			vacancyID:      1,
			applicantID:    1,
			expectedResult: true,
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectQuery(query).
					WithArgs(vacancyID, applicantID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
		},
		{
			name:           "Успешное получение - отклик не существует",
			vacancyID:      1,
			applicantID:    2,
			expectedResult: false,
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectQuery(query).
					WithArgs(vacancyID, applicantID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
		},
		{
			name:           "Ошибка - внутренняя ошибка базы данных",
			vacancyID:      1,
			applicantID:    1,
			expectedResult: false,
			expectedErr:    errors.New("database error"),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectQuery(query).
					WithArgs(vacancyID, applicantID).
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
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.vacancyID, tc.applicantID)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.ResponseExists(ctx, tc.vacancyID, tc.applicantID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.False(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_FindSpecializationIDByName(t *testing.T) {
	t.Parallel()

	selectQuery := regexp.QuoteMeta(`
		SELECT id
		FROM specialization
		WHERE name = $1
	`)

	insertQuery := regexp.QuoteMeta(`
		INSERT INTO specialization (name)
		VALUES ($1)
		RETURNING id
	`)

	testCases := []struct {
		name               string
		specializationName string
		expectedResult     int
		expectedErr        error
		setupMock          func(mock sqlmock.Sqlmock, specializationName string)
	}{
		{
			name:               "Успешное получение ID существующей специализации",
			specializationName: "Backend Development",
			expectedResult:     1,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name:               "Успешное создание новой специализации",
			specializationName: "DevOps",
			expectedResult:     2,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				// Симулируем, что специализация не существует
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
				// Симулируем создание новой специализации
				mock.ExpectQuery(insertQuery).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
		},
		{
			name:               "Ошибка - ошибка при поиске специализации",
			specializationName: "Frontend Development",
			expectedResult:     0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования специализации: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:               "Ошибка - ошибка при создании новой специализации",
			specializationName: "Data Science",
			expectedResult:     0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании специализации: %w", errors.New("insert error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				// Симулируем, что специализация не существует
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
				// Симулируем ошибку при создании
				mock.ExpectQuery(insertQuery).
					WithArgs(specializationName).
					WillReturnError(errors.New("insert error"))
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

			tc.setupMock(mock, tc.specializationName)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.FindSpecializationIDByName(ctx, tc.specializationName)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Equal(t, tc.expectedResult, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_CreateSpecializationIfNotExists(t *testing.T) {
	t.Parallel()

	selectQuery := regexp.QuoteMeta(`
        SELECT id
        FROM specialization
        WHERE name = $1
    `)

	insertQuery := regexp.QuoteMeta(`
        INSERT INTO specialization (name)
        VALUES ($1)
        RETURNING id
    `)

	testCases := []struct {
		name               string
		specializationName string
		expectedResult     int
		expectedErr        error
		setupMock          func(mock sqlmock.Sqlmock, specializationName string)
	}{
		{
			name:               "Успешное получение ID существующей специализации",
			specializationName: "Backend Development",
			expectedResult:     1,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name:               "Успешное создание новой специализации",
			specializationName: "DevOps",
			expectedResult:     2,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
		},
		{
			name:               "Успешное создание при конфликте уникальности",
			specializationName: "Data Science",
			expectedResult:     3,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(specializationName).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
			},
		},
		{
			name:               "Ошибка - ошибка при проверке существования",
			specializationName: "Frontend Development",
			expectedResult:     0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования специализации: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:               "Ошибка - ошибка при создании специализации",
			specializationName: "Machine Learning",
			expectedResult:     0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании специализации: %w", errors.New("insert error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(specializationName).
					WillReturnError(errors.New("insert error"))
			},
		},
		{
			name:               "Ошибка - ошибка при повторном получении после конфликта",
			specializationName: "Cloud Computing",
			expectedResult:     0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении ID специализации после конфликта: %w", errors.New("retry error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(insertQuery).
					WithArgs(specializationName).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectQuery(selectQuery).
					WithArgs(specializationName).
					WillReturnError(errors.New("retry error"))
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

			tc.setupMock(mock, tc.specializationName)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.CreateSpecializationIfNotExists(ctx, tc.specializationName)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Equal(t, tc.expectedResult, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_GetActiveVacanciesByEmployerID(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
        SELECT id, title, employer_id, specialization_id, work_format, employment,
               schedule, working_hours, salary_from, salary_to, taxes_included, experience,
               description, tasks, requirements, optional_requirements, city, created_at, updated_at
        FROM vacancy
        WHERE employer_id = $1 AND is_active = TRUE
        ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3;
    `)

	createdAt := time.Now().Add(-48 * time.Hour)
	updatedAt := time.Now()

	columns := []string{
		"id", "title", "employer_id", "specialization_id", "work_format", "employment",
		"schedule", "working_hours", "salary_from", "salary_to", "taxes_included", "experience",
		"description", "tasks", "requirements", "optional_requirements", "city", "created_at", "updated_at",
	}

	testCases := []struct {
		name           string
		employerID     int
		limit          int
		offset         int
		expectedResult []*entity.Vacancy
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, employerID, limit, offset int)
	}{
		{
			name:       "Успешное получение активных вакансий",
			employerID: 1,
			limit:      2,
			offset:     0,
			expectedResult: []*entity.Vacancy{
				{
					ID:                   1,
					Title:                "Senior Go Developer",
					EmployerID:           1,
					SpecializationID:     2,
					WorkFormat:           "remote",
					Employment:           "full_time",
					Schedule:             "5/2",
					WorkingHours:         40,
					SalaryFrom:           150000,
					SalaryTo:             200000,
					TaxesIncluded:        true,
					Experience:           "3_6_years",
					Description:          "Develop backend services",
					Tasks:                "Write clean code",
					Requirements:         "Go, SQL",
					OptionalRequirements: "Docker",
					City:                 "Москва",
					CreatedAt:            createdAt,
					UpdatedAt:            updatedAt,
				},
				{
					ID:                   2,
					Title:                "DevOps Engineer",
					EmployerID:           1,
					SpecializationID:     3,
					WorkFormat:           "hybrid",
					Employment:           "full_time",
					Schedule:             "5/2",
					WorkingHours:         40,
					SalaryFrom:           180000,
					SalaryTo:             250000,
					TaxesIncluded:        false,
					Experience:           "3_6_years",
					Description:          "Manage CI/CD pipelines",
					Tasks:                "Automate deployments",
					Requirements:         "Kubernetes, AWS",
					OptionalRequirements: "Terraform",
					City:                 "Санкт-Петербург",
					CreatedAt:            createdAt,
					UpdatedAt:            updatedAt,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, employerID, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1, "Senior Go Developer", 1, 2, "remote", "full_time",
						"5/2", 40, 150000, 200000, true, "3_6_years",
						"Develop backend services", "Write clean code", "Go, SQL", "Docker",
						"Москва", createdAt, updatedAt,
					).
					AddRow(
						2, "DevOps Engineer", 1, 3, "hybrid", "full_time",
						"5/2", 40, 180000, 250000, false, "3_6_years",
						"Manage CI/CD pipelines", "Automate deployments", "Kubernetes, AWS", "Terraform",
						"Санкт-Петербург", createdAt, updatedAt,
					)
				mock.ExpectQuery(query).
					WithArgs(employerID, limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение - пустой список",
			employerID:     2,
			limit:          2,
			offset:         0,
			expectedResult: []*entity.Vacancy{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, employerID, limit, offset int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(employerID, limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при выполнении запроса",
			employerID:     1,
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении активных вакансий работодателя: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, employerID, limit, offset int) {
				mock.ExpectQuery(query).
					WithArgs(employerID, limit, offset).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка - ошибка при сканировании",
			employerID:     1,
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка обработки данных вакансии: %w", errors.New("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock, employerID, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						"invalid", "Senior Go Developer", 1, 2, "remote", "full_time",
						"5/2", 40, 150000, 200000, true, "3_6_years",
						"Develop backend services", "Write clean code", "Go, SQL", "Docker",
						"Москва", createdAt, updatedAt,
					)
				mock.ExpectQuery(query).
					WithArgs(employerID, limit, offset).
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

			tc.setupMock(mock, tc.employerID, tc.limit, tc.offset)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetActiveVacanciesByEmployerID(ctx, tc.employerID, tc.limit, tc.offset)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))
				for i, expectedVacancy := range tc.expectedResult {
					require.Equal(t, expectedVacancy.ID, result[i].ID)
					require.Equal(t, expectedVacancy.Title, result[i].Title)
					require.Equal(t, expectedVacancy.EmployerID, result[i].EmployerID)
					require.Equal(t, expectedVacancy.SpecializationID, result[i].SpecializationID)
					require.Equal(t, expectedVacancy.WorkFormat, result[i].WorkFormat)
					require.Equal(t, expectedVacancy.Employment, result[i].Employment)
					require.Equal(t, expectedVacancy.Schedule, result[i].Schedule)
					require.Equal(t, expectedVacancy.WorkingHours, result[i].WorkingHours)
					require.Equal(t, expectedVacancy.SalaryFrom, result[i].SalaryFrom)
					require.Equal(t, expectedVacancy.SalaryTo, result[i].SalaryTo)
					require.Equal(t, expectedVacancy.TaxesIncluded, result[i].TaxesIncluded)
					require.Equal(t, expectedVacancy.Experience, result[i].Experience)
					require.Equal(t, expectedVacancy.Description, result[i].Description)
					require.Equal(t, expectedVacancy.Tasks, result[i].Tasks)
					require.Equal(t, expectedVacancy.Requirements, result[i].Requirements)
					require.Equal(t, expectedVacancy.OptionalRequirements, result[i].OptionalRequirements)
					require.Equal(t, expectedVacancy.City, result[i].City)
					require.Equal(t, expectedVacancy.CreatedAt, result[i].CreatedAt)
					require.Equal(t, expectedVacancy.UpdatedAt, result[i].UpdatedAt)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_SearchVacancies(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
        SELECT v.id, v.title, v.is_active, v.employer_id, v.specialization_id, v.work_format,
               v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to,
               v.taxes_included, v.experience, v.description, v.tasks, v.requirements,
               v.optional_requirements, v.city, v.created_at, v.updated_at
        FROM vacancy v
        JOIN employer e ON v.employer_id = e.id
        JOIN specialization s ON v.specialization_id = s.id
        WHERE v.title ILIKE $1
           OR s.name ILIKE $1
           OR e.company_name ILIKE $1
        ORDER BY v.updated_at DESC
        LIMIT $2 OFFSET $3
    `)

	createdAt := time.Now().Add(-48 * time.Hour)
	updatedAt := time.Now()

	columns := []string{
		"id", "title", "is_active", "employer_id", "specialization_id", "work_format", "employment",
		"schedule", "working_hours", "salary_from", "salary_to", "taxes_included", "experience",
		"description", "tasks", "requirements", "optional_requirements", "city", "created_at", "updated_at",
	}

	testCases := []struct {
		name           string
		searchQuery    string
		limit          int
		offset         int
		expectedResult []*entity.Vacancy
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, searchQuery string, limit, offset int)
	}{
		{
			name:        "Успешное получение вакансий",
			searchQuery: "developer",
			limit:       2,
			offset:      0,
			expectedResult: []*entity.Vacancy{
				{
					ID:                   1,
					Title:                "Senior Go Developer",
					IsActive:             true,
					EmployerID:           1,
					SpecializationID:     2,
					WorkFormat:           "remote",
					Employment:           "full_time",
					Schedule:             "5/2",
					WorkingHours:         40,
					SalaryFrom:           150000,
					SalaryTo:             200000,
					TaxesIncluded:        true,
					Experience:           "3_6_years",
					Description:          "Develop backend services",
					Tasks:                "Write clean code",
					Requirements:         "Go, SQL",
					OptionalRequirements: "Docker",
					City:                 "Москва",
					CreatedAt:            createdAt,
					UpdatedAt:            updatedAt,
				},
				{
					ID:                   2,
					Title:                "Frontend Developer",
					IsActive:             true,
					EmployerID:           1,
					SpecializationID:     3,
					WorkFormat:           "hybrid",
					Employment:           "full_time",
					Schedule:             "5/2",
					WorkingHours:         40,
					SalaryFrom:           120000,
					SalaryTo:             180000,
					TaxesIncluded:        false,
					Experience:           "1_3_years",
					Description:          "Develop UI components",
					Tasks:                "Implement responsive designs",
					Requirements:         "React, JavaScript",
					OptionalRequirements: "TypeScript",
					City:                 "Санкт-Петербург",
					CreatedAt:            createdAt,
					UpdatedAt:            updatedAt,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, searchQuery string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1, "Senior Go Developer", true, 1, 2, "remote", "full_time",
						"5/2", 40, 150000, 200000, true, "3_6_years",
						"Develop backend services", "Write clean code", "Go, SQL", "Docker",
						"Москва", createdAt, updatedAt,
					).
					AddRow(
						2, "Frontend Developer", true, 1, 3, "hybrid", "full_time",
						"5/2", 40, 120000, 180000, false, "1_3_years",
						"Develop UI components", "Implement responsive designs", "React, JavaScript", "TypeScript",
						"Санкт-Петербург", createdAt, updatedAt,
					)
				mock.ExpectQuery(query).
					WithArgs("%"+searchQuery+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение - пустой список",
			searchQuery:    "nonexistent",
			limit:          2,
			offset:         0,
			expectedResult: []*entity.Vacancy{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, searchQuery string, limit, offset int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs("%"+searchQuery+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при выполнении запроса",
			searchQuery:    "developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске вакансий: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, searchQuery string, limit, offset int) {
				mock.ExpectQuery(query).
					WithArgs("%"+searchQuery+"%", limit, offset).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка - ошибка при сканировании",
			searchQuery:    "developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка обработки данных вакансии: %w", errors.New("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock, searchQuery string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						"invalid", "Senior Go Developer", true, 1, 2, "remote", "full_time",
						"5/2", 40, 150000, 200000, true, "3_6_years",
						"Develop backend services", "Write clean code", "Go, SQL", "Docker",
						"Москва", createdAt, updatedAt,
					)
				mock.ExpectQuery(query).
					WithArgs("%"+searchQuery+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при итерации",
			searchQuery:    "developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обработке результатов запроса вакансий: %w", errors.New("iteration error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, searchQuery string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1, "Senior Go Developer", true, 1, 2, "remote", "full_time",
						"5/2", 40, 150000, 200000, true, "3_6_years",
						"Develop backend services", "Write clean code", "Go, SQL", "Docker",
						"Москва", createdAt, updatedAt,
					)
				mock.ExpectQuery(query).
					WithArgs("%"+searchQuery+"%", limit, offset).
					WillReturnRows(rows)
				rows.CloseError(errors.New("iteration error"))
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

			tc.setupMock(mock, tc.searchQuery, tc.limit, tc.offset)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.SearchVacancies(ctx, tc.searchQuery, tc.limit, tc.offset)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))
				for i, expectedVacancy := range tc.expectedResult {
					require.Equal(t, expectedVacancy.ID, result[i].ID)
					require.Equal(t, expectedVacancy.Title, result[i].Title)
					require.Equal(t, expectedVacancy.IsActive, result[i].IsActive)
					require.Equal(t, expectedVacancy.EmployerID, result[i].EmployerID)
					require.Equal(t, expectedVacancy.SpecializationID, result[i].SpecializationID)
					require.Equal(t, expectedVacancy.WorkFormat, result[i].WorkFormat)
					require.Equal(t, expectedVacancy.Employment, result[i].Employment)
					require.Equal(t, expectedVacancy.Schedule, result[i].Schedule)
					require.Equal(t, expectedVacancy.WorkingHours, result[i].WorkingHours)
					require.Equal(t, expectedVacancy.SalaryFrom, result[i].SalaryFrom)
					require.Equal(t, expectedVacancy.SalaryTo, result[i].SalaryTo)
					require.Equal(t, expectedVacancy.TaxesIncluded, result[i].TaxesIncluded)
					require.Equal(t, expectedVacancy.Experience, result[i].Experience)
					require.Equal(t, expectedVacancy.Description, result[i].Description)
					require.Equal(t, expectedVacancy.Tasks, result[i].Tasks)
					require.Equal(t, expectedVacancy.Requirements, result[i].Requirements)
					require.Equal(t, expectedVacancy.OptionalRequirements, result[i].OptionalRequirements)
					require.Equal(t, expectedVacancy.City, result[i].City)
					require.Equal(t, expectedVacancy.CreatedAt, result[i].CreatedAt)
					require.Equal(t, expectedVacancy.UpdatedAt, result[i].UpdatedAt)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// func TestVacancyRepository_GetVacanciesByApplicantID(t *testing.T) {
// 	t.Parallel()

// 	query := regexp.QuoteMeta(`
// 		SELECT v.id, v.title, v.employer_id, v.specialization_id, v.work_format,
// 			v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to,
// 			v.taxes_included, v.experience, v.description, v.tasks, v.requirements,
// 			v.optional_requirements, v.city, v.created_at, v.updated_at
// 		FROM vacancy v
// 		JOIN (
// 			SELECT vacancy_id, MAX(applied_at) as last_applied_at
// 			FROM vacancy_response
// 			WHERE applicant_id = $1
// 			GROUP BY vacancy_id
// 		) vr ON v.id = vr.vacancy_id
// 		ORDER BY vr.last_applied_at DESC
// 		LIMIT $2 OFFSET $3
// 	`)

// 	createdAt := time.Now().Add(-48 * time.Hour)
// 	updatedAt := time.Now()

// 	columns := []string{
// 		"id", "title", "employer_id", "specialization_id", "work_format", "employment",
// 		"schedule", "working_hours", "salary_from", "salary_to", "taxes_included", "experience",
// 		"description", "tasks", "requirements", "optional_requirements", "city", "created_at", "updated_at",
// 	}

// 	testCases := []struct {
// 		name           string
// 		applicantID    int
// 		limit          int
// 		offset         int
// 		expectedResult []*entity.Vacancy
// 		expectedErr    error
// 		setupMock      func(mock sqlmock.Sqlmock, applicantID int, limit, offset int)
// 	}{
// 		{
// 			name:        "Успешное получение вакансий",
// 			applicantID: 1,
// 			limit:       2,
// 			offset:      0,
// 			expectedResult: []*entity.Vacancy{
// 				{
// 					ID:                   1,
// 					Title:                "Senior Go Developer",
// 					EmployerID:           1,
// 					SpecializationID:     2,
// 					WorkFormat:           "remote",
// 					Employment:           "full_time",
// 					Schedule:             "5/2",
// 					WorkingHours:         40,
// 					SalaryFrom:           150000,
// 					SalaryTo:             200000,
// 					TaxesIncluded:        true,
// 					Experience:           "3_6_years",
// 					Description:          "Develop backend services",
// 					Tasks:                "Write clean code",
// 					Requirements:         "Go, SQL",
// 					OptionalRequirements: "Docker",
// 					City:                 "Москва",
// 					CreatedAt:            createdAt,
// 					UpdatedAt:            updatedAt,
// 				},
// 				{
// 					ID:                   2,
// 					Title:                "DevOps Engineer",
// 					EmployerID:           1,
// 					SpecializationID:     3,
// 					WorkFormat:           "hybrid",
// 					Employment:           "full_time",
// 					Schedule:             "5/2",
// 					WorkingHours:         40,
// 					SalaryFrom:           180000,
// 					SalaryTo:             250000,
// 					TaxesIncluded:        false,
// 					Experience:           "3_6_years",
// 					Description:          "Manage CI/CD pipelines",
// 					Tasks:                "Automate deployments",
// 					Requirements:         "Kubernetes, AWS",
// 					OptionalRequirements: "Terraform",
// 					City:                 "Санкт-Петербург",
// 					CreatedAt:            createdAt,
// 					UpdatedAt:            updatedAt,
// 				},
// 			},
// 			expectedErr: nil,
// 			setupMock: func(mock sqlmock.Sqlmock, applicantID int, limit, offset int) {
// 				rows := sqlmock.NewRows(columns).
// 					AddRow(
// 						1, "Senior Go Developer", 1, 2, "remote", "full_time",
// 						"5/2", 40, 150000, 200000, true, "3_6_years",
// 						"Develop backend services", "Write clean code", "Go, SQL", "Docker",
// 						"Москва", createdAt, updatedAt,
// 					).
// 					AddRow(
// 						2, "DevOps Engineer", 1, 3, "hybrid", "full_time",
// 						"5/2", 40, 180000, 250000, false, "3_6_years",
// 						"Manage CI/CD pipelines", "Automate deployments", "Kubernetes, AWS", "Terraform",
// 						"Санкт-Петербург", createdAt, updatedAt,
// 					)
// 				mock.ExpectQuery(query).
// 					WithArgs(applicantID).
// 					WillReturnRows(rows)
// 			},
// 		},
// 		{
// 			name:           "Успешное получение - пустой список",
// 			applicantID:    2,
// 			limit:          2,
// 			offset:         0,
// 			expectedResult: []*entity.Vacancy{},
// 			expectedErr:    nil,
// 			setupMock: func(mock sqlmock.Sqlmock, applicantID int, limit, offset int) {
// 				rows := sqlmock.NewRows(columns)
// 				mock.ExpectQuery(query).
// 					WithArgs(applicantID).
// 					WillReturnRows(rows)
// 			},
// 		},
// 		{
// 			name:           "Ошибка - ошибка при выполнении запроса",
// 			applicantID:    1,
// 			limit:          2,
// 			offset:         0,
// 			expectedResult: nil,
// 			expectedErr: entity.NewError(
// 				entity.ErrInternal,
// 				fmt.Errorf("ошибка при получении списка вакансий: %w", errors.New("database error")),
// 			),
// 			setupMock: func(mock sqlmock.Sqlmock, applicantID int, limit, offset int) {
// 				mock.ExpectQuery(query).
// 					WithArgs(applicantID).
// 					WillReturnError(errors.New("database error"))
// 			},
// 		},
// 		{
// 			name:           "Ошибка - ошибка при сканировании",
// 			applicantID:    1,
// 			limit:          2,
// 			offset:         0,
// 			expectedResult: nil,
// 			expectedErr: entity.NewError(
// 				entity.ErrInternal,
// 				fmt.Errorf("ошибка обработки данных вакансии: %w", errors.New("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax")),
// 			),
// 			setupMock: func(mock sqlmock.Sqlmock, applicantID int, limit, offset int) {
// 				rows := sqlmock.NewRows(columns).
// 					AddRow(
// 						"invalid", "Senior Go Developer", 1, 2, "remote", "full_time",
// 						"5/2", 40, 150000, 200000, true, "3_6_years",
// 						"Develop backend services", "Write clean code", "Go, SQL", "Docker",
// 						"Москва", createdAt, updatedAt,
// 					)
// 				mock.ExpectQuery(query).
// 					WithArgs(applicantID).
// 					WillReturnRows(rows)
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			t.Parallel()

// 			db, mock, err := sqlmock.New()
// 			require.NoError(t, err)
// 			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
// 				mock.ExpectClose()
// 				err := db.Close()
// 				require.NoError(t, err)
// 			}(db, mock)

// 			tc.setupMock(mock, tc.applicantID, 2, 0)

// 			repo := &VacancyRepository{DB: db}
// 			ctx := context.Background()

// 			result, err := repo.GetVacanciesByApplicantID(ctx, tc.applicantID, tc.limit, tc.offset)

// 			if tc.expectedErr != nil {
// 				require.Error(t, err)
// 				var repoErr entity.Error
// 				require.ErrorAs(t, err, &repoErr)
// 				require.Equal(t, tc.expectedErr.Error(), err.Error())
// 				require.Nil(t, result)
// 			} else {
// 				require.NoError(t, err)
// 				require.Equal(t, len(tc.expectedResult), len(result))
// 				for i, expectedVacancy := range tc.expectedResult {
// 					require.Equal(t, expectedVacancy.ID, result[i].ID)
// 					require.Equal(t, expectedVacancy.Title, result[i].Title)
// 					require.Equal(t, expectedVacancy.EmployerID, result[i].EmployerID)
// 					require.Equal(t, expectedVacancy.SpecializationID, result[i].SpecializationID)
// 					require.Equal(t, expectedVacancy.WorkFormat, result[i].WorkFormat)
// 					require.Equal(t, expectedVacancy.Employment, result[i].Employment)
// 					require.Equal(t, expectedVacancy.Schedule, result[i].Schedule)
// 					require.Equal(t, expectedVacancy.WorkingHours, result[i].WorkingHours)
// 					require.Equal(t, expectedVacancy.SalaryFrom, result[i].SalaryFrom)
// 					require.Equal(t, expectedVacancy.SalaryTo, result[i].SalaryTo)
// 					require.Equal(t, expectedVacancy.TaxesIncluded, result[i].TaxesIncluded)
// 					require.Equal(t, expectedVacancy.Experience, result[i].Experience)
// 					require.Equal(t, expectedVacancy.Description, result[i].Description)
// 					require.Equal(t, expectedVacancy.Tasks, result[i].Tasks)
// 					require.Equal(t, expectedVacancy.Requirements, result[i].Requirements)
// 					require.Equal(t, expectedVacancy.OptionalRequirements, result[i].OptionalRequirements)
// 					require.Equal(t, expectedVacancy.City, result[i].City)
// 					require.Equal(t, expectedVacancy.CreatedAt, result[i].CreatedAt)
// 					require.Equal(t, expectedVacancy.UpdatedAt, result[i].UpdatedAt)
// 				}
// 			}
// 			require.NoError(t, mock.ExpectationsWereMet())
// 		})
// 	}
// }

func TestVacancyRepository_SearchVacanciesByEmployerID(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
        SELECT v.id, v.title, v.is_active, v.employer_id, v.specialization_id, v.work_format,
               v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to,
               v.taxes_included, v.experience, v.description, v.tasks, v.requirements,
               v.optional_requirements, v.city, v.created_at, v.updated_at
        FROM vacancy v
        JOIN specialization s ON v.specialization_id = s.id
        WHERE v.employer_id = $1
          AND (v.title ILIKE $2 OR s.name ILIKE $2)
        ORDER BY v.updated_at DESC
        LIMIT $3 OFFSET $4
    `)

	createdAt := time.Now().Add(-48 * time.Hour)
	updatedAt := time.Now()

	columns := []string{
		"id", "title", "is_active", "employer_id", "specialization_id", "work_format", "employment",
		"schedule", "working_hours", "salary_from", "salary_to", "taxes_included", "experience",
		"description", "tasks", "requirements", "optional_requirements", "city", "created_at", "updated_at",
	}

	testCases := []struct {
		name           string
		employerID     int
		searchQuery    string
		limit          int
		offset         int
		expectedResult []*entity.Vacancy
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, employerID int, searchQuery string, limit, offset int)
	}{
		{
			name:        "Успешное получение вакансий",
			employerID:  1,
			searchQuery: "developer",
			limit:       2,
			offset:      0,
			expectedResult: []*entity.Vacancy{
				{
					ID:                   1,
					Title:                "Senior Go Developer",
					IsActive:             true,
					EmployerID:           1,
					SpecializationID:     2,
					WorkFormat:           "remote",
					Employment:           "full_time",
					Schedule:             "5/2",
					WorkingHours:         40,
					SalaryFrom:           150000,
					SalaryTo:             200000,
					TaxesIncluded:        true,
					Experience:           "3_6_years",
					Description:          "Develop backend services",
					Tasks:                "Write clean code",
					Requirements:         "Go, SQL",
					OptionalRequirements: "Docker",
					City:                 "Москва",
					CreatedAt:            createdAt,
					UpdatedAt:            updatedAt,
				},
				{
					ID:                   2,
					Title:                "Frontend Developer",
					IsActive:             true,
					EmployerID:           1,
					SpecializationID:     3,
					WorkFormat:           "hybrid",
					Employment:           "full_time",
					Schedule:             "5/2",
					WorkingHours:         40,
					SalaryFrom:           120000,
					SalaryTo:             180000,
					TaxesIncluded:        false,
					Experience:           "1_3_years",
					Description:          "Develop UI components",
					Tasks:                "Implement responsive designs",
					Requirements:         "React, JavaScript",
					OptionalRequirements: "TypeScript",
					City:                 "Санкт-Петербург",
					CreatedAt:            createdAt,
					UpdatedAt:            updatedAt,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, employerID int, searchQuery string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1, "Senior Go Developer", true, 1, 2, "remote", "full_time",
						"5/2", 40, 150000, 200000, true, "3_6_years",
						"Develop backend services", "Write clean code", "Go, SQL", "Docker",
						"Москва", createdAt, updatedAt,
					).
					AddRow(
						2, "Frontend Developer", true, 1, 3, "hybrid", "full_time",
						"5/2", 40, 120000, 180000, false, "1_3_years",
						"Develop UI components", "Implement responsive designs", "React, JavaScript", "TypeScript",
						"Санкт-Петербург", createdAt, updatedAt,
					)
				mock.ExpectQuery(query).
					WithArgs(employerID, "%"+searchQuery+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение - пустой список",
			employerID:     1,
			searchQuery:    "nonexistent",
			limit:          2,
			offset:         0,
			expectedResult: []*entity.Vacancy{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, employerID int, searchQuery string, limit, offset int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(employerID, "%"+searchQuery+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при выполнении запроса",
			employerID:     1,
			searchQuery:    "developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске вакансий работодателя: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, employerID int, searchQuery string, limit, offset int) {
				mock.ExpectQuery(query).
					WithArgs(employerID, "%"+searchQuery+"%", limit, offset).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка - ошибка при сканировании",
			employerID:     1,
			searchQuery:    "developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка обработки данных вакансии: %w", errors.New("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock, employerID int, searchQuery string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						"invalid", "Senior Go Developer", true, 1, 2, "remote", "full_time",
						"5/2", 40, 150000, 200000, true, "3_6_years",
						"Develop backend services", "Write clean code", "Go, SQL", "Docker",
						"Москва", createdAt, updatedAt,
					)
				mock.ExpectQuery(query).
					WithArgs(employerID, "%"+searchQuery+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при итерации",
			employerID:     1,
			searchQuery:    "developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обработке результатов запроса вакансий: %w", errors.New("iteration error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, employerID int, searchQuery string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1, "Senior Go Developer", true, 1, 2, "remote", "full_time",
						"5/2", 40, 150000, 200000, true, "3_6_years",
						"Develop backend services", "Write clean code", "Go, SQL", "Docker",
						"Москва", createdAt, updatedAt,
					)
				mock.ExpectQuery(query).
					WithArgs(employerID, "%"+searchQuery+"%", limit, offset).
					WillReturnRows(rows)
				rows.CloseError(errors.New("iteration error"))
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

			tc.setupMock(mock, tc.employerID, tc.searchQuery, tc.limit, tc.offset)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.SearchVacanciesByEmployerID(ctx, tc.employerID, tc.searchQuery, tc.limit, tc.offset)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))
				for i, expectedVacancy := range tc.expectedResult {
					require.Equal(t, expectedVacancy.ID, result[i].ID)
					require.Equal(t, expectedVacancy.Title, result[i].Title)
					require.Equal(t, expectedVacancy.IsActive, result[i].IsActive)
					require.Equal(t, expectedVacancy.EmployerID, result[i].EmployerID)
					require.Equal(t, expectedVacancy.SpecializationID, result[i].SpecializationID)
					require.Equal(t, expectedVacancy.WorkFormat, result[i].WorkFormat)
					require.Equal(t, expectedVacancy.Employment, result[i].Employment)
					require.Equal(t, expectedVacancy.Schedule, result[i].Schedule)
					require.Equal(t, expectedVacancy.WorkingHours, result[i].WorkingHours)
					require.Equal(t, expectedVacancy.SalaryFrom, result[i].SalaryFrom)
					require.Equal(t, expectedVacancy.SalaryTo, result[i].SalaryTo)
					require.Equal(t, expectedVacancy.TaxesIncluded, result[i].TaxesIncluded)
					require.Equal(t, expectedVacancy.Experience, result[i].Experience)
					require.Equal(t, expectedVacancy.Description, result[i].Description)
					require.Equal(t, expectedVacancy.Tasks, result[i].Tasks)
					require.Equal(t, expectedVacancy.Requirements, result[i].Requirements)
					require.Equal(t, expectedVacancy.OptionalRequirements, result[i].OptionalRequirements)
					require.Equal(t, expectedVacancy.City, result[i].City)
					require.Equal(t, expectedVacancy.CreatedAt, result[i].CreatedAt)
					require.Equal(t, expectedVacancy.UpdatedAt, result[i].UpdatedAt)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_FindSpecializationIDsByNames(t *testing.T) {
	t.Parallel()

	columns := []string{"id"}

	testCases := []struct {
		name                string
		specializationNames []string
		expectedResult      []int
		expectedErr         error
		setupMock           func(mock sqlmock.Sqlmock, specializationNames []string)
	}{
		{
			name:                "Успешное получение ID специализаций",
			specializationNames: []string{"Go Developer", "DevOps"},
			expectedResult:      []int{1, 2},
			expectedErr:         nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationNames []string) {
				rows := sqlmock.NewRows(columns).
					AddRow(1).
					AddRow(2)
				query := regexp.QuoteMeta(`
		SELECT id
		FROM specialization
		WHERE name IN ($1, $2)
	`)
				mock.ExpectQuery(query).
					WithArgs(specializationNames[0], specializationNames[1]).
					WillReturnRows(rows)
			},
		},
		{
			name:                "Успешное получение - пустой входной список",
			specializationNames: []string{},
			expectedResult:      []int{},
			expectedErr:         nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationNames []string) {
				// No query is executed for empty input
			},
		},
		{
			name:                "Успешное получение - пустой результат",
			specializationNames: []string{"Nonexistent"},
			expectedResult:      []int(nil),
			expectedErr:         nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationNames []string) {
				rows := sqlmock.NewRows(columns)
				query := regexp.QuoteMeta(`
		SELECT id
		FROM specialization
		WHERE name IN ($1)
	`)
				mock.ExpectQuery(query).
					WithArgs(specializationNames[0]).
					WillReturnRows(rows)
			},
		},
		{
			name:                "Ошибка - ошибка при выполнении запроса",
			specializationNames: []string{"Go Developer"},
			expectedResult:      nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске ID специализаций по названиям: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationNames []string) {
				query := regexp.QuoteMeta(`
		SELECT id
		FROM specialization
		WHERE name IN ($1)
	`)
				mock.ExpectQuery(query).
					WithArgs(specializationNames[0]).
					WillReturnError(errors.New("database error"))
			},
		},

		{
			name:                "Ошибка - ошибка при итерации",
			specializationNames: []string{"Go Developer"},
			expectedResult:      nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по ID специализаций: %w", errors.New("iteration error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationNames []string) {
				rows := sqlmock.NewRows(columns).
					AddRow(1)
				query := regexp.QuoteMeta(`
		SELECT id
		FROM specialization
		WHERE name IN ($1)
	`)
				mock.ExpectQuery(query).
					WithArgs(specializationNames[0]).
					WillReturnRows(rows)
				rows.CloseError(errors.New("iteration error"))
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

			tc.setupMock(mock, tc.specializationNames)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.FindSpecializationIDsByNames(ctx, tc.specializationNames)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_SearchVacanciesBySpecializations(t *testing.T) {
	t.Parallel()

	now := time.Now()

	columns := []string{
		"id", "title", "is_active", "employer_id", "specialization_id",
		"work_format", "employment", "schedule", "working_hours",
		"salary_from", "salary_to", "taxes_included", "experience",
		"description", "tasks", "requirements", "optional_requirements",
		"city", "created_at", "updated_at",
	}

	testCases := []struct {
		name              string
		specializationIDs []int
		limit             int
		offset            int
		expectedResult    []*entity.Vacancy
		expectedErr       error
		setupMock         func(mock sqlmock.Sqlmock, specializationIDs []int, limit, offset int)
	}{
		{
			name:              "Пустой список specializationIDs",
			specializationIDs: []int{},
			limit:             10,
			offset:            0,
			expectedResult:    []*entity.Vacancy{},
			expectedErr:       nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationIDs []int, limit, offset int) {
				// No database queries expected for empty input
			},
		},
		{
			name:              "Успешное получение списка вакансий",
			specializationIDs: []int{1, 2},
			limit:             2,
			offset:            0,
			expectedResult: []*entity.Vacancy{
				{
					ID:                   1,
					Title:                "Software Engineer",
					IsActive:             true,
					EmployerID:           1,
					SpecializationID:     1,
					WorkFormat:           "hybrid",
					Employment:           "full_time",
					Schedule:             "5/2",
					WorkingHours:         40,
					SalaryFrom:           50000,
					SalaryTo:             70000,
					TaxesIncluded:        true,
					Experience:           "3_6_years",
					Description:          "Develop software solutions",
					Tasks:                "Write code, review PRs",
					Requirements:         "Go, SQL",
					OptionalRequirements: "Docker",
					City:                 "Moscow",
					CreatedAt:            now,
					UpdatedAt:            now,
				},
				{
					ID:                   2,
					Title:                "Frontend Developer",
					IsActive:             false,
					EmployerID:           2,
					SpecializationID:     2,
					WorkFormat:           "remote",
					Employment:           "part_time",
					Schedule:             "by_agreement",
					WorkingHours:         20,
					SalaryFrom:           30000,
					SalaryTo:             40000,
					TaxesIncluded:        false,
					Experience:           "1_3_years",
					Description:          "Build UI components",
					Tasks:                "Develop React components",
					Requirements:         "React, JavaScript",
					OptionalRequirements: "TypeScript",
					City:                 "Saint Petersburg",
					CreatedAt:            now,
					UpdatedAt:            now,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationIDs []int, limit, offset int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT v.id, v.title, v.is_active, v.employer_id, v.specialization_id, v.work_format,
						v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to,
						v.taxes_included, v.experience, v.description, v.tasks, v.requirements,
						v.optional_requirements, v.city, v.created_at, v.updated_at
					FROM vacancy v
					WHERE v.specialization_id IN (%s)
					ORDER BY v.updated_at DESC
					LIMIT $%d OFFSET $%d
				`, strings.Join([]string{"$1", "$2"}, ", "), len(specializationIDs)+1, len(specializationIDs)+2))
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						"Software Engineer",
						true,
						1,
						1,
						"hybrid",
						"full_time",
						"5/2",
						40,
						50000,
						70000,
						true,
						"3_6_years",
						"Develop software solutions",
						"Write code, review PRs",
						"Go, SQL",
						"Docker",
						"Moscow",
						now,
						now,
					).
					AddRow(
						2,
						"Frontend Developer",
						false,
						2,
						2,
						"remote",
						"part_time",
						"by_agreement",
						20,
						30000,
						40000,
						false,
						"1_3_years",
						"Build UI components",
						"Develop React components",
						"React, JavaScript",
						"TypeScript",
						"Saint Petersburg",
						now,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs(1, 2, limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:              "Пустой список вакансий",
			specializationIDs: []int{3, 4},
			limit:             10,
			offset:            0,
			expectedResult:    []*entity.Vacancy{},
			expectedErr:       nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationIDs []int, limit, offset int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT v.id, v.title, v.is_active, v.employer_id, v.specialization_id, v.work_format,
						v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to,
						v.taxes_included, v.experience, v.description, v.tasks, v.requirements,
						v.optional_requirements, v.city, v.created_at, v.updated_at
					FROM vacancy v
					WHERE v.specialization_id IN (%s)
					ORDER BY v.updated_at DESC
					LIMIT $%d OFFSET $%d
				`, strings.Join([]string{"$1", "$2"}, ", "), len(specializationIDs)+1, len(specializationIDs)+2))
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(3, 4, limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:              "Ошибка базы данных при выполнении запроса",
			specializationIDs: []int{1, 2},
			limit:             10,
			offset:            0,
			expectedResult:    nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске вакансий по специализациям: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationIDs []int, limit, offset int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT v.id, v.title, v.is_active, v.employer_id, v.specialization_id, v.work_format,
						v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to,
						v.taxes_included, v.experience, v.description, v.tasks, v.requirements,
						v.optional_requirements, v.city, v.created_at, v.updated_at
					FROM vacancy v
					WHERE v.specialization_id IN (%s)
					ORDER BY v.updated_at DESC
					LIMIT $%d OFFSET $%d
				`, strings.Join([]string{"$1", "$2"}, ", "), len(specializationIDs)+1, len(specializationIDs)+2))
				mock.ExpectQuery(query).
					WithArgs(1, 2, limit, offset).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:              "Ошибка сканирования строк",
			specializationIDs: []int{1, 2},
			limit:             2,
			offset:            0,
			expectedResult:    nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обработке результатов запроса вакансий: %w", errors.New("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationIDs []int, limit, offset int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT v.id, v.title, v.is_active, v.employer_id, v.specialization_id, v.work_format,
						v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to,
						v.taxes_included, v.experience, v.description, v.tasks, v.requirements,
						v.optional_requirements, v.city, v.created_at, v.updated_at
					FROM vacancy v
					WHERE v.specialization_id IN (%s)
					ORDER BY v.updated_at DESC
					LIMIT $%d OFFSET $%d
				`, strings.Join([]string{"$1", "$2"}, ", "), len(specializationIDs)+1, len(specializationIDs)+2))
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						"Software Engineer",
						true,
						1,
						1,
						"hybrid",
						"full_time",
						"5/2",
						40,
						50000,
						70000,
						true,
						"3_6_years",
						"Develop software solutions",
						"Write code, review PRs",
						"Go, SQL",
						"Docker",
						"Moscow",
						now,
						now,
					).
					RowError(0, errors.New("scan error"))
				mock.ExpectQuery(query).
					WithArgs(1, 2, limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:              "Ошибка итерации по строкам",
			specializationIDs: []int{1, 2},
			limit:             2,
			offset:            0,
			expectedResult:    nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обработке результатов запроса вакансий: %w", errors.New("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationIDs []int, limit, offset int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT v.id, v.title, v.is_active, v.employer_id, v.specialization_id, v.work_format,
						v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to,
						v.taxes_included, v.experience, v.description, v.tasks, v.requirements,
						v.optional_requirements, v.city, v.created_at, v.updated_at
					FROM vacancy v
					WHERE v.specialization_id IN (%s)
					ORDER BY v.updated_at DESC
					LIMIT $%d OFFSET $%d
				`, strings.Join([]string{"$1", "$2"}, ", "), len(specializationIDs)+1, len(specializationIDs)+2))
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						"Software Engineer",
						true,
						1,
						1,
						"hybrid",
						"full_time",
						"5/2",
						40,
						50000,
						70000,
						true,
						"3_6_years",
						"Develop software solutions",
						"Write code, review PRs",
						"Go, SQL",
						"Docker",
						"Moscow",
						now,
						now,
					)
				rows.CloseError(errors.New("rows error"))
				mock.ExpectQuery(query).
					WithArgs(1, 2, limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:              "Ошибка закрытия строк",
			specializationIDs: []int{1, 2},
			limit:             2,
			offset:            0,
			expectedResult:    nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обработке результатов запроса вакансий: %w", errors.New("close error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationIDs []int, limit, offset int) {
				query := regexp.QuoteMeta(fmt.Sprintf(`
					SELECT v.id, v.title, v.is_active, v.employer_id, v.specialization_id, v.work_format,
						v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to,
						v.taxes_included, v.experience, v.description, v.tasks, v.requirements,
						v.optional_requirements, v.city, v.created_at, v.updated_at
					FROM vacancy v
					WHERE v.specialization_id IN (%s)
					ORDER BY v.updated_at DESC
					LIMIT $%d OFFSET $%d
				`, strings.Join([]string{"$1", "$2"}, ", "), len(specializationIDs)+1, len(specializationIDs)+2))
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						"Software Engineer",
						true,
						1,
						1,
						"hybrid",
						"full_time",
						"5/2",
						40,
						50000,
						70000,
						true,
						"3_6_years",
						"Develop software solutions",
						"Write code, review PRs",
						"Go, SQL",
						"Docker",
						"Moscow",
						now,
						now,
					)
				rows.CloseError(errors.New("close error"))
				mock.ExpectQuery(query).
					WithArgs(1, 2, limit, offset).
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

			tc.setupMock(mock, tc.specializationIDs, tc.limit, tc.offset)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.SearchVacanciesBySpecializations(ctx, tc.specializationIDs, tc.limit, tc.offset)

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
					require.Equal(t, expected.Title, result[i].Title)
					require.Equal(t, expected.IsActive, result[i].IsActive)
					require.Equal(t, expected.EmployerID, result[i].EmployerID)
					require.Equal(t, expected.SpecializationID, result[i].SpecializationID)
					require.Equal(t, expected.WorkFormat, result[i].WorkFormat)
					require.Equal(t, expected.Employment, result[i].Employment)
					require.Equal(t, expected.Schedule, result[i].Schedule)
					require.Equal(t, expected.WorkingHours, result[i].WorkingHours)
					require.Equal(t, expected.SalaryFrom, result[i].SalaryFrom)
					require.Equal(t, expected.SalaryTo, result[i].SalaryTo)
					require.Equal(t, expected.TaxesIncluded, result[i].TaxesIncluded)
					require.Equal(t, expected.Experience, result[i].Experience)
					require.Equal(t, expected.Description, result[i].Description)
					require.Equal(t, expected.Tasks, result[i].Tasks)
					require.Equal(t, expected.Requirements, result[i].Requirements)
					require.Equal(t, expected.OptionalRequirements, result[i].OptionalRequirements)
					require.Equal(t, expected.City, result[i].City)
					require.False(t, result[i].CreatedAt.IsZero())
					require.False(t, result[i].UpdatedAt.IsZero())
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_CreateLike(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		INSERT INTO vacancy_like (
			vacancy_id,
			applicant_id,
			liked_at
		) VALUES ($1, $2, NOW())
	`)

	testCases := []struct {
		name        string
		vacancyID   int
		applicantID int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, vacancyID, applicantID int)
	}{
		{
			name:        "Успешное создание лайка",
			vacancyID:   1,
			applicantID: 1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID, applicantID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:        "Нарушение внешнего ключа",
			vacancyID:   999,
			applicantID: 1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("vacancy or applicant does not exist"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID, applicantID).
					WillReturnError(&pq.Error{Code: "23503"})
			},
		},
		{
			name:        "Нарушение уникальности",
			vacancyID:   1,
			applicantID: 1,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("response already exists"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID, applicantID).
					WillReturnError(&pq.Error{Code: "23505"})
			},
		},
		{
			name:        "Общая ошибка базы данных",
			vacancyID:   1,
			applicantID: 1,
			expectedErr: fmt.Errorf("failed to create vacancy response: database error"),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID, applicantID).
					WillReturnError(errors.New("database error"))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.vacancyID, tc.applicantID)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			err = repo.CreateLike(ctx, tc.vacancyID, tc.applicantID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				if exp, ok := tc.expectedErr.(entity.Error); ok {
					var actual entity.Error
					require.ErrorAs(t, err, &actual)
					require.Equal(t, exp.Error(), actual.Error())
				} else {
					require.EqualError(t, err, tc.expectedErr.Error())
				}
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_DeleteLike(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		DELETE FROM vacancy_like
		WHERE vacancy_id = $1 AND applicant_id = $2
	`)

	testCases := []struct {
		name        string
		vacancyID   int
		applicantID int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, vacancyID, applicantID int)
	}{
		{
			name:        "Успешное удаление лайка",
			vacancyID:   1,
			applicantID: 1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID, applicantID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:        "Лайк не найден",
			vacancyID:   1,
			applicantID: 1,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("like not found for vacancy %d and applicant %d", 1, 1),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID, applicantID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:        "Ошибка при удалении",
			vacancyID:   1,
			applicantID: 1,
			expectedErr: fmt.Errorf("failed to delete vacancy like: %w", errors.New("database error")),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID, applicantID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:        "Ошибка получения количества затронутых строк",
			vacancyID:   1,
			applicantID: 1,
			expectedErr: fmt.Errorf("failed to get rows affected: %w", errors.New("rows affected error")),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectExec(query).
					WithArgs(vacancyID, applicantID).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))
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

			tc.setupMock(mock, tc.vacancyID, tc.applicantID)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			err = repo.DeleteLike(ctx, tc.vacancyID, tc.applicantID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				if _, ok := tc.expectedErr.(entity.Error); ok {
					var repoErr entity.Error
					require.ErrorAs(t, err, &repoErr)
					require.Equal(t, tc.expectedErr.Error(), err.Error())
				} else {
					require.EqualError(t, err, tc.expectedErr.Error())
				}
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_LikeExists(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM vacancy_like WHERE vacancy_id = $1 AND applicant_id = $2)`)

	testCases := []struct {
		name           string
		vacancyID      int
		applicantID    int
		expectedResult bool
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, vacancyID, applicantID int)
	}{
		{
			name:           "Лайк существует",
			vacancyID:      1,
			applicantID:    1,
			expectedResult: true,
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
				mock.ExpectQuery(query).
					WithArgs(vacancyID, applicantID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Лайк не существует",
			vacancyID:      1,
			applicantID:    1,
			expectedResult: false,
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
				mock.ExpectQuery(query).
					WithArgs(vacancyID, applicantID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка базы данных",
			vacancyID:      1,
			applicantID:    1,
			expectedResult: false,
			expectedErr:    errors.New("database error"),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID, applicantID int) {
				mock.ExpectQuery(query).
					WithArgs(vacancyID, applicantID).
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
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.vacancyID, tc.applicantID)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.LikeExists(ctx, tc.vacancyID, tc.applicantID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Equal(t, tc.expectedResult, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
