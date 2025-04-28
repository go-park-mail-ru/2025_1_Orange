package postgres

import (
	"ResuMatch/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
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

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

	now := time.Now()

	repo := &VacancyRepository{DB: db}
	insertQuery := `INSERT INTO vacancy (
	employer_id, title, specialization_id, work_format, 
	employment, schedule, working_hours, salary_from, 
	salary_to, taxes_included, experience, description, 
	tasks, requirements, optional_requirements, city, 
	created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
		RETURNING id, employer_id, title, is_active, specialization_id, 
		work_format, employment, schedule, working_hours, salary_from, 
		salary_to, taxes_included, experience, description, tasks, 
		requirements, optional_requirements, city, created_at, updated_at`

	tests := []struct {
		name        string
		vacancy     *entity.Vacancy
		mock        func()
		want        *entity.Vacancy
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Успешное создание вакансии",
			vacancy: &entity.Vacancy{
				ID:                   1,
				EmployerID:           1,
				Title:                "Backend Developer",
				IsActive:             true,
				SpecializationID:     1,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "flexible",
				WorkingHours:         40,
				SalaryFrom:           100000,
				SalaryTo:             150000,
				TaxesIncluded:        true,
				Experience:           "1-3 years",
				Description:          "Разработка backend на Go",
				Tasks:                "Разработка API, оптимизация запросов",
				Requirements:         "Опыт работы с Go от 1 года",
				OptionalRequirements: "Знание Docker, Kubernetes",
				City:                 "Москва",
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			mock: func() {
				exactQuery := insertQuery
				mock.ExpectQuery(regexp.QuoteMeta(exactQuery)).
					WithArgs(
						1,                          // employer_id
						"Backend Developer",        // title
						1,                          // specialization_id
						"remote",                   // work_format
						"full_time",                // employment
						"flexible",                 // schedule
						40,                         // working_hours
						100000,                     // salary_from
						150000,                     // salary_to
						true,                       // taxes_included
						"1-3 years",                // experience
						"Разработка backend на Go", // description
						"Разработка API, оптимизация запросов", // tasks
						"Опыт работы с Go от 1 года",           // requirements
						"Знание Docker, Kubernetes",            // optional_requirements
						"Москва",
					).
					WillReturnRows(
						sqlmock.NewRows([]string{
							"id", "employer_id", "title", "is_active", "specialization_id",
							"work_format", "employment", "schedule", "working_hours", "salary_from",
							"salary_to", "taxes_included", "experience", "description", "tasks",
							"requirements", "optional_requirements", "city", "created_at", "updated_at",
						}).AddRow(
							1,                          // id
							1,                          // employer_id
							"Backend Developer",        // title
							true,                       // is_active
							1,                          // specialization_id
							"remote",                   // work_format
							"full_time",                // employment
							"flexible",                 // schedule
							40,                         // working_hours
							100000,                     // salary_from
							150000,                     // salary_to
							true,                       // taxes_included
							"1-3 years",                // experience
							"Разработка backend на Go", // description
							"Разработка API, оптимизация запросов", // tasks
							"Опыт работы с Go от 1 года",           // requirements
							"Знание Docker, Kubernetes",            // optional_requirements
							"Москва",                               // city
							now,                                    // created_at
							now,                                    // updated_at
						))
			},
			want: &entity.Vacancy{
				ID:                   1,
				EmployerID:           1,
				Title:                "Backend Developer",
				IsActive:             true,
				SpecializationID:     1,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "flexible",
				WorkingHours:         40,
				SalaryFrom:           100000,
				SalaryTo:             150000,
				TaxesIncluded:        true,
				Experience:           "1-3 years",
				Description:          "Разработка backend на Go",
				Tasks:                "Разработка API, оптимизация запросов",
				Requirements:         "Опыт работы с Go от 1 года",
				OptionalRequirements: "Знание Docker, Kubernetes",
				City:                 "Москва",
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			wantErr: false,
		},
		{
			name: "Ошибка - обязательное поле отсутствует",
			vacancy: &entity.Vacancy{
				// Title отсутствует - обязательное поле
				EmployerID: 1,
			},
			mock:        func() {}, // Мок не нужен, так как валидация до запроса
			wantErr:     true,
			expectedErr: "bad request\nобязательное поле отсутствует",
		},
		{
			name: "Ошибка - нарушение уникальности",
			vacancy: &entity.Vacancy{
				ID:                   1,
				EmployerID:           1,
				Title:                "Backend Developer",
				IsActive:             true,
				SpecializationID:     1,
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "flexible",
				WorkingHours:         40,
				SalaryFrom:           100000,
				SalaryTo:             150000,
				TaxesIncluded:        true,
				Experience:           "1-3 years",
				Description:          "Разработка backend на Go",
				Tasks:                "Разработка API, оптимизация запросов",
				Requirements:         "Опыт работы с Go от 1 года",
				OptionalRequirements: "Знание Docker, Kubernetes",
				City:                 "Москва",
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			mock: func() {
				exactQuery := insertQuery

				mock.ExpectQuery(regexp.QuoteMeta(exactQuery)).
					WithArgs(
						1,                          // employer_id
						"Backend Developer",        // title
						1,                          // specialization_id
						"remote",                   // work_format
						"full_time",                // employment
						"flexible",                 // schedule
						40,                         // working_hours
						100000,                     // salary_from
						150000,                     // salary_to
						true,                       // taxes_included
						"1-3 years",                // experience
						"Разработка backend на Go", // description
											"Разработка API, оптимизация запросов", // tasks
											"Опыт работы с Go от 1 года",           // requirements
											"Знание Docker, Kubernetes",            // optional_requirements
											"Москва").
					WillReturnError(&pq.Error{Code: "23505"}) // Код нарушения уникальности
			},
			wantErr:     true,
			expectedErr: "already exists\nвакансия с такими параметрами уже существует",
		},
		{
			name: "Ошибка - нарушение NOT NULL",
			vacancy: &entity.Vacancy{
				Title:            "Backend Developer",
				SpecializationID: 1,
				// EmployerID отсутствует (должен быть NOT NULL)
			},
			mock: func() {
				exactQuery := insertQuery

				mock.ExpectQuery(regexp.QuoteMeta(exactQuery)).
					WithArgs(
						nil,
						"Backend Developer",
						1,
						"", "", "", 0, 0, 0, false, "", "", "", "", "", "",
					).
					WillReturnError(&pq.Error{Code: "23502"}) // Код нарушения NOT NULL
			},
			wantErr:     true,
			expectedErr: "bad request\nобязательное поле отсутствует",
		},
		{
			name: "Ошибка - нарушение типа данных",
			vacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Backend Developer",
				SpecializationID: 1,
				WorkingHours:     12, // должно быть число
			},
			mock: func() {
				exactQuery := `...`
				mock.ExpectQuery(regexp.QuoteMeta(exactQuery)).
					WithArgs(
						1,
						"Backend Developer",
						1,
						"", "", "", "forty", 0, 0, false, "", "", "", "", "", "",
					).
					WillReturnError(&pq.Error{Code: "42804"}) // Код нарушения типа данных
			},
			wantErr:     true,
			expectedErr: "bad request\nнеправильный формат данных",
		},
		{
			name: "Ошибка - нарушение CHECK",
			vacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Backend Developer",
				SpecializationID: 1,
				SalaryFrom:       -100, // отрицательная зарплата
			},
			mock: func() {
				exactQuery := `...`
				mock.ExpectQuery(regexp.QuoteMeta(exactQuery)).
					WithArgs(
						1,
						"Backend Developer",
						1,
						"", "", "", 0, -100, 0, false, "", "", "", "", "", "",
					).
					WillReturnError(&pq.Error{Code: "23514"}) // Код нарушения CHECK
			},
			wantErr:     true,
			expectedErr: "bad request\nнеправильные данные",
		},
		{
			name: "Ошибка - внутренняя ошибка сервера",
			vacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Backend Developer",
				SpecializationID: 1,
			},
			mock: func() {
				exactQuery := `...`
				mock.ExpectQuery(regexp.QuoteMeta(exactQuery)).
					WithArgs(
						1,
						"Backend Developer",
						1,
						"", "", "", 0, 0, 0, false, "", "", "", "", "", "",
					).
					WillReturnError(fmt.Errorf("connection timeout")) // Не PostgreSQL ошибка
			},
			wantErr:     true,
			expectedErr: "internal server error\nошибка при создании вакансии: connection timeout",
		},
		{
			name: "Ошибка - нарушение NOT NULL (PSQLNotNullViolation)",
			vacancy: &entity.Vacancy{
				Title:            "Backend Developer",
				SpecializationID: 1,
				// EmployerID отсутствует (должен быть NOT NULL)
			},
			mock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(insertQuery)).
					WithArgs(
						nil, // employer_id = NULL
						"Backend Developer",
						1,
						"", "", "", 0, 0, 0, false, "", "", "", "", "", "",
					).
					WillReturnError(&pq.Error{Code: "23502"}) // Код нарушения NOT NULL
			},
			wantErr:     true,
			expectedErr: "bad request\nобязательное поле отсутствует",
		},
		{
			name: "Ошибка - нарушение типа данных (PSQLDatatypeViolation)",
			vacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Backend Developer",
				SpecializationID: 1,
				WorkingHours:     0,
				// Предположим, что WorkingHours передано как строка в SQL запросе
			},
			mock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(insertQuery)).
					WithArgs(
						1,
						"Backend Developer",
						1,
						"", "", "", "invalid", 0, 0, false, "", "", "", "", "", "", // WorkingHours как строка
					).
					WillReturnError(&pq.Error{Code: "42804"}) // Код нарушения типа данных
			},
			wantErr:     true,
			expectedErr: "bad request\nнеправильный формат данных",
		},
		{
			name: "Ошибка - нарушение CHECK (PSQLCheckViolation)",
			vacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Backend Developer",
				SpecializationID: 1,
				SalaryFrom:       -100, // отрицательная зарплата
			},
			mock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(insertQuery)).
					WithArgs(
						1,
						"Backend Developer",
						1,
						"", "", "", 0, -100, 0, false, "", "", "", "", "", "",
					).
					WillReturnError(&pq.Error{Code: "23514"}) // Код нарушения CHECK
			},
			wantErr:     true,
			expectedErr: "bad request\nнеправильные данные",
		},
		{
			name: "Ошибка - другая ошибка PostgreSQL",
			vacancy: &entity.Vacancy{
				EmployerID:       1,
				Title:            "Backend Developer",
				SpecializationID: 1,
			},
			mock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(insertQuery)).
					WithArgs(
						1,
						"Backend Developer",
						1,
						"", "", "", 0, 0, 0, false, "", "", "", "", "", "",
					).
					WillReturnError(&pq.Error{Code: "XX000", Message: "internal database error"}) // Другая ошибка PostgreSQL
			},
			wantErr:     true,
			expectedErr: "internal server error\nошибка при создании вакансии: pq: internal database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.Create(context.Background(), tt.vacancy)

			if (err != nil) != tt.wantErr {
				t.Errorf("VacancyRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err.Error() != tt.expectedErr {
					t.Errorf("VacancyRepository.Create() error = %v, expectedErr %v", err.Error(), tt.expectedErr)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VacancyRepository.Create() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
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
		expectedResult *entity.Vacancy
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name: "Успешное получение вакансии",
			id:   1,
			expectedResult: &entity.Vacancy{
				ID:                   1,
				Title:                "Разработчик Go",
				EmployerID:           101,
				SpecializationID:     1,
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
								1, "Разработчик Go", 101, 1, "remote", // specialization_id как число
								"full_time", "flexible", 40, 100000, 150000, true, "no_experience",
								"Работа с высоконагруженными системами.", "Разработка новых сервисов.",
								"Go, PostgreSQL.", "Docker, Kubernetes.", "Москва", now, now,
							),
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
			defer func() {
				err := db.Close()
				require.NoError(t, err)
			}()

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
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_AddCity(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

	repo := &VacancyRepository{DB: db}

	tests := []struct {
		name        string
		vacancyID   int
		cityIDs     []int
		mock        func()
		wantErr     bool
		expectedErr string
	}{
		{
			name:      "Успешное добавление городов",
			vacancyID: 1,
			cityIDs:   []int{1, 2, 3},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WithArgs(1, 2).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WithArgs(1, 3).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:      "Ошибка начала транзакции",
			vacancyID: 1,
			cityIDs:   []int{1},
			mock: func() {
				mock.ExpectBegin().WillReturnError(fmt.Errorf("tx begin error"))
			},
			wantErr:     true,
			expectedErr: "internal server error\nошибка при начале транзакции для добавления городов: tx begin error",
		},
		{
			name:      "Ошибка подготовки запроса",
			vacancyID: 1,
			cityIDs:   []int{1},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WillReturnError(fmt.Errorf("prepare error"))
				mock.ExpectRollback()
			},
			wantErr:     true,
			expectedErr: "internal server error\nошибка при подготовке запроса для добавления городов: prepare error",
		},
		{
			name:      "Нарушение уникальности - пропускаем город",
			vacancyID: 1,
			cityIDs:   []int{1, 2, 3},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WithArgs(1, 2).WillReturnError(&pq.Error{Code: "23505"}) // Unique violation
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WithArgs(1, 3).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:      "Ошибка NOT NULL violation",
			vacancyID: 1,
			cityIDs:   []int{1},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WithArgs(1, 1).WillReturnError(&pq.Error{Code: "23502"}) // NOT NULL violation
				mock.ExpectRollback()
			},
			wantErr:     true,
			expectedErr: "bad request\nобязательное поле отсутствует",
		},
		{
			name:      "Ошибка Datatype violation",
			vacancyID: 1,
			cityIDs:   []int{1},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WithArgs(1, 1).WillReturnError(&pq.Error{Code: "42804"}) // Datatype violation
				mock.ExpectRollback()
			},
			wantErr:     true,
			expectedErr: "bad request\nнеправильный формат данных",
		},
		{
			name:      "Ошибка Check violation",
			vacancyID: 1,
			cityIDs:   []int{1},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WithArgs(1, 1).WillReturnError(&pq.Error{Code: "23514"}) // Check violation
				mock.ExpectRollback()
			},
			wantErr:     true,
			expectedErr: "bad request\nнеправильные данные",
		},
		{
			name:      "Ошибка выполнения запроса",
			vacancyID: 1,
			cityIDs:   []int{1},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WithArgs(1, 1).WillReturnError(fmt.Errorf("exec error"))
				mock.ExpectRollback()
			},
			wantErr:     true,
			expectedErr: "internal server error\nошибка при добавлении города к вакансии: exec error",
		},
		{
			name:      "Ошибка коммита транзакции",
			vacancyID: 1,
			cityIDs:   []int{1},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vacancy_city (vacancy_id, city_id) VALUES ($1, $2)`)).
					WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))
			},
			wantErr:     true,
			expectedErr: "internal server error\nошибка при коммите транзакции добавления городов: commit error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.AddCity(context.Background(), tt.vacancyID, tt.cityIDs)

			if (err != nil) != tt.wantErr {
				t.Errorf("VacancyRepository.AddCity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err.Error() != tt.expectedErr {
					t.Errorf("VacancyRepository.AddCity() error = %v, expectedErr %v", err.Error(), tt.expectedErr)
				}
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestVacancyRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

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
				fmt.Errorf("не удалось удалить вакансию с id=2"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM vacancy
                    WHERE id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(id).
					WillReturnError(errors.New("database error"))
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
					WillReturnResult(result)
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
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

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
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

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
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

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
				// Правильный запрос для получения резюме
				mock.ExpectQuery("SELECT id FROM resume WHERE applicant_id = $1 ORDER BY created_at DESC LIMIT 1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Запрос для создания отклика
				mock.ExpectExec("INSERT INTO vacancy_response (vacancy_id, applicant_id, resume_id, applied_at) VALUES ($1, $2, $3, NOW())").
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
				mock.ExpectQuery("SELECT id FROM resume WHERE applicant_id = $1 ORDER BY created_at DESC LIMIT 1").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			vacancyID:   1,
			applicantID: 1,
			expectedErr: entity.NewError(entity.ErrNotFound, fmt.Errorf("no active resumes found for applicant")),
		},
		{
			name: "No Vacancy",
			mock: func() {
				mock.ExpectQuery("SELECT id FROM resume WHERE applicant_id = $1 ORDER BY created_at DESC LIMIT 1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				mock.ExpectExec("INSERT INTO vacancy_response (vacancy_id, applicant_id, resume_id, applied_at) VALUES ($1, $2, $3, NOW())").
					WithArgs(1, 1, 1).
					WillReturnError(&pq.Error{Code: "23503"}) // foreign key violation
			},
			vacancyID:   1,
			applicantID: 1,
			expectedErr: entity.NewError(entity.ErrBadRequest, fmt.Errorf("vacancy or applicant does not exist")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.CreateResponse(context.Background(), tt.vacancyID, tt.applicantID)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_AddSkills(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		vacancyID   int
		skillIDs    []int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int)
	}{
		{
			name:        "Успешное добавление навыков",
			vacancyID:   1,
			skillIDs:    []int{1, 2, 3},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO vacancy_skill (vacancy_id, skill_id)
					VALUES ($1, $2)
				`))

				for _, skillID := range skillIDs {
					stmt.ExpectExec().
						WithArgs(vacancyID, skillID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}

				mock.ExpectCommit()
			},
		},

		{
			name:      "Ошибка начала транзакции",
			vacancyID: 1,
			skillIDs:  []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при начале транзакции для добавления навыков: %w", errors.New("tx error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin().WillReturnError(errors.New("tx error"))
			},
		},
		{
			name:      "Ошибка подготовки запроса",
			vacancyID: 1,
			skillIDs:  []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при подготовке запроса для добавления навыков: %w", errors.New("prepare error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO vacancy_skill (vacancy_id, skill_id)
					VALUES ($1, $2)
				`)).WillReturnError(errors.New("prepare error"))
				mock.ExpectRollback()
			},
		},
		{
			name:      "Ошибка выполнения запроса",
			vacancyID: 1,
			skillIDs:  []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении навыка к резюме: %w", errors.New("exec error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO vacancy_skill (vacancy_id, skill_id)
					VALUES ($1, $2)
				`))

				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[0]).
					WillReturnError(errors.New("exec error"))

				mock.ExpectRollback()
			},
		},
		{
			name:        "Нарушение уникальности (пропускаем дубликаты)",
			vacancyID:   1,
			skillIDs:    []int{1, 2, 3},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO vacancy_skill (vacancy_id, skill_id)
					VALUES ($1, $2)
				`))

				// Первый навык добавляется успешно
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[0]).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Второй навык - дубликат
				pqErr := &pq.Error{Code: entity.PSQLUniqueViolation}
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[1]).
					WillReturnError(pqErr)

				// Третий навык добавляется успешно
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[2]).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
		},
		{
			name:      "Отсутствует обязательное поле",
			vacancyID: 1,
			skillIDs:  []int{1},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO vacancy_skill (vacancy_id, skill_id)
					VALUES ($1, $2)
				`))

				pqErr := &pq.Error{Code: entity.PSQLNotNullViolation}
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[0]).
					WillReturnError(pqErr)

				mock.ExpectRollback()
			},
		},
		{
			name:      "Неправильный формат данных",
			vacancyID: 1,
			skillIDs:  []int{1},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO vacancy_skill (vacancy_id, skill_id)
					VALUES ($1, $2)
				`))

				pqErr := &pq.Error{Code: entity.PSQLDatatypeViolation}
				stmt.ExpectExec().
					WithArgs(vacancyID, skillIDs[0]).
					WillReturnError(pqErr)

				mock.ExpectRollback()
			},
		},
		{
			name:      "Ошибка коммита транзакции",
			vacancyID: 1,
			skillIDs:  []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при коммите транзакции добавления навыков: %w", errors.New("commit error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO vacancy_skill (vacancy_id, skill_id)
					VALUES ($1, $2)
				`))

				for _, skillID := range skillIDs {
					stmt.ExpectExec().
						WithArgs(vacancyID, skillID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}

				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
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

func TestVacancyRepository_FindSkillIDsByNames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		skillNames     []string
		expectedResult []int
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, skillNames []string)
	}{
		{
			name:           "Успешное получение ID существующих навыков",
			skillNames:     []string{"Go", "SQL", "Docker"},
			expectedResult: []int{1, 2, 3},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				// Мокаем запрос на поиск существующих навыков
				query := regexp.QuoteMeta(`SELECT id FROM skill WHERE name = $1`)
				mock.ExpectQuery(query).WithArgs("Go").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectQuery(query).WithArgs("SQL").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				mock.ExpectQuery(query).WithArgs("Docker").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
			},
		},
		{
			name:           "Создание новых навыков",
			skillNames:     []string{"NewSkill1", "NewSkill2"},
			expectedResult: []int{4, 5},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				// Для NewSkill1 - сначала не найден, затем создан
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM skill WHERE name = $1`)).
					WithArgs("NewSkill1").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO skill (name) VALUES ($1) RETURNING id`)).
					WithArgs("NewSkill1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))

				// Для NewSkill2 - сначала не найден, затем создан
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM skill WHERE name = $1`)).
					WithArgs("NewSkill2").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO skill (name) VALUES ($1) RETURNING id`)).
					WithArgs("NewSkill2").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
			},
		},
		{
			name:           "Пустой список навыков",
			skillNames:     []string{},
			expectedResult: []int{},
			expectedErr:    nil,
			setupMock:      func(mock sqlmock.Sqlmock, skillNames []string) {},
		},
		{
			name:           "Ошибка при поиске навыка",
			skillNames:     []string{"Go"},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка при поиске навыка"),
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM skill WHERE name = $1`)).
					WithArgs("Go").
					WillReturnError(fmt.Errorf("ошибка при поиске навыка"))
			},
		},
		{
			name:           "Ошибка при создании навыка",
			skillNames:     []string{"NewSkill"},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка при создании навыка"),
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM skill WHERE name = $1`)).
					WithArgs("NewSkill").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO skill (name) VALUES ($1) RETURNING id`)).
					WithArgs("NewSkill").
					WillReturnError(fmt.Errorf("ошибка при создании навыка"))
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

			tc.setupMock(mock, tc.skillNames)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			result, err := repo.FindSkillIDsByNames(ctx, tc.skillNames)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
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

	testCases := []struct {
		name           string
		cityNames      []string
		expectedResult []int
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, cityNames []string)
	}{
		{
			name:           "Успешное получение ID существующих городов",
			cityNames:      []string{"Москва", "Санкт-Петербург", "Казань"},
			expectedResult: []int{1, 2, 3},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, cityNames []string) {
				query := regexp.QuoteMeta(`
                    SELECT id
                    FROM city
                    WHERE name IN ($1, $2, $3)
                `)
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1). // Москва
					AddRow(2). // Санкт-Петербург
					AddRow(3)  // Казань
				mock.ExpectQuery(query).
					WithArgs("Москва", "Санкт-Петербург", "Казань").
					WillReturnRows(rows)
			},
		},
		{
			name:           "Частичное нахождение городов",
			cityNames:      []string{"Москва", "Новосибирск", "Екатеринбург"},
			expectedResult: []int{1, 4}, // Новосибирск не найден
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, cityNames []string) {
				query := regexp.QuoteMeta(`
                    SELECT id
                    FROM city
                    WHERE name IN ($1, $2, $3)
                `)
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1). // Москва
					AddRow(4)  // Екатеринбург
				mock.ExpectQuery(query).
					WithArgs("Москва", "Новосибирск", "Екатеринбург").
					WillReturnRows(rows)
			},
		},
		{
			name:           "Пустой список городов",
			cityNames:      []string{},
			expectedResult: []int{},
			expectedErr:    nil,
			setupMock:      func(mock sqlmock.Sqlmock, cityNames []string) {},
		},
		{
			name:           "Ошибка выполнения запроса",
			cityNames:      []string{"Москва", "Санкт-Петербург"},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске ID городов по названиям: %w", fmt.Errorf("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, cityNames []string) {
				query := regexp.QuoteMeta(`
                    SELECT id
                    FROM city
                    WHERE name IN ($1, $2)
                `)
				mock.ExpectQuery(query).
					WithArgs("Москва", "Санкт-Петербург").
					WillReturnError(fmt.Errorf("database error"))
			},
		},
		{
			name:           "Ошибка сканирования результата",
			cityNames:      []string{"Москва"},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании ID города: %w", fmt.Errorf("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, cityNames []string) {
				query := regexp.QuoteMeta(`
                    SELECT id
                    FROM city
                    WHERE name IN ($1)
                `)
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow("invalid") // Неправильный тип данных
				mock.ExpectQuery(query).
					WithArgs("Москва").
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка при итерации по результатам",
			cityNames:      []string{"Москва", "Санкт-Петербург"},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по ID городов: %w", fmt.Errorf("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, cityNames []string) {
				query := regexp.QuoteMeta(`
                    SELECT id
                    FROM city
                    WHERE name IN ($1, $2)
                `)
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(1).
					RowError(0, fmt.Errorf("rows error"))
				mock.ExpectQuery(query).
					WithArgs("Москва", "Санкт-Петербург").
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

			tc.setupMock(mock, tc.cityNames)

			repo := &VacancyRepository{
				DB: db,
			}
			ctx := context.Background()

			result, err := repo.FindCityIDsByNames(ctx, tc.cityNames)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
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

	testCases := []struct {
		name               string
		specializationName string
		expectedID         int
		expectedErr        error
		setupMock          func(mock sqlmock.Sqlmock, name string)
	}{
		{
			name:               "Успешное получение ID существующей специализации",
			specializationName: "Backend разработка",
			expectedID:         1,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				// Сначала проверка существования
				mock.ExpectQuery(regexp.QuoteMeta(`
                    SELECT id FROM specialization WHERE name = $1
                `)).
					WithArgs(name).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name:               "Создание новой специализации",
			specializationName: "Новая специализация",
			expectedID:         2,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				// Проверка существования - не найдено
				mock.ExpectQuery(regexp.QuoteMeta(`
                    SELECT id FROM specialization WHERE name = $1
                `)).
					WithArgs(name).
					WillReturnError(sql.ErrNoRows)

				// Создание новой специализации
				mock.ExpectQuery(regexp.QuoteMeta(`
                    INSERT INTO specialization (name) VALUES ($1) RETURNING id
                `)).
					WithArgs(name).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
		},
		{
			name:               "Ошибка при проверке существования",
			specializationName: "Специализация с ошибкой",
			expectedID:         0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования специализации: %w", errors.New("db error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				mock.ExpectQuery(regexp.QuoteMeta(`
                    SELECT id FROM specialization WHERE name = $1
                `)).
					WithArgs(name).
					WillReturnError(errors.New("db error"))
			},
		},
		{
			name:               "Ошибка при создании",
			specializationName: "Специализация с ошибкой создания",
			expectedID:         0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании специализации: %w", errors.New("create error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				// Проверка существования - не найдено
				mock.ExpectQuery(regexp.QuoteMeta(`
                    SELECT id FROM specialization WHERE name = $1
                `)).
					WithArgs(name).
					WillReturnError(sql.ErrNoRows)

				// Ошибка при создании
				mock.ExpectQuery(regexp.QuoteMeta(`
                    INSERT INTO specialization (name) VALUES ($1) RETURNING id
                `)).
					WithArgs(name).
					WillReturnError(errors.New("create error"))
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

			tc.setupMock(mock, tc.specializationName)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			id, err := repo.FindSpecializationIDByName(ctx, tc.specializationName)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				if errors.As(err, &repoErr) {
					require.Equal(t, tc.expectedErr.Error(), err.Error())
				} else {
					require.Fail(t, "Ожидалась ошибка типа entity.Error")
				}
				require.Equal(t, tc.expectedID, id)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, id)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_CreateSpecializationIfNotExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		specializationName string
		expectedID         int
		expectedErr        error
		setupMock          func(mock sqlmock.Sqlmock, name string)
	}{
		{
			name:               "Специализация уже существует",
			specializationName: "Backend разработка",
			expectedID:         1,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				// Мок для проверки существования
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id
					FROM specialization
					WHERE name = $1
				`)).
					WithArgs(name).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name:               "Специализация не существует - успешное создание",
			specializationName: "New Specialization",
			expectedID:         2,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				// Мок для проверки существования (вернет ErrNoRows)
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id
					FROM specialization
					WHERE name = $1
				`)).
					WithArgs(name).
					WillReturnError(sql.ErrNoRows)

				// Мок для создания новой специализации
				mock.ExpectQuery(regexp.QuoteMeta(`
					INSERT INTO specialization (name)
					VALUES ($1)
					RETURNING id
				`)).
					WithArgs(name).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
		},
		{
			name:               "Ошибка при проверке существования",
			specializationName: "Backend разработка",
			expectedID:         0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования специализации: %w", errors.New("db error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id
					FROM specialization
					WHERE name = $1
				`)).
					WithArgs(name).
					WillReturnError(errors.New("db error"))
			},
		},
		{
			name:               "Конфликт при создании (специализация создана другим запросом)",
			specializationName: "Backend разработка",
			expectedID:         3,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				// Первая проверка - специализация не найдена
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id
					FROM specialization
					WHERE name = $1
				`)).
					WithArgs(name).
					WillReturnError(sql.ErrNoRows)

				// Попытка создания - возвращает ошибку уникальности
				mock.ExpectQuery(regexp.QuoteMeta(`
					INSERT INTO specialization (name)
					VALUES ($1)
					RETURNING id
				`)).
					WithArgs(name).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})

				// Повторная проверка - теперь специализация существует
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id
					FROM specialization
					WHERE name = $1
				`)).
					WithArgs(name).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
			},
		},
		{
			name:               "Ошибка при создании специализации",
			specializationName: "New Specialization",
			expectedID:         0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании специализации: %w", errors.New("creation error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				// Первая проверка - специализация не найдена
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id
					FROM specialization
					WHERE name = $1
				`)).
					WithArgs(name).
					WillReturnError(sql.ErrNoRows)

				// Попытка создания - возвращает ошибку
				mock.ExpectQuery(regexp.QuoteMeta(`
					INSERT INTO specialization (name)
					VALUES ($1)
					RETURNING id
				`)).
					WithArgs(name).
					WillReturnError(errors.New("creation error"))
			},
		},
		{
			name:               "Ошибка при повторной проверке после конфликта",
			specializationName: "Backend разработка",
			expectedID:         0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении ID специализации после конфликта: %w", errors.New("check error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				// Первая проверка - специализация не найдена
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id
					FROM specialization
					WHERE name = $1
				`)).
					WithArgs(name).
					WillReturnError(sql.ErrNoRows)

				// Попытка создания - возвращает ошибку уникальности
				mock.ExpectQuery(regexp.QuoteMeta(`
					INSERT INTO specialization (name)
					VALUES ($1)
					RETURNING id
				`)).
					WithArgs(name).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})

				// Повторная проверка - возвращает ошибку
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id
					FROM specialization
					WHERE name = $1
				`)).
					WithArgs(name).
					WillReturnError(errors.New("check error"))
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

			tc.setupMock(mock, tc.specializationName)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			id, err := repo.CreateSpecializationIfNotExists(ctx, tc.specializationName)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, id)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_GetVacanciesByApplicantID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		name           string
		applicantID    int
		expectedResult []*entity.Vacancy
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, applicantID int)
	}{
		{
			name:        "Успешное получение вакансий по applicantID",
			applicantID: 1,
			expectedResult: []*entity.Vacancy{
				{
					ID:                   1,
					Title:                "Backend Developer",
					EmployerID:           1,
					SpecializationID:     1,
					WorkFormat:           "remote",
					Employment:           "full_time",
					Schedule:             "flexible",
					WorkingHours:         40,
					SalaryFrom:           100000,
					SalaryTo:             150000,
					TaxesIncluded:        true,
					Experience:           "1-3 years",
					Description:          "Разработка backend на Go",
					Tasks:                "Разработка API, оптимизация запросов",
					Requirements:         "Опыт работы с Go от 1 года",
					OptionalRequirements: "Знание Docker, Kubernetes",
					City:                 "Москва",
					CreatedAt:            now,
					UpdatedAt:            now,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, applicantID int) {
				query := regexp.QuoteMeta(`
                    SELECT v.id, v.title, v.employer_id, v.specialization_id, v.work_format, 
                           v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to, 
                           v.taxes_included, v.experience, v.description, v.tasks, v.requirements, 
                           v.optional_requirements, v.city, v.created_at, v.updated_at
                    FROM vacancy v
                    JOIN vacancy_response vr ON v.id = vr.vacancy_id
                    WHERE vr.applicant_id = $1
                    ORDER BY vr.applied_at DESC;
                `)
				rows := sqlmock.NewRows([]string{
					"id", "title", "employer_id", "specialization_id", "work_format",
					"employment", "schedule", "working_hours", "salary_from", "salary_to",
					"taxes_included", "experience", "description", "tasks", "requirements",
					"optional_requirements", "city", "created_at", "updated_at",
				}).
					AddRow(
						1, "Backend Developer", 1, 1, "remote",
						"full_time", "flexible", 40, 100000, 150000,
						true, "1-3 years", "Разработка backend на Go",
						"Разработка API, оптимизация запросов", "Опыт работы с Go от 1 года",
						"Знание Docker, Kubernetes", "Москва", now, now,
					)
				mock.ExpectQuery(query).
					WithArgs(applicantID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Нет вакансий для applicantID",
			applicantID:    2,
			expectedResult: []*entity.Vacancy{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, applicantID int) {
				query := regexp.QuoteMeta(`
                    SELECT v.id, v.title, v.employer_id, v.specialization_id, v.work_format, 
                           v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to, 
                           v.taxes_included, v.experience, v.description, v.tasks, v.requirements, 
                           v.optional_requirements, v.city, v.created_at, v.updated_at
                    FROM vacancy v
                    JOIN vacancy_response vr ON v.id = vr.vacancy_id
                    WHERE vr.applicant_id = $1
                    ORDER BY vr.applied_at DESC;
                `)
				rows := sqlmock.NewRows([]string{
					"id", "title", "employer_id", "specialization_id", "work_format",
					"employment", "schedule", "working_hours", "salary_from", "salary_to",
					"taxes_included", "experience", "description", "tasks", "requirements",
					"optional_requirements", "city", "created_at", "updated_at",
				})
				mock.ExpectQuery(query).
					WithArgs(applicantID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка выполнения запроса",
			applicantID:    3,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка вакансий: %w", fmt.Errorf("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicantID int) {
				query := regexp.QuoteMeta(`
                    SELECT v.id, v.title, v.employer_id, v.specialization_id, v.work_format, 
                           v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to, 
                           v.taxes_included, v.experience, v.description, v.tasks, v.requirements, 
                           v.optional_requirements, v.city, v.created_at, v.updated_at
                    FROM vacancy v
                    JOIN vacancy_response vr ON v.id = vr.vacancy_id
                    WHERE vr.applicant_id = $1
                    ORDER BY vr.applied_at DESC;
                `)
				mock.ExpectQuery(query).
					WithArgs(applicantID).
					WillReturnError(fmt.Errorf("database error"))
			},
		},
		{
			name:           "Ошибка сканирования вакансии",
			applicantID:    4,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка обработки данных вакансии: %w", fmt.Errorf("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicantID int) {
				query := regexp.QuoteMeta(`
                    SELECT v.id, v.title, v.employer_id, v.specialization_id, v.work_format, 
                           v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to, 
                           v.taxes_included, v.experience, v.description, v.tasks, v.requirements, 
                           v.optional_requirements, v.city, v.created_at, v.updated_at
                    FROM vacancy v
                    JOIN vacancy_response vr ON v.id = vr.vacancy_id
                    WHERE vr.applicant_id = $1
                    ORDER BY vr.applied_at DESC;
                `)
				rows := sqlmock.NewRows([]string{
					"id", "title", "employer_id", "specialization_id", "work_format",
					"employment", "schedule", "working_hours", "salary_from", "salary_to",
					"taxes_included", "experience", "description", "tasks", "requirements",
					"optional_requirements", "city", "created_at", "updated_at",
				}).
					AddRow(
						"invalid", "Backend Developer", 1, 1, "remote", // Неправильный тип для id
						"full_time", "flexible", 40, 100000, 150000,
						true, "1-3 years", "Разработка backend на Go",
						"Разработка API, оптимизация запросов", "Опыт работы с Go от 1 года",
						"Знание Docker, Kubernetes", "Москва", now, now,
					)
				mock.ExpectQuery(query).
					WithArgs(applicantID).
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

			tc.setupMock(mock, tc.applicantID)

			repo := &VacancyRepository{
				DB: db,
			}
			ctx := context.Background()

			result, err := repo.GetVacanciesByApplicantID(ctx, tc.applicantID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))

				for i, expected := range tc.expectedResult {
					require.Equal(t, expected.ID, result[i].ID)
					require.Equal(t, expected.Title, result[i].Title)
					// Добавьте проверки для остальных полей по необходимости
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_CreateSkillIfNotExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		skillName   string
		expectedID  int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, skillName string)
	}{
		{
			name:        "Навык уже существует",
			skillName:   "Go",
			expectedID:  1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				// Мок для проверки существования навыка
				query := regexp.QuoteMeta(`
					SELECT id
					FROM skill
					WHERE name = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(skillName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name:        "Навык не существует - успешное создание",
			skillName:   "Rust",
			expectedID:  2,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				// Первый запрос - проверка существования (вернет ErrNoRows)
				queryCheck := regexp.QuoteMeta(`
					SELECT id
					FROM skill
					WHERE name = $1
				`)
				mock.ExpectQuery(queryCheck).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)

				// Второй запрос - создание навыка
				queryCreate := regexp.QuoteMeta(`
					INSERT INTO skill (name)
					VALUES ($1)
					RETURNING id
				`)
				mock.ExpectQuery(queryCreate).
					WithArgs(skillName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
		},
		{
			name:       "Ошибка при проверке существования навыка",
			skillName:  "Python",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования навыка: %w", errors.New("db error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				query := regexp.QuoteMeta(`
					SELECT id
					FROM skill
					WHERE name = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(skillName).
					WillReturnError(errors.New("db error"))
			},
		},
		{
			name:        "Конфликт при создании - навык уже создан другим запросом",
			skillName:   "Java",
			expectedID:  3,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				// Первый запрос - проверка существования (вернет ErrNoRows)
				queryCheck := regexp.QuoteMeta(`
					SELECT id
					FROM skill
					WHERE name = $1
				`)
				mock.ExpectQuery(queryCheck).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)

				// Второй запрос - попытка создания (вернет UniqueViolation)
				queryCreate := regexp.QuoteMeta(`
					INSERT INTO skill (name)
					VALUES ($1)
					RETURNING id
				`)
				pqErr := &pq.Error{Code: entity.PSQLUniqueViolation}
				mock.ExpectQuery(queryCreate).
					WithArgs(skillName).
					WillReturnError(pqErr)

				// Третий запрос - повторная проверка существования
				queryCheckAgain := regexp.QuoteMeta(`
					SELECT id
					FROM skill
					WHERE name = $1
				`)
				mock.ExpectQuery(queryCheckAgain).
					WithArgs(skillName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
			},
		},
		{
			name:       "Ошибка при создании навыка",
			skillName:  "C++",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании навыка: %w", errors.New("create error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				// Первый запрос - проверка существования (вернет ErrNoRows)
				queryCheck := regexp.QuoteMeta(`
					SELECT id
					FROM skill
					WHERE name = $1
				`)
				mock.ExpectQuery(queryCheck).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)

				// Второй запрос - попытка создания (вернет ошибку)
				queryCreate := regexp.QuoteMeta(`
					INSERT INTO skill (name)
					VALUES ($1)
					RETURNING id
				`)
				mock.ExpectQuery(queryCreate).
					WithArgs(skillName).
					WillReturnError(errors.New("create error"))
			},
		},
		{
			name:       "Ошибка при повторной проверке после конфликта",
			skillName:  "TypeScript",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении ID навыка после конфликта: %w", errors.New("check after conflict error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				// Первый запрос - проверка существования (вернет ErrNoRows)
				queryCheck := regexp.QuoteMeta(`
					SELECT id
					FROM skill
					WHERE name = $1
				`)
				mock.ExpectQuery(queryCheck).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)

				// Второй запрос - попытка создания (вернет UniqueViolation)
				queryCreate := regexp.QuoteMeta(`
					INSERT INTO skill (name)
					VALUES ($1)
					RETURNING id
				`)
				pqErr := &pq.Error{Code: entity.PSQLUniqueViolation}
				mock.ExpectQuery(queryCreate).
					WithArgs(skillName).
					WillReturnError(pqErr)

				// Третий запрос - повторная проверка существования (вернет ошибку)
				queryCheckAgain := regexp.QuoteMeta(`
					SELECT id
					FROM skill
					WHERE name = $1
				`)
				mock.ExpectQuery(queryCheckAgain).
					WithArgs(skillName).
					WillReturnError(errors.New("check after conflict error"))
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

			tc.setupMock(mock, tc.skillName)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			id, err := repo.CreateSkillIfNotExists(ctx, tc.skillName)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, id)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_DeleteCity(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		vacancyID   int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, vacancyID int)
	}{
		{
			name:        "Успешное удаление городов вакансии",
			vacancyID:   1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM vacancy_city
                    WHERE vacancy_id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:        "Вакансия не существует (нет городов для удаления)",
			vacancyID:   999,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM vacancy_city
                    WHERE vacancy_id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:      "Ошибка базы данных",
			vacancyID: 2,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении городов вакансии: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM vacancy_city
                    WHERE vacancy_id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:      "Неверный ID вакансии",
			vacancyID: 0,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный ID вакансии"),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				// Мок не нужен, так как проверка ID происходит до запроса
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

			tc.setupMock(mock, tc.vacancyID)

			// Инициализируем репозиторий с моком базы данных
			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			// Вызываем тестируемую функцию
			err = repo.DeleteCity(ctx, tc.vacancyID)

			// Проверяем результаты
			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			// Проверяем, что все ожидания по моку выполнены
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestVacancyRepository_DeleteSkills(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		vacancyID   int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, vacancyID int)
	}{
		{
			name:        "Успешное удаление навыков резюме",
			vacancyID:   1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM vacancy_skill
                    WHERE vacancy_id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:        "Резюме не существует (нет навыков для удаления)",
			vacancyID:   999,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM vacancy_skill
                    WHERE vacancy_id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(vacancyID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:      "Ошибка базы данных",
			vacancyID: 2,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении навыков резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, vacancyID int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM vacancy_skill
                    WHERE vacancy_id = $1
                `)
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
			defer func() {
				err := db.Close()
				require.NoError(t, err)
			}()

			tc.setupMock(mock, tc.vacancyID)

			repo := &VacancyRepository{DB: db}
			ctx := context.Background()

			// Добавляем валидацию ID перед вызовом метода
			if tc.vacancyID <= 0 {
				err := repo.DeleteSkills(ctx, tc.vacancyID)
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				return
			}

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

func TestVacancyRepository_GetCityByVacancyID(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

	repo := &VacancyRepository{DB: db}

	//now := time.Now()
	testCities := []entity.City{
		{ID: 1, Name: "Москва"},
		{ID: 2, Name: "Санкт-Петербург"},
	}

	tests := []struct {
		name        string
		vacancyID   int
		mock        func()
		want        []entity.City
		wantErr     bool
		expectedErr string
	}{
		{
			name:      "Успешное получение городов",
			vacancyID: 1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Москва").
					AddRow(2, "Санкт-Петербург")

				mock.ExpectQuery(regexp.QuoteMeta(`
                    SELECT c.id, c.name
                    FROM city c
                    JOIN vacancy_city vc ON c.id = vc.city_id
                    WHERE vc.vacancy_id = $1
                `)).WithArgs(1).WillReturnRows(rows)
			},
			want:    testCities,
			wantErr: false,
		},
		{
			name:      "Нет городов у вакансии",
			vacancyID: 2,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "name"})

				mock.ExpectQuery(regexp.QuoteMeta(`
                    SELECT c.id, c.name
                    FROM city c
                    JOIN vacancy_city vc ON c.id = vc.city_id
                    WHERE vc.vacancy_id = $1
                `)).WithArgs(2).WillReturnRows(rows)
			},
			want:    []entity.City{},
			wantErr: false,
		},
		{
			name:      "Ошибка выполнения запроса",
			vacancyID: 1,
			mock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`
                    SELECT c.id, c.name
                    FROM city c
                    JOIN vacancy_city vc ON c.id = vc.city_id
                    WHERE vc.vacancy_id = $1
                `)).WithArgs(1).WillReturnError(fmt.Errorf("query error"))
			},
			wantErr:     true,
			expectedErr: "internal server error\nошибка при получении городов резюме: query error",
		},
		{
			name:      "Ошибка сканирования строки",
			vacancyID: 1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Москва").
					AddRow("invalid", "Санкт-Петербург") // Неправильный тип для id

				mock.ExpectQuery(regexp.QuoteMeta(`
                    SELECT c.id, c.name
                    FROM city c
                    JOIN vacancy_city vc ON c.id = vc.city_id
                    WHERE vc.vacancy_id = $1
                `)).WithArgs(1).WillReturnRows(rows)
			},
			wantErr:     true,
			expectedErr: "internal server error\nошибка при сканировании навыка: sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax",
		},
		{
			name:      "Ошибка при итерации по результатам",
			vacancyID: 1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Москва").
					RowError(1, fmt.Errorf("row error"))

				mock.ExpectQuery(regexp.QuoteMeta(`
                    SELECT c.id, c.name
                    FROM city c
                    JOIN vacancy_city vc ON c.id = vc.city_id
                    WHERE vc.vacancy_id = $1
                `)).WithArgs(1).WillReturnRows(rows)
			},
			wantErr:     true,
			expectedErr: "internal server error\nошибка при итерации по навыкам: row error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.GetCityByVacancyID(context.Background(), tt.vacancyID)

			if (err != nil) != tt.wantErr {
				t.Errorf("VacancyRepository.GetCityByVacancyID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err.Error() != tt.expectedErr {
					t.Errorf("VacancyRepository.GetCityByVacancyID() error = %v, expectedErr %v", err.Error(), tt.expectedErr)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VacancyRepository.GetCityByVacancyID() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestVacancyRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

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

func TestVacancyRepository_GetActiveVacanciesByEmployerID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		name           string
		employerID     int
		expectedResult []*entity.Vacancy
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, employerID int)
	}{
		{
			name:       "Успешное получение активных вакансий",
			employerID: 1,
			expectedResult: []*entity.Vacancy{
				{
					ID:                   1,
					Title:                "Backend Developer",
					EmployerID:           1,
					SpecializationID:     1,
					WorkFormat:           "remote",
					Employment:           "full_time",
					Schedule:             "flexible",
					WorkingHours:         40,
					SalaryFrom:           100000,
					SalaryTo:             150000,
					TaxesIncluded:        true,
					Experience:           "1-3 years",
					Description:          "Разработка backend на Go",
					Tasks:                "Разработка API, оптимизация запросов",
					Requirements:         "Опыт работы с Go от 1 года",
					OptionalRequirements: "Знание Docker, Kubernetes",
					City:                 "Москва",
					CreatedAt:            now,
					UpdatedAt:            now,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, employerID int) {
				query := regexp.QuoteMeta(`
                    SELECT id, title, employer_id, specialization_id, work_format, employment, 
                           schedule, working_hours, salary_from, salary_to, taxes_included, experience, 
                           description, tasks, requirements, optional_requirements, city, created_at, updated_at
                    FROM vacancy
                    WHERE employer_id = $1 AND is_active = TRUE
                    ORDER BY updated_at DESC;
                `)
				rows := sqlmock.NewRows([]string{
					"id", "title", "employer_id", "specialization_id", "work_format",
					"employment", "schedule", "working_hours", "salary_from", "salary_to",
					"taxes_included", "experience", "description", "tasks", "requirements",
					"optional_requirements", "city", "created_at", "updated_at",
				}).
					AddRow(
						1, "Backend Developer", 1, 1, "remote",
						"full_time", "flexible", 40, 100000, 150000,
						true, "1-3 years", "Разработка backend на Go",
						"Разработка API, оптимизация запросов", "Опыт работы с Go от 1 года",
						"Знание Docker, Kubernetes", "Москва", now, now,
					)
				mock.ExpectQuery(query).
					WithArgs(employerID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Нет активных вакансий",
			employerID:     2,
			expectedResult: []*entity.Vacancy{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, employerID int) {
				query := regexp.QuoteMeta(`
                    SELECT id, title, employer_id, specialization_id, work_format, employment, 
                           schedule, working_hours, salary_from, salary_to, taxes_included, experience, 
                           description, tasks, requirements, optional_requirements, city, created_at, updated_at
                    FROM vacancy
                    WHERE employer_id = $1 AND is_active = TRUE
                    ORDER BY updated_at DESC;
                `)
				rows := sqlmock.NewRows([]string{
					"id", "title", "employer_id", "specialization_id", "work_format",
					"employment", "schedule", "working_hours", "salary_from", "salary_to",
					"taxes_included", "experience", "description", "tasks", "requirements",
					"optional_requirements", "city", "created_at", "updated_at",
				})
				mock.ExpectQuery(query).
					WithArgs(employerID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка выполнения запроса",
			employerID:     3,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении активных вакансий работодателя: %w", fmt.Errorf("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, employerID int) {
				query := regexp.QuoteMeta(`
                    SELECT id, title, employer_id, specialization_id, work_format, employment, 
                           schedule, working_hours, salary_from, salary_to, taxes_included, experience, 
                           description, tasks, requirements, optional_requirements, city, created_at, updated_at
                    FROM vacancy
                    WHERE employer_id = $1 AND is_active = TRUE
                    ORDER BY updated_at DESC;
                `)
				mock.ExpectQuery(query).
					WithArgs(employerID).
					WillReturnError(fmt.Errorf("database error"))
			},
		},
		{
			name:           "Ошибка сканирования вакансии",
			employerID:     4,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка обработки данных вакансии: %w", fmt.Errorf("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, employerID int) {
				query := regexp.QuoteMeta(`
                    SELECT id, title, employer_id, specialization_id, work_format, employment, 
                           schedule, working_hours, salary_from, salary_to, taxes_included, experience, 
                           description, tasks, requirements, optional_requirements, city, created_at, updated_at
                    FROM vacancy
                    WHERE employer_id = $1 AND is_active = TRUE
                    ORDER BY updated_at DESC;
                `)
				rows := sqlmock.NewRows([]string{
					"id", "title", "employer_id", "specialization_id", "work_format",
					"employment", "schedule", "working_hours", "salary_from", "salary_to",
					"taxes_included", "experience", "description", "tasks", "requirements",
					"optional_requirements", "city", "created_at", "updated_at",
				}).
					AddRow(
						"invalid", "Backend Developer", 1, 1, "remote", // Неправильный тип для id
						"full_time", "flexible", 40, 100000, 150000,
						true, "1-3 years", "Разработка backend на Go",
						"Разработка API, оптимизация запросов", "Опыт работы с Go от 1 года",
						"Знание Docker, Kubernetes", "Москва", now, now,
					)
				mock.ExpectQuery(query).
					WithArgs(employerID).
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

			tc.setupMock(mock, tc.employerID)

			repo := &VacancyRepository{
				DB: db,
			}
			ctx := context.Background()

			result, err := repo.GetActiveVacanciesByEmployerID(ctx, tc.employerID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))

				for i, expected := range tc.expectedResult {
					require.Equal(t, expected.ID, result[i].ID)
					require.Equal(t, expected.Title, result[i].Title)
					// Добавьте проверки для остальных полей по необходимости
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
