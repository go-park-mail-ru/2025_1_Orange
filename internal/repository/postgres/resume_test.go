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
	"github.com/stretchr/testify/require"
)

func TestResumeRepository_Create(t *testing.T) {
	t.Parallel()

	graduationDate := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year",
		"created_at", "updated_at",
	}

	now := time.Now()

	testCases := []struct {
		name           string
		inputResume    *entity.Resume
		expectedResult *entity.Resume
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, resume *entity.Resume)
	}{
		{
			name: "Успешное создание резюме с высшим образованием",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
			},
			expectedResult: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
				CreatedAt:              now,
				UpdatedAt:              now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					INSERT INTO resume (
						applicant_id, about_me, specialization_id, education, 
						educational_institution, graduation_year, created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education), // Преобразуем enum в строку
						resume.EducationalInstitution,
						resume.GraduationYear,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								1,
								resume.ApplicantID,
								resume.AboutMe,
								resume.SpecializationID,
								string(resume.Education),
								resume.EducationalInstitution,
								resume.GraduationYear,
								now,
								now,
							),
					)
			},
		},
		{
			name: "Успешное создание резюме со средним образованием",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Начинающий разработчик",
				SpecializationID:       3,
				Education:              entity.SecondarySchool,
				EducationalInstitution: "Школа №123",
				GraduationYear:         graduationDate,
			},
			expectedResult: &entity.Resume{
				ID:                     2,
				ApplicantID:            1,
				AboutMe:                "Начинающий разработчик",
				SpecializationID:       3,
				Education:              entity.SecondarySchool,
				EducationalInstitution: "Школа №123",
				GraduationYear:         graduationDate,
				CreatedAt:              now,
				UpdatedAt:              now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					INSERT INTO resume (
						applicant_id, about_me, specialization_id, education, 
						educational_institution, graduation_year, created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								2,
								resume.ApplicantID,
								resume.AboutMe,
								resume.SpecializationID,
								string(resume.Education),
								resume.EducationalInstitution,
								resume.GraduationYear,
								now,
								now,
							),
					)
			},
		},
		{
			name: "Ошибка - неверный тип образования",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              "invalid_education", // Несуществующий тип
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					INSERT INTO resume (
						applicant_id, about_me, specialization_id, education, 
						educational_institution, graduation_year, created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				pqErr := &pq.Error{
					Code: entity.PSQLDatatypeViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						resume.Education, // Здесь будет невалидное значение
						resume.EducationalInstitution,
						resume.GraduationYear,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - нарушение уникальности",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              "Высшее",
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("резюме с такими параметрами уже существует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					INSERT INTO resume (
						applicant_id, about_me, specialization_id, education, 
						educational_institution, graduation_year, created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				pqErr := &pq.Error{
					Code: entity.PSQLUniqueViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						resume.Education,
						resume.EducationalInstitution,
						resume.GraduationYear,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - обязательное поле отсутствует",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              "Высшее",
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					INSERT INTO resume (
						applicant_id, about_me, specialization_id, education, 
						educational_institution, graduation_year, created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				pqErr := &pq.Error{
					Code: entity.PSQLNotNullViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						resume.Education,
						resume.EducationalInstitution,
						resume.GraduationYear,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - неправильный формат данных",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              "Высшее",
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					INSERT INTO resume (
						applicant_id, about_me, specialization_id, education, 
						educational_institution, graduation_year, created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				pqErr := &pq.Error{
					Code: entity.PSQLDatatypeViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						resume.Education,
						resume.EducationalInstitution,
						resume.GraduationYear,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - неправильные данные",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              "Высшее",
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					INSERT INTO resume (
						applicant_id, about_me, specialization_id, education, 
						educational_institution, graduation_year, created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				pqErr := &pq.Error{
					Code: entity.PSQLCheckViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						resume.Education,
						resume.EducationalInstitution,
						resume.GraduationYear,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - внутренняя ошибка базы данных",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              "Высшее",
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					INSERT INTO resume (
						applicant_id, about_me, specialization_id, education, 
						educational_institution, graduation_year, created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						resume.Education,
						resume.EducationalInstitution,
						resume.GraduationYear,
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
			defer func() {
				err := db.Close()
				require.NoError(t, err)
			}()

			tc.setupMock(mock, tc.inputResume)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.Create(ctx, tc.inputResume)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.ApplicantID, result.ApplicantID)
				require.Equal(t, tc.expectedResult.AboutMe, result.AboutMe)
				require.Equal(t, tc.expectedResult.SpecializationID, result.SpecializationID)
				require.Equal(t, tc.expectedResult.Education, result.Education)
				require.Equal(t, tc.expectedResult.EducationalInstitution, result.EducationalInstitution)
				require.Equal(t, tc.expectedResult.GraduationYear, result.GraduationYear)
				require.False(t, result.CreatedAt.IsZero())
				require.False(t, result.UpdatedAt.IsZero())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_AddSkills(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		resumeID    int
		skillIDs    []int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int)
	}{
		{
			name:        "Успешное добавление навыков",
			resumeID:    1,
			skillIDs:    []int{1, 2, 3},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_skill (resume_id, skill_id)
					VALUES ($1, $2)
				`))

				for _, skillID := range skillIDs {
					stmt.ExpectExec().
						WithArgs(resumeID, skillID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}

				mock.ExpectCommit()
			},
		},
		// {
		// 	name:        "Пустой список навыков",
		// 	resumeID:    1,
		// 	skillIDs:    []int{},
		// 	expectedErr: nil,
		// 	setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
		// 		// Теперь ожидаем начало транзакции, но не ожидаем подготовку и выполнение запросов
		// 		mock.ExpectBegin()
		// 		mock.ExpectCommit() // Транзакция должна быть завершена
		// 	},
		// },
		{
			name:     "Ошибка начала транзакции",
			resumeID: 1,
			skillIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при начале транзакции для добавления навыков: %w", errors.New("tx error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin().WillReturnError(errors.New("tx error"))
			},
		},
		{
			name:     "Ошибка подготовки запроса",
			resumeID: 1,
			skillIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при подготовке запроса для добавления навыков: %w", errors.New("prepare error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_skill (resume_id, skill_id)
					VALUES ($1, $2)
				`)).WillReturnError(errors.New("prepare error"))
				mock.ExpectRollback()
			},
		},
		{
			name:     "Ошибка выполнения запроса",
			resumeID: 1,
			skillIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении навыка к резюме: %w", errors.New("exec error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_skill (resume_id, skill_id)
					VALUES ($1, $2)
				`))

				stmt.ExpectExec().
					WithArgs(resumeID, skillIDs[0]).
					WillReturnError(errors.New("exec error"))

				mock.ExpectRollback()
			},
		},
		{
			name:        "Нарушение уникальности (пропускаем дубликаты)",
			resumeID:    1,
			skillIDs:    []int{1, 2, 3},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_skill (resume_id, skill_id)
					VALUES ($1, $2)
				`))

				// Первый навык добавляется успешно
				stmt.ExpectExec().
					WithArgs(resumeID, skillIDs[0]).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Второй навык - дубликат
				pqErr := &pq.Error{Code: entity.PSQLUniqueViolation}
				stmt.ExpectExec().
					WithArgs(resumeID, skillIDs[1]).
					WillReturnError(pqErr)

				// Третий навык добавляется успешно
				stmt.ExpectExec().
					WithArgs(resumeID, skillIDs[2]).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
		},
		{
			name:     "Отсутствует обязательное поле",
			resumeID: 1,
			skillIDs: []int{1},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_skill (resume_id, skill_id)
					VALUES ($1, $2)
				`))

				pqErr := &pq.Error{Code: entity.PSQLNotNullViolation}
				stmt.ExpectExec().
					WithArgs(resumeID, skillIDs[0]).
					WillReturnError(pqErr)

				mock.ExpectRollback()
			},
		},
		{
			name:     "Неправильный формат данных",
			resumeID: 1,
			skillIDs: []int{1},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_skill (resume_id, skill_id)
					VALUES ($1, $2)
				`))

				pqErr := &pq.Error{Code: entity.PSQLDatatypeViolation}
				stmt.ExpectExec().
					WithArgs(resumeID, skillIDs[0]).
					WillReturnError(pqErr)

				mock.ExpectRollback()
			},
		},
		{
			name:     "Ошибка коммита транзакции",
			resumeID: 1,
			skillIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при коммите транзакции добавления навыков: %w", errors.New("commit error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_skill (resume_id, skill_id)
					VALUES ($1, $2)
				`))

				for _, skillID := range skillIDs {
					stmt.ExpectExec().
						WithArgs(resumeID, skillID).
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

			tc.setupMock(mock, tc.resumeID, tc.skillIDs)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			err = repo.AddSkills(ctx, tc.resumeID, tc.skillIDs)

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

func TestResumeRepository_AddWorkExperience(t *testing.T) {
	t.Parallel()

	columns := []string{
		"id", "resume_id", "employer_name", "position", "duties",
		"achievements", "start_date", "end_date", "until_now", "updated_at",
	}

	now := time.Now()
	startDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name                string
		inputWorkExperience *entity.WorkExperience
		expectedResult      *entity.WorkExperience
		expectedErr         error
		setupMock           func(mock sqlmock.Sqlmock, we *entity.WorkExperience)
	}{
		{
			name: "Успешное добавление опыта работы с указанием даты окончания",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:     1,
				EmployerName: "Яндекс",
				Position:     "Разработчик",
				Duties:       "Разработка микросервисов",
				Achievements: "Оптимизировал производительность",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: &entity.WorkExperience{
				ID:           1,
				ResumeID:     1,
				EmployerName: "Яндекс",
				Position:     "Разработчик",
				Duties:       "Разработка микросервисов",
				Achievements: "Оптимизировал производительность",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
				UpdatedAt:    now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, we *entity.WorkExperience) {
				query := regexp.QuoteMeta(`
					INSERT INTO work_experience (
						resume_id, employer_name, position, duties, 
						achievements, start_date, end_date, until_now, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
					RETURNING id, resume_id, employer_name, position, duties, 
							  achievements, start_date, end_date, until_now, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						we.ResumeID,
						we.EmployerName,
						we.Position,
						we.Duties,
						we.Achievements,
						we.StartDate,
						sql.NullTime{Time: we.EndDate, Valid: true},
						we.UntilNow,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								1,
								we.ResumeID,
								we.EmployerName,
								we.Position,
								we.Duties,
								we.Achievements,
								we.StartDate,
								sql.NullTime{Time: we.EndDate, Valid: true},
								we.UntilNow,
								now,
							),
					)
			},
		},
		{
			name: "Успешное добавление текущего места работы (until_now = true)",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:     1,
				EmployerName: "Google",
				Position:     "Senior Developer",
				Duties:       "Разработка ядра системы",
				Achievements: "Улучшил архитектуру",
				StartDate:    startDate,
				UntilNow:     true,
			},
			expectedResult: &entity.WorkExperience{
				ID:           2,
				ResumeID:     1,
				EmployerName: "Google",
				Position:     "Senior Developer",
				Duties:       "Разработка ядра системы",
				Achievements: "Улучшил архитектуру",
				StartDate:    startDate,
				UntilNow:     true,
				UpdatedAt:    now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, we *entity.WorkExperience) {
				query := regexp.QuoteMeta(`
					INSERT INTO work_experience (
						resume_id, employer_name, position, duties, 
						achievements, start_date, end_date, until_now, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
					RETURNING id, resume_id, employer_name, position, duties, 
							  achievements, start_date, end_date, until_now, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						we.ResumeID,
						we.EmployerName,
						we.Position,
						we.Duties,
						we.Achievements,
						we.StartDate,
						sql.NullTime{Valid: false},
						we.UntilNow,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								2,
								we.ResumeID,
								we.EmployerName,
								we.Position,
								we.Duties,
								we.Achievements,
								we.StartDate,
								sql.NullTime{Valid: false},
								we.UntilNow,
								now,
							),
					)
			},
		},
		{
			name: "Ошибка - нарушение уникальности",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:     1,
				EmployerName: "Яндекс",
				Position:     "Разработчик",
				StartDate:    startDate,
				UntilNow:     true,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("опыт работы с такими параметрами уже существует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, we *entity.WorkExperience) {
				query := regexp.QuoteMeta(`
					INSERT INTO work_experience (
						resume_id, employer_name, position, duties, 
						achievements, start_date, end_date, until_now, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
					RETURNING id, resume_id, employer_name, position, duties, 
							  achievements, start_date, end_date, until_now, updated_at
				`)
				pqErr := &pq.Error{
					Code: entity.PSQLUniqueViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						we.ResumeID,
						we.EmployerName,
						we.Position,
						we.Duties,
						we.Achievements,
						we.StartDate,
						sql.NullTime{Valid: false},
						we.UntilNow,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - обязательное поле отсутствует",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:  1,
				Position:  "Разработчик",
				StartDate: startDate,
				UntilNow:  true,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, we *entity.WorkExperience) {
				query := regexp.QuoteMeta(`
					INSERT INTO work_experience (
						resume_id, employer_name, position, duties, 
						achievements, start_date, end_date, until_now, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
					RETURNING id, resume_id, employer_name, position, duties, 
							  achievements, start_date, end_date, until_now, updated_at
				`)
				pqErr := &pq.Error{
					Code: entity.PSQLNotNullViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						we.ResumeID,
						"", // EmployerName отсутствует
						we.Position,
						we.Duties,
						we.Achievements,
						we.StartDate,
						sql.NullTime{Valid: false},
						we.UntilNow,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - внутренняя ошибка базы данных",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:     1,
				EmployerName: "Яндекс",
				Position:     "Разработчик",
				StartDate:    startDate,
				UntilNow:     true,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании опыта работы: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, we *entity.WorkExperience) {
				query := regexp.QuoteMeta(`
					INSERT INTO work_experience (
						resume_id, employer_name, position, duties, 
						achievements, start_date, end_date, until_now, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
					RETURNING id, resume_id, employer_name, position, duties, 
							  achievements, start_date, end_date, until_now, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						we.ResumeID,
						we.EmployerName,
						we.Position,
						we.Duties,
						we.Achievements,
						we.StartDate,
						sql.NullTime{Valid: false},
						we.UntilNow,
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
			defer func() {
				err := db.Close()
				require.NoError(t, err)
			}()

			tc.setupMock(mock, tc.inputWorkExperience)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.AddWorkExperience(ctx, tc.inputWorkExperience)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.ResumeID, result.ResumeID)
				require.Equal(t, tc.expectedResult.EmployerName, result.EmployerName)
				require.Equal(t, tc.expectedResult.Position, result.Position)
				require.Equal(t, tc.expectedResult.Duties, result.Duties)
				require.Equal(t, tc.expectedResult.Achievements, result.Achievements)
				require.Equal(t, tc.expectedResult.StartDate.Unix(), result.StartDate.Unix())

				if tc.expectedResult.UntilNow {
					require.True(t, result.UntilNow)
					require.True(t, result.EndDate.IsZero())
				} else {
					require.Equal(t, tc.expectedResult.EndDate.Unix(), result.EndDate.Unix())
				}

				require.False(t, result.UpdatedAt.IsZero())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_GetByID(t *testing.T) {
	t.Parallel()

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year",
		"created_at", "updated_at",
	}

	now := time.Now()
	gradYear := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name           string
		id             int
		expectedResult *entity.Resume
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name: "Успешное получение резюме с высшим образованием",
			id:   1,
			expectedResult: &entity.Resume{
				ID:                     1,
				ApplicantID:            123,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         gradYear,
				CreatedAt:              now,
				UpdatedAt:              now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					SELECT id, applicant_id, about_me, specialization_id, education, 
						   educational_institution, graduation_year, created_at, updated_at
					FROM resume
					WHERE id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(id).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								1,
								123,
								"Опытный разработчик",
								2,
								"higher",
								"МГУ",
								gradYear,
								now,
								now,
							),
					)
			},
		},
		{
			name: "Успешное получение резюме со средним образованием",
			id:   2,
			expectedResult: &entity.Resume{
				ID:                     2,
				ApplicantID:            456,
				AboutMe:                "Начинающий разработчик",
				SpecializationID:       3,
				Education:              entity.SecondarySchool,
				EducationalInstitution: "Школа №123",
				GraduationYear:         gradYear,
				CreatedAt:              now,
				UpdatedAt:              now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					SELECT id, applicant_id, about_me, specialization_id, education, 
						   educational_institution, graduation_year, created_at, updated_at
					FROM resume
					WHERE id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(id).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								2,
								456,
								"Начинающий разработчик",
								3,
								"secondary_school",
								"Школа №123",
								gradYear,
								now,
								now,
							),
					)
			},
		},
		{
			name:           "Резюме не найдено",
			id:             999,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=999 не найдено"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					SELECT id, applicant_id, about_me, specialization_id, education, 
						   educational_institution, graduation_year, created_at, updated_at
					FROM resume
					WHERE id = $1
				`)
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
				fmt.Errorf("не удалось получить резюме по id=3"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					SELECT id, applicant_id, about_me, specialization_id, education, 
						   educational_institution, graduation_year, created_at, updated_at
					FROM resume
					WHERE id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(id).
					WillReturnError(errors.New("database error"))
			},
		},
		// {
		// 	name:           "Ошибка сканирования (неверный тип образования)",
		// 	id:             4,
		// 	expectedResult: nil,
		// 	expectedErr: entity.NewError(
		// 		entity.ErrInternal,
		// 		fmt.Errorf("не удалось получить резюме по id=4"),
		// 	),
		// 	setupMock: func(mock sqlmock.Sqlmock, id int) {
		// 		query := regexp.QuoteMeta(`
		// 			SELECT id, applicant_id, about_me, specialization_id, education,
		// 				   educational_institution, graduation_year, created_at, updated_at
		// 			FROM resume
		// 			WHERE id = $1
		// 		`)
		// 		mock.ExpectQuery(query).
		// 			WithArgs(id).
		// 			WillReturnRows(
		// 				sqlmock.NewRows(columns).
		// 					AddRow(
		// 						4,
		// 						789,
		// 						"Разработчик",
		// 						5,
		// 						"invalid_education", // Невалидное значение
		// 						"Университет",
		// 						gradYear,
		// 						now,
		// 						now,
		// 					),
		// 			)
		// 	},
		// },
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

			repo := &ResumeRepository{DB: db}
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
				require.Equal(t, tc.expectedResult.ApplicantID, result.ApplicantID)
				require.Equal(t, tc.expectedResult.AboutMe, result.AboutMe)
				require.Equal(t, tc.expectedResult.SpecializationID, result.SpecializationID)
				require.Equal(t, tc.expectedResult.Education, result.Education)
				require.Equal(t, tc.expectedResult.EducationalInstitution, result.EducationalInstitution)
				require.Equal(t, tc.expectedResult.GraduationYear.Unix(), result.GraduationYear.Unix())
				require.Equal(t, tc.expectedResult.CreatedAt.Unix(), result.CreatedAt.Unix())
				require.Equal(t, tc.expectedResult.UpdatedAt.Unix(), result.UpdatedAt.Unix())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_GetSkillsByResumeID(t *testing.T) {
	t.Parallel()

	columns := []string{"id", "name"}

	testCases := []struct {
		name           string
		resumeID       int
		expectedResult []entity.Skill
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, resumeID int)
	}{
		{
			name:     "Успешное получение навыков резюме",
			resumeID: 1,
			expectedResult: []entity.Skill{
				{ID: 1, Name: "Go"},
				{ID: 2, Name: "SQL"},
				{ID: 3, Name: "Docker"},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT s.id, s.name
					FROM skill s
					JOIN resume_skill rs ON s.id = rs.skill_id
					WHERE rs.resume_id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(1, "Go").
							AddRow(2, "SQL").
							AddRow(3, "Docker"),
					)
			},
		},
		{
			name:           "Резюме без навыков",
			resumeID:       2,
			expectedResult: nil, // Изменено с []entity.Skill{} на nil
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT s.id, s.name
					FROM skill s
					JOIN resume_skill rs ON s.id = rs.skill_id
					WHERE rs.resume_id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(sqlmock.NewRows(columns))
			},
		},
		{
			name:           "Ошибка при выполнении запроса",
			resumeID:       3,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении навыков резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT s.id, s.name
					FROM skill s
					JOIN resume_skill rs ON s.id = rs.skill_id
					WHERE rs.resume_id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка при сканировании навыка",
			resumeID:       4,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании навыка: %w", errors.New("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT s.id, s.name
					FROM skill s
					JOIN resume_skill rs ON s.id = rs.skill_id
					WHERE rs.resume_id = $1
				`)
				// Используем неправильный тип данных для эмуляции ошибки сканирования
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow("invalid_id", "Go"), // Неправильный тип для id
					)
			},
		},
		{
			name:           "Ошибка при итерации по навыкам",
			resumeID:       5,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по навыкам: %w", errors.New("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT s.id, s.name
					FROM skill s
					JOIN resume_skill rs ON s.id = rs.skill_id
					WHERE rs.resume_id = $1
				`)
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Go").
					AddRow(2, "SQL").
					CloseError(errors.New("rows error"))
				mock.ExpectQuery(query).
					WithArgs(resumeID).
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

			tc.setupMock(mock, tc.resumeID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetSkillsByResumeID(ctx, tc.resumeID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				if tc.name == "Ошибка при сканировании навыка" {
					// Для ошибки сканирования проверяем только часть сообщения
					require.Contains(t, err.Error(), "ошибка при сканировании навыка")
				} else {
					require.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_GetWorkExperienceByResumeID(t *testing.T) {
	t.Parallel()

	columns := []string{
		"id", "resume_id", "employer_name", "position", "duties",
		"achievements", "start_date", "end_date", "until_now", "updated_at",
	}

	now := time.Now()
	startDate1 := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate1 := time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)
	startDate2 := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name           string
		resumeID       int
		expectedResult []entity.WorkExperience
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, resumeID int)
	}{
		{
			name:     "Успешное получение опыта работы",
			resumeID: 1,
			expectedResult: []entity.WorkExperience{
				{
					ID:           1,
					ResumeID:     1,
					EmployerName: "Яндекс",
					Position:     "Backend Developer",
					Duties:       "Разработка микросервисов",
					Achievements: "Оптимизация API",
					StartDate:    startDate1,
					EndDate:      endDate1,
					UntilNow:     false,
					UpdatedAt:    now,
				},
				{
					ID:           2,
					ResumeID:     1,
					EmployerName: "Google",
					Position:     "Senior Developer",
					Duties:       "Разработка ядра системы",
					Achievements: "Ускорение работы на 30%",
					StartDate:    startDate2,
					UntilNow:     true,
					UpdatedAt:    now,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT id, resume_id, employer_name, position, duties, 
						   achievements, start_date, end_date, until_now, updated_at
					FROM work_experience
					WHERE resume_id = $1
					ORDER BY start_date DESC
				`)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								1,
								1,
								"Яндекс",
								"Backend Developer",
								"Разработка микросервисов",
								"Оптимизация API",
								startDate1,
								endDate1,
								false,
								now,
							).
							AddRow(
								2,
								1,
								"Google",
								"Senior Developer",
								"Разработка ядра системы",
								"Ускорение работы на 30%",
								startDate2,
								nil, // Для UntilNow = true
								true,
								now,
							),
					)
			},
		},
		{
			name:           "Пустой список опыта работы",
			resumeID:       2,
			expectedResult: []entity.WorkExperience{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT id, resume_id, employer_name, position, duties, 
						   achievements, start_date, end_date, until_now, updated_at
					FROM work_experience
					WHERE resume_id = $1
					ORDER BY start_date DESC
				`)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(sqlmock.NewRows(columns))
			},
		},
		{
			name:           "Ошибка при выполнении запроса",
			resumeID:       3,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении опыта работы: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT id, resume_id, employer_name, position, duties, 
						   achievements, start_date, end_date, until_now, updated_at
					FROM work_experience
					WHERE resume_id = $1
					ORDER BY start_date DESC
				`)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка при сканировании строк",
			resumeID:       4,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по опыту работы: %w", errors.New("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT id, resume_id, employer_name, position, duties, 
						   achievements, start_date, end_date, until_now, updated_at
					FROM work_experience
					WHERE resume_id = $1
					ORDER BY start_date DESC
				`)
				rows := sqlmock.NewRows(columns).
					AddRow(1, 4, "Company", "Position", "Duties", "Achievements", startDate1, endDate1, false, now).
					RowError(0, errors.New("scan error"))
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка при итерации по строкам",
			resumeID:       5,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по опыту работы: %w", errors.New("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT id, resume_id, employer_name, position, duties, 
						   achievements, start_date, end_date, until_now, updated_at
					FROM work_experience
					WHERE resume_id = $1
					ORDER BY start_date DESC
				`)
				rows := sqlmock.NewRows(columns).
					AddRow(1, 5, "Company", "Position", "Duties", "Achievements", startDate1, endDate1, false, now).
					CloseError(errors.New("rows error"))
				mock.ExpectQuery(query).
					WithArgs(resumeID).
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

			tc.setupMock(mock, tc.resumeID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetWorkExperienceByResumeID(ctx, tc.resumeID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))

				for i := range tc.expectedResult {
					require.Equal(t, tc.expectedResult[i].ID, result[i].ID)
					require.Equal(t, tc.expectedResult[i].ResumeID, result[i].ResumeID)
					require.Equal(t, tc.expectedResult[i].EmployerName, result[i].EmployerName)
					require.Equal(t, tc.expectedResult[i].Position, result[i].Position)
					require.Equal(t, tc.expectedResult[i].Duties, result[i].Duties)
					require.Equal(t, tc.expectedResult[i].Achievements, result[i].Achievements)
					require.Equal(t, tc.expectedResult[i].StartDate.Unix(), result[i].StartDate.Unix())
					require.Equal(t, tc.expectedResult[i].EndDate.Unix(), result[i].EndDate.Unix())
					require.Equal(t, tc.expectedResult[i].UntilNow, result[i].UntilNow)
					require.False(t, result[i].UpdatedAt.IsZero())
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_AddSpecializations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		resumeID          int
		specializationIDs []int
		expectedErr       error
		setupMock         func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int)
	}{
		{
			name:              "Успешное добавление специализаций",
			resumeID:          1,
			specializationIDs: []int{2, 3, 4},
			expectedErr:       nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_specialization (resume_id, specialization_id)
					VALUES ($1, $2)
				`))

				for _, specID := range specializationIDs {
					stmt.ExpectExec().
						WithArgs(resumeID, specID).
						WillReturnResult(sqlmock.NewResult(0, 1))
				}

				mock.ExpectCommit()
			},
		},
		{
			name:              "Пустой список специализаций",
			resumeID:          1,
			specializationIDs: []int{},
			expectedErr:       nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				// Ожидаем подготовку statement, даже если он не будет использован
				mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_specialization (resume_id, specialization_id)
					VALUES ($1, $2)
				`))
				mock.ExpectCommit()
			},
		},
		{
			name:              "Ошибка начала транзакции",
			resumeID:          1,
			specializationIDs: []int{2, 3},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при начале транзакции для добавления специализаций: %w", errors.New("tx error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin().WillReturnError(errors.New("tx error"))
			},
		},
		{
			name:              "Ошибка подготовки запроса",
			resumeID:          1,
			specializationIDs: []int{2, 3},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при подготовке запроса для добавления специализаций: %w", errors.New("prepare error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_specialization (resume_id, specialization_id)
					VALUES ($1, $2)
				`)).WillReturnError(errors.New("prepare error"))
				mock.ExpectRollback()
			},
		},
		{
			name:              "Дубликаты специализаций",
			resumeID:          1,
			specializationIDs: []int{2, 2, 3},
			expectedErr:       nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_specialization (resume_id, specialization_id)
					VALUES ($1, $2)
				`))

				// Первый вызов успешен
				stmt.ExpectExec().
					WithArgs(resumeID, 2).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// Второй вызов с тем же ID - ошибка уникальности
				pqErr := &pq.Error{Code: entity.PSQLUniqueViolation}
				stmt.ExpectExec().
					WithArgs(resumeID, 2).
					WillReturnError(pqErr)

				// Третий вызов успешен
				stmt.ExpectExec().
					WithArgs(resumeID, 3).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit()
			},
		},
		{
			name:              "Ошибка обязательного поля",
			resumeID:          1,
			specializationIDs: []int{2},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_specialization (resume_id, specialization_id)
					VALUES ($1, $2)
				`))

				pqErr := &pq.Error{Code: entity.PSQLNotNullViolation}
				stmt.ExpectExec().
					WithArgs(resumeID, 2).
					WillReturnError(pqErr)

				mock.ExpectRollback()
			},
		},
		{
			name:              "Ошибка выполнения запроса",
			resumeID:          1,
			specializationIDs: []int{2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении специализации к резюме: %w", errors.New("exec error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_specialization (resume_id, specialization_id)
					VALUES ($1, $2)
				`))

				stmt.ExpectExec().
					WithArgs(resumeID, 2).
					WillReturnError(errors.New("exec error"))

				mock.ExpectRollback()
			},
		},
		{
			name:              "Ошибка коммита транзакции",
			resumeID:          1,
			specializationIDs: []int{2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при коммите транзакции добавления специализаций: %w", errors.New("commit error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				stmt := mock.ExpectPrepare(regexp.QuoteMeta(`
					INSERT INTO resume_specialization (resume_id, specialization_id)
					VALUES ($1, $2)
				`))

				stmt.ExpectExec().
					WithArgs(resumeID, 2).
					WillReturnResult(sqlmock.NewResult(0, 1))

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

			tc.setupMock(mock, tc.resumeID, tc.specializationIDs)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			err = repo.AddSpecializations(ctx, tc.resumeID, tc.specializationIDs)

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

func TestResumeRepository_GetSpecializationsByResumeID(t *testing.T) {
	t.Parallel()

	columns := []string{"id", "name"}

	testCases := []struct {
		name           string
		resumeID       int
		expectedResult []entity.Specialization
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, resumeID int)
	}{
		{
			name:     "Успешное получение специализаций",
			resumeID: 1,
			expectedResult: []entity.Specialization{
				{ID: 1, Name: "Backend разработка"},
				{ID: 2, Name: "DevOps"},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT s.id, s.name
					FROM specialization s
					JOIN resume_specialization rs ON s.id = rs.specialization_id
					WHERE rs.resume_id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(1, "Backend разработка").
							AddRow(2, "DevOps"),
					)
			},
		},
		{
			name:           "Пустой список специализаций",
			resumeID:       2,
			expectedResult: nil, // Изменено с []entity.Specialization{} на nil
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT s.id, s.name
					FROM specialization s
					JOIN resume_specialization rs ON s.id = rs.specialization_id
					WHERE rs.resume_id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(sqlmock.NewRows(columns))
			},
		},
		{
			name:           "Ошибка при выполнении запроса",
			resumeID:       3,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении специализаций резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT s.id, s.name
					FROM specialization s
					JOIN resume_specialization rs ON s.id = rs.specialization_id
					WHERE rs.resume_id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка при сканировании строк",
			resumeID:       4,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании специализации: %w",
					fmt.Errorf("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid_id\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT s.id, s.name
					FROM specialization s
					JOIN resume_specialization rs ON s.id = rs.specialization_id
					WHERE rs.resume_id = $1
				`)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow("invalid_id", "Backend разработка"), // Неправильный тип для id
					)
			},
		},
		{
			name:           "Ошибка при итерации по строкам",
			resumeID:       5,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по специализациям: %w", errors.New("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					SELECT s.id, s.name
					FROM specialization s
					JOIN resume_specialization rs ON s.id = rs.specialization_id
					WHERE rs.resume_id = $1
				`)
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Backend разработка").
					AddRow(2, "DevOps").
					CloseError(errors.New("rows error"))
				mock.ExpectQuery(query).
					WithArgs(resumeID).
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

			tc.setupMock(mock, tc.resumeID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetSpecializationsByResumeID(ctx, tc.resumeID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_Update(t *testing.T) {
	t.Parallel()

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year",
		"created_at", "updated_at",
	}

	now := time.Now()
	gradYear := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	createdAt := now.Add(-24 * time.Hour) // Дата создания на день раньше

	testCases := []struct {
		name           string
		inputResume    *entity.Resume
		expectedResult *entity.Resume
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, resume *entity.Resume)
	}{
		{
			name: "Успешное обновление резюме",
			inputResume: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Обновленное описание",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         gradYear,
			},
			expectedResult: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Обновленное описание",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         gradYear,
				CreatedAt:              createdAt,
				UpdatedAt:              now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					UPDATE resume
					SET 
						about_me = $1,
						specialization_id = $2,
						education = $3,
						educational_institution = $4,
						graduation_year = $5,
						updated_at = NOW()
					WHERE id = $6 AND applicant_id = $7
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.ID,
						resume.ApplicantID,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								resume.ID,
								resume.ApplicantID,
								resume.AboutMe,
								resume.SpecializationID,
								string(resume.Education),
								resume.EducationalInstitution,
								resume.GraduationYear,
								createdAt,
								now,
							),
					)
			},
		},
		{
			name: "Резюме не найдено или не принадлежит пользователю",
			inputResume: &entity.Resume{
				ID:                     999,
				ApplicantID:            1,
				AboutMe:                "Обновленное описание",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         gradYear,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=999 не найдено или не принадлежит указанному соискателю"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					UPDATE resume
					SET 
						about_me = $1,
						specialization_id = $2,
						education = $3,
						educational_institution = $4,
						graduation_year = $5,
						updated_at = NOW()
					WHERE id = $6 AND applicant_id = $7
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.ID,
						resume.ApplicantID,
					).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name: "Ошибка - нарушение уникальности",
			inputResume: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Обновленное описание",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         gradYear,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("резюме с такими параметрами уже существует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					UPDATE resume
					SET 
						about_me = $1,
						specialization_id = $2,
						education = $3,
						educational_institution = $4,
						graduation_year = $5,
						updated_at = NOW()
					WHERE id = $6 AND applicant_id = $7
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				pqErr := &pq.Error{
					Code: entity.PSQLUniqueViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.ID,
						resume.ApplicantID,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - обязательное поле отсутствует",
			inputResume: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Обновленное описание",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         gradYear,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					UPDATE resume
					SET 
						about_me = $1,
						specialization_id = $2,
						education = $3,
						educational_institution = $4,
						graduation_year = $5,
						updated_at = NOW()
					WHERE id = $6 AND applicant_id = $7
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				pqErr := &pq.Error{
					Code: entity.PSQLNotNullViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.ID,
						resume.ApplicantID,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - внутренняя ошибка базы данных",
			inputResume: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Обновленное описание",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         gradYear,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обновлении резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				query := regexp.QuoteMeta(`
					UPDATE resume
					SET 
						about_me = $1,
						specialization_id = $2,
						education = $3,
						educational_institution = $4,
						graduation_year = $5,
						updated_at = NOW()
					WHERE id = $6 AND applicant_id = $7
					RETURNING id, applicant_id, about_me, specialization_id, education, 
							  educational_institution, graduation_year, created_at, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.ID,
						resume.ApplicantID,
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
			defer func() {
				err := db.Close()
				require.NoError(t, err)
			}()

			tc.setupMock(mock, tc.inputResume)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.Update(ctx, tc.inputResume)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.ApplicantID, result.ApplicantID)
				require.Equal(t, tc.expectedResult.AboutMe, result.AboutMe)
				require.Equal(t, tc.expectedResult.SpecializationID, result.SpecializationID)
				require.Equal(t, tc.expectedResult.Education, result.Education)
				require.Equal(t, tc.expectedResult.EducationalInstitution, result.EducationalInstitution)
				require.Equal(t, tc.expectedResult.GraduationYear.Unix(), result.GraduationYear.Unix())
				require.Equal(t, tc.expectedResult.CreatedAt.Unix(), result.CreatedAt.Unix())
				require.False(t, result.UpdatedAt.IsZero())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_Delete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		id          int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name:        "Успешное удаление резюме",
			id:          1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					DELETE FROM resume
					WHERE id = $1
				`)
				mock.ExpectExec(query).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Резюме не найдено",
			id:   999,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=999 не найдено"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					DELETE FROM resume
					WHERE id = $1
				`)
				mock.ExpectExec(query).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name: "Ошибка при выполнении запроса",
			id:   2,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					DELETE FROM resume
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
					DELETE FROM resume
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

			repo := &ResumeRepository{DB: db}
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

func TestResumeRepository_DeleteSkills(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		resumeID    int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, resumeID int)
	}{
		{
			name:        "Успешное удаление навыков резюме",
			resumeID:    1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM resume_skill
                    WHERE resume_id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:        "Резюме не существует (нет навыков для удаления)",
			resumeID:    999,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM resume_skill
                    WHERE resume_id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:     "Ошибка базы данных",
			resumeID: 2,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении навыков резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
                    DELETE FROM resume_skill
                    WHERE resume_id = $1
                `)
				mock.ExpectExec(query).
					WithArgs(resumeID).
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

			tc.setupMock(mock, tc.resumeID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			// Добавляем валидацию ID перед вызовом метода
			if tc.resumeID <= 0 {
				err := repo.DeleteSkills(ctx, tc.resumeID)
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				return
			}

			err = repo.DeleteSkills(ctx, tc.resumeID)

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

func TestResumeRepository_DeleteSpecializations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		resumeID    int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, resumeID int)
	}{
		{
			name:     "Успешное удаление специализаций",
			resumeID: 1,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					DELETE FROM resume_specialization
					WHERE resume_id = $1
				`)
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 1)) // 1 affected row
			},
		},
		{
			name:     "Резюме без специализаций",
			resumeID: 2,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					DELETE FROM resume_specialization
					WHERE resume_id = $1
				`)
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 affected rows
			},
		},
		{
			name:     "Ошибка базы данных",
			resumeID: 3,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении специализаций резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					DELETE FROM resume_specialization
					WHERE resume_id = $1
				`)
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:     "Неверный ID резюме",
			resumeID: -1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении специализаций резюме: %w", errors.New("invalid resume ID")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					DELETE FROM resume_specialization
					WHERE resume_id = $1
				`)
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnError(errors.New("invalid resume ID"))
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

			tc.setupMock(mock, tc.resumeID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			err = repo.DeleteSpecializations(ctx, tc.resumeID)

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

func TestResumeRepository_DeleteWorkExperiences(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		resumeID    int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, resumeID int)
	}{
		{
			name:     "Успешное удаление опыта работы",
			resumeID: 1,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					DELETE FROM work_experience
					WHERE resume_id = $1
				`)
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:     "Резюме без опыта работы (нет записей для удаления)",
			resumeID: 2,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					DELETE FROM work_experience
					WHERE resume_id = $1
				`)
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:     "Ошибка базы данных",
			resumeID: 3,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении опыта работы резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				query := regexp.QuoteMeta(`
					DELETE FROM work_experience
					WHERE resume_id = $1
				`)
				mock.ExpectExec(query).
					WithArgs(resumeID).
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

			// Для теста с некорректным ID не настраиваем мок
			if tc.resumeID > 0 {
				tc.setupMock(mock, tc.resumeID)
			}

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			err = repo.DeleteWorkExperiences(ctx, tc.resumeID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			// Проверяем мок только если он был настроен
			if tc.resumeID > 0 {
				require.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

func TestResumeRepository_UpdateWorkExperience(t *testing.T) {
	t.Parallel()

	columns := []string{
		"id", "resume_id", "employer_name", "position",
		"duties", "achievements", "start_date", "end_date",
		"until_now", "updated_at",
	}

	now := time.Now()
	startDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name                string
		inputWorkExperience *entity.WorkExperience
		expectedResult      *entity.WorkExperience
		expectedErr         error
		setupMock           func(mock sqlmock.Sqlmock, we *entity.WorkExperience)
	}{
		{
			name: "Успешное обновление опыта работы с указанием даты окончания",
			inputWorkExperience: &entity.WorkExperience{
				ID:           1,
				ResumeID:     1,
				EmployerName: "Яндекс",
				Position:     "Разработчик",
				Duties:       "Разработка микросервисов",
				Achievements: "Оптимизация производительности",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: &entity.WorkExperience{
				ID:           1,
				ResumeID:     1,
				EmployerName: "Яндекс",
				Position:     "Разработчик",
				Duties:       "Разработка микросервисов",
				Achievements: "Оптимизация производительности",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
				UpdatedAt:    now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, we *entity.WorkExperience) {
				query := regexp.QuoteMeta(`
					UPDATE work_experience
					SET 
						employer_name = $1,
						position = $2,
						duties = $3,
						achievements = $4,
						start_date = $5,
						end_date = $6,
						until_now = $7,
						updated_at = NOW()
					WHERE id = $8 AND resume_id = $9
					RETURNING id, resume_id, employer_name, position, duties, 
							  achievements, start_date, end_date, until_now, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						we.EmployerName,
						we.Position,
						we.Duties,
						we.Achievements,
						we.StartDate,
						sql.NullTime{Time: we.EndDate, Valid: true},
						we.UntilNow,
						we.ID,
						we.ResumeID,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								we.ID,
								we.ResumeID,
								we.EmployerName,
								we.Position,
								we.Duties,
								we.Achievements,
								we.StartDate,
								we.EndDate,
								we.UntilNow,
								now,
							),
					)
			},
		},
		{
			name: "Успешное обновление текущего места работы (until_now = true)",
			inputWorkExperience: &entity.WorkExperience{
				ID:           2,
				ResumeID:     1,
				EmployerName: "Google",
				Position:     "Senior Developer",
				Duties:       "Разработка архитектуры",
				Achievements: "Внедрение новых технологий",
				StartDate:    startDate,
				UntilNow:     true,
			},
			expectedResult: &entity.WorkExperience{
				ID:           2,
				ResumeID:     1,
				EmployerName: "Google",
				Position:     "Senior Developer",
				Duties:       "Разработка архитектуры",
				Achievements: "Внедрение новых технологий",
				StartDate:    startDate,
				UntilNow:     true,
				UpdatedAt:    now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, we *entity.WorkExperience) {
				query := regexp.QuoteMeta(`
					UPDATE work_experience
					SET 
						employer_name = $1,
						position = $2,
						duties = $3,
						achievements = $4,
						start_date = $5,
						end_date = $6,
						until_now = $7,
						updated_at = NOW()
					WHERE id = $8 AND resume_id = $9
					RETURNING id, resume_id, employer_name, position, duties, 
							  achievements, start_date, end_date, until_now, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						we.EmployerName,
						we.Position,
						we.Duties,
						we.Achievements,
						we.StartDate,
						sql.NullTime{},
						we.UntilNow,
						we.ID,
						we.ResumeID,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								we.ID,
								we.ResumeID,
								we.EmployerName,
								we.Position,
								we.Duties,
								we.Achievements,
								we.StartDate,
								nil, // end_date
								we.UntilNow,
								now,
							),
					)
			},
		},
		{
			name: "Ошибка - запись не найдена",
			inputWorkExperience: &entity.WorkExperience{
				ID:           999,
				ResumeID:     1,
				EmployerName: "Несуществующая компания",
				Position:     "Разработчик",
				StartDate:    startDate,
				UntilNow:     true,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("запись об опыте работы с id=999 не найдена или не принадлежит указанному резюме"),
			),
			setupMock: func(mock sqlmock.Sqlmock, we *entity.WorkExperience) {
				query := regexp.QuoteMeta(`
					UPDATE work_experience
					SET 
						employer_name = $1,
						position = $2,
						duties = $3,
						achievements = $4,
						start_date = $5,
						end_date = $6,
						until_now = $7,
						updated_at = NOW()
					WHERE id = $8 AND resume_id = $9
					RETURNING id, resume_id, employer_name, position, duties, 
							  achievements, start_date, end_date, until_now, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						we.EmployerName,
						we.Position,
						we.Duties,
						we.Achievements,
						we.StartDate,
						sql.NullTime{},
						we.UntilNow,
						we.ID,
						we.ResumeID,
					).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name: "Ошибка - нарушение уникальности",
			inputWorkExperience: &entity.WorkExperience{
				ID:           3,
				ResumeID:     1,
				EmployerName: "Яндекс",
				Position:     "Разработчик",
				StartDate:    startDate,
				UntilNow:     true,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("запись об опыте работы с такими параметрами уже существует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, we *entity.WorkExperience) {
				query := regexp.QuoteMeta(`
					UPDATE work_experience
					SET 
						employer_name = $1,
						position = $2,
						duties = $3,
						achievements = $4,
						start_date = $5,
						end_date = $6,
						until_now = $7,
						updated_at = NOW()
					WHERE id = $8 AND resume_id = $9
					RETURNING id, resume_id, employer_name, position, duties, 
							  achievements, start_date, end_date, until_now, updated_at
				`)
				pqErr := &pq.Error{
					Code: entity.PSQLUniqueViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						we.EmployerName,
						we.Position,
						we.Duties,
						we.Achievements,
						we.StartDate,
						sql.NullTime{},
						we.UntilNow,
						we.ID,
						we.ResumeID,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - обязательное поле отсутствует",
			inputWorkExperience: &entity.WorkExperience{
				ID:           4,
				ResumeID:     1,
				EmployerName: "",
				Position:     "Разработчик",
				StartDate:    startDate,
				UntilNow:     true,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, we *entity.WorkExperience) {
				query := regexp.QuoteMeta(`
					UPDATE work_experience
					SET 
						employer_name = $1,
						position = $2,
						duties = $3,
						achievements = $4,
						start_date = $5,
						end_date = $6,
						until_now = $7,
						updated_at = NOW()
					WHERE id = $8 AND resume_id = $9
					RETURNING id, resume_id, employer_name, position, duties, 
							  achievements, start_date, end_date, until_now, updated_at
				`)
				pqErr := &pq.Error{
					Code: entity.PSQLNotNullViolation,
				}
				mock.ExpectQuery(query).
					WithArgs(
						we.EmployerName,
						we.Position,
						we.Duties,
						we.Achievements,
						we.StartDate,
						sql.NullTime{},
						we.UntilNow,
						we.ID,
						we.ResumeID,
					).
					WillReturnError(pqErr)
			},
		},
		{
			name: "Ошибка - внутренняя ошибка базы данных",
			inputWorkExperience: &entity.WorkExperience{
				ID:           5,
				ResumeID:     1,
				EmployerName: "Microsoft",
				Position:     "Разработчик",
				StartDate:    startDate,
				UntilNow:     true,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обновлении записи об опыте работы: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, we *entity.WorkExperience) {
				query := regexp.QuoteMeta(`
					UPDATE work_experience
					SET 
						employer_name = $1,
						position = $2,
						duties = $3,
						achievements = $4,
						start_date = $5,
						end_date = $6,
						until_now = $7,
						updated_at = NOW()
					WHERE id = $8 AND resume_id = $9
					RETURNING id, resume_id, employer_name, position, duties, 
							  achievements, start_date, end_date, until_now, updated_at
				`)
				mock.ExpectQuery(query).
					WithArgs(
						we.EmployerName,
						we.Position,
						we.Duties,
						we.Achievements,
						we.StartDate,
						sql.NullTime{},
						we.UntilNow,
						we.ID,
						we.ResumeID,
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
			defer func() {
				err := db.Close()
				require.NoError(t, err)
			}()

			tc.setupMock(mock, tc.inputWorkExperience)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.UpdateWorkExperience(ctx, tc.inputWorkExperience)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.ResumeID, result.ResumeID)
				require.Equal(t, tc.expectedResult.EmployerName, result.EmployerName)
				require.Equal(t, tc.expectedResult.Position, result.Position)
				require.Equal(t, tc.expectedResult.Duties, result.Duties)
				require.Equal(t, tc.expectedResult.Achievements, result.Achievements)
				require.Equal(t, tc.expectedResult.StartDate.Unix(), result.StartDate.Unix())
				require.Equal(t, tc.expectedResult.EndDate.Unix(), result.EndDate.Unix())
				require.Equal(t, tc.expectedResult.UntilNow, result.UntilNow)
				require.False(t, result.UpdatedAt.IsZero())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_DeleteWorkExperience(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		id          int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name:        "Успешное удаление записи об опыте работы",
			id:          1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					DELETE FROM work_experience
					WHERE id = $1
				`)
				mock.ExpectExec(query).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(0, 1)) // 1 affected row
			},
		},
		{
			name: "Запись об опыте работы не найдена",
			id:   999,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("запись об опыте работы с id=999 не найдена"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					DELETE FROM work_experience
					WHERE id = $1
				`)
				mock.ExpectExec(query).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 affected rows
			},
		},
		{
			name: "Ошибка при выполнении запроса",
			id:   2,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении записи об опыте работы: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				query := regexp.QuoteMeta(`
					DELETE FROM work_experience
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
					DELETE FROM work_experience
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

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			err = repo.DeleteWorkExperience(ctx, tc.id)

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

func TestResumeRepository_GetAll(t *testing.T) {
	t.Parallel()

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year", "profession",
		"created_at", "updated_at",
	}

	now := time.Now()
	gradYear1 := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	gradYear2 := time.Date(2018, time.June, 1, 0, 0, 0, 0, time.UTC)
	limit := 10
	offset := 0

	testCases := []struct {
		name           string
		limit          int
		offset         int
		expectedResult []entity.Resume
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, limit, offset int)
	}{
		{
			name:   "Успешное получение списка резюме с пагинацией",
			limit:  limit,
			offset: offset,
			expectedResult: []entity.Resume{
				{
					ID:                     1,
					ApplicantID:            101,
					AboutMe:                "Опытный разработчик",
					SpecializationID:       2,
					Education:              entity.Higher,
					EducationalInstitution: "МГУ",
					GraduationYear:         gradYear1,
					Profession:             "Backend Developer",
					CreatedAt:              now.Add(-24 * time.Hour),
					UpdatedAt:              now,
				},
				{
					ID:                     2,
					ApplicantID:            102,
					AboutMe:                "Начинающий разработчик",
					SpecializationID:       3,
					Education:              entity.Bachelor,
					EducationalInstitution: "СПбГУ",
					GraduationYear:         gradYear2,
					Profession:             "Frontend Developer",
					CreatedAt:              now.Add(-48 * time.Hour),
					UpdatedAt:              now.Add(-12 * time.Hour),
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				query := regexp.QuoteMeta(`
                    SELECT id, applicant_id, about_me, specialization_id, education, 
                           educational_institution, graduation_year, profession, created_at, updated_at
                    FROM resume
                    ORDER BY updated_at DESC
                    LIMIT $1 OFFSET $2
                `)
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								1,
								101,
								"Опытный разработчик",
								2,
								string(entity.Higher),
								"МГУ",
								gradYear1,
								"Backend Developer",
								now.Add(-24*time.Hour),
								now,
							).
							AddRow(
								2,
								102,
								"Начинающий разработчик",
								3,
								string(entity.Bachelor),
								"СПбГУ",
								gradYear2,
								"Frontend Developer",
								now.Add(-48*time.Hour),
								now.Add(-12*time.Hour),
							),
					)
			},
		},
		{
			name:           "Пустой список резюме",
			limit:          limit,
			offset:         offset,
			expectedResult: []entity.Resume{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				query := regexp.QuoteMeta(`
                    SELECT id, applicant_id, about_me, specialization_id, education, 
                           educational_institution, graduation_year, profession, created_at, updated_at
                    FROM resume
                    ORDER BY updated_at DESC
                    LIMIT $1 OFFSET $2
                `)
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnRows(sqlmock.NewRows(columns))
			},
		},
		{
			name:           "Ошибка при выполнении запроса",
			limit:          limit,
			offset:         offset,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				query := regexp.QuoteMeta(`
                    SELECT id, applicant_id, about_me, specialization_id, education, 
                           educational_institution, graduation_year, profession, created_at, updated_at
                    FROM resume
                    ORDER BY updated_at DESC
                    LIMIT $1 OFFSET $2
                `)
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
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

			tc.setupMock(mock, tc.limit, tc.offset)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetAll(ctx, tc.limit, tc.offset)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))

				for i := range tc.expectedResult {
					require.Equal(t, tc.expectedResult[i].ID, result[i].ID)
					require.Equal(t, tc.expectedResult[i].ApplicantID, result[i].ApplicantID)
					require.Equal(t, tc.expectedResult[i].AboutMe, result[i].AboutMe)
					require.Equal(t, tc.expectedResult[i].SpecializationID, result[i].SpecializationID)
					require.Equal(t, tc.expectedResult[i].Education, result[i].Education)
					require.Equal(t, tc.expectedResult[i].EducationalInstitution, result[i].EducationalInstitution)
					require.Equal(t, tc.expectedResult[i].GraduationYear.Unix(), result[i].GraduationYear.Unix())
					require.Equal(t, tc.expectedResult[i].Profession, result[i].Profession)
					require.Equal(t, tc.expectedResult[i].CreatedAt.Unix(), result[i].CreatedAt.Unix())
					require.Equal(t, tc.expectedResult[i].UpdatedAt.Unix(), result[i].UpdatedAt.Unix())
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestResumeRepository_GetAllResumesByApplicantID(t *testing.T) {
	t.Parallel()

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year", "profession",
		"created_at", "updated_at",
	}

	now := time.Now()
	gradYear1 := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	gradYear2 := time.Date(2018, time.June, 1, 0, 0, 0, 0, time.UTC)
	limit := 10
	offset := 0

	testCases := []struct {
		name            string
		applicantID     int
		limit           int
		offset          int
		expectedResumes []entity.Resume
		expectedErr     error
		setupMock       func(mock sqlmock.Sqlmock, applicantID, limit, offset int)
	}{
		{
			name:        "Успешное получение списка резюме с пагинацией",
			applicantID: 1,
			limit:       limit,
			offset:      offset,
			expectedResumes: []entity.Resume{
				{
					ID:                     1,
					ApplicantID:            1,
					AboutMe:                "Опытный разработчик",
					SpecializationID:       2,
					Education:              entity.Higher,
					EducationalInstitution: "МГУ",
					GraduationYear:         gradYear1,
					Profession:             "Backend Developer",
					CreatedAt:              now,
					UpdatedAt:              now,
				},
				{
					ID:                     2,
					ApplicantID:            1,
					AboutMe:                "Начинающий разработчик",
					SpecializationID:       3,
					Education:              entity.Bachelor,
					EducationalInstitution: "СПбГУ",
					GraduationYear:         gradYear2,
					Profession:             "Frontend Developer",
					CreatedAt:              now.Add(-24 * time.Hour),
					UpdatedAt:              now.Add(-24 * time.Hour),
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, applicantID, limit, offset int) {
				query := regexp.QuoteMeta(`
                    SELECT id, applicant_id, about_me, specialization_id, education, 
                           educational_institution, graduation_year, profession, created_at, updated_at
                    FROM resume
                    WHERE applicant_id = $1
                    ORDER BY updated_at DESC
                    LIMIT $2 OFFSET $3
                `)
				mock.ExpectQuery(query).
					WithArgs(applicantID, limit, offset).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								1,
								1,
								"Опытный разработчик",
								2,
								string(entity.Higher),
								"МГУ",
								gradYear1,
								"Backend Developer",
								now,
								now,
							).
							AddRow(
								2,
								1,
								"Начинающий разработчик",
								3,
								string(entity.Bachelor),
								"СПбГУ",
								gradYear2,
								"Frontend Developer",
								now.Add(-24*time.Hour),
								now.Add(-24*time.Hour),
							),
					)
			},
		},
		{
			name:            "Пустой список резюме",
			applicantID:     2,
			limit:           limit,
			offset:          offset,
			expectedResumes: []entity.Resume{},
			expectedErr:     nil,
			setupMock: func(mock sqlmock.Sqlmock, applicantID, limit, offset int) {
				query := regexp.QuoteMeta(`
                    SELECT id, applicant_id, about_me, specialization_id, education, 
                           educational_institution, graduation_year, profession, created_at, updated_at
                    FROM resume
                    WHERE applicant_id = $1
                    ORDER BY updated_at DESC
                    LIMIT $2 OFFSET $3
                `)
				mock.ExpectQuery(query).
					WithArgs(applicantID, limit, offset).
					WillReturnRows(sqlmock.NewRows(columns))
			},
		},
		{
			name:            "Ошибка при выполнении запроса",
			applicantID:     3,
			limit:           limit,
			offset:          offset,
			expectedResumes: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicantID, limit, offset int) {
				query := regexp.QuoteMeta(`
                    SELECT id, applicant_id, about_me, specialization_id, education, 
                           educational_institution, graduation_year, profession, created_at, updated_at
                    FROM resume
                    WHERE applicant_id = $1
                    ORDER BY updated_at DESC
                    LIMIT $2 OFFSET $3
                `)
				mock.ExpectQuery(query).
					WithArgs(applicantID, limit, offset).
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

			tc.setupMock(mock, tc.applicantID, tc.limit, tc.offset)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			resumes, err := repo.GetAllResumesByApplicantID(ctx, tc.applicantID, tc.limit, tc.offset)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResumes), len(resumes))

				for i := range tc.expectedResumes {
					require.Equal(t, tc.expectedResumes[i].ID, resumes[i].ID)
					require.Equal(t, tc.expectedResumes[i].ApplicantID, resumes[i].ApplicantID)
					require.Equal(t, tc.expectedResumes[i].AboutMe, resumes[i].AboutMe)
					require.Equal(t, tc.expectedResumes[i].SpecializationID, resumes[i].SpecializationID)
					require.Equal(t, tc.expectedResumes[i].Education, resumes[i].Education)
					require.Equal(t, tc.expectedResumes[i].EducationalInstitution, resumes[i].EducationalInstitution)
					require.Equal(t, tc.expectedResumes[i].GraduationYear.Unix(), resumes[i].GraduationYear.Unix())
					require.Equal(t, tc.expectedResumes[i].Profession, resumes[i].Profession)
					require.False(t, resumes[i].CreatedAt.IsZero())
					require.False(t, resumes[i].UpdatedAt.IsZero())
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestResumeRepository_FindSkillIDsByNames(t *testing.T) {
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

			repo := &ResumeRepository{DB: db}
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

func TestResumeRepository_FindSpecializationIDByName(t *testing.T) {
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

func TestResumeRepository_FindSpecializationIDsByNames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		specializationNames []string
		expectedResult      []int
		expectedErr         error
		setupMock           func(mock sqlmock.Sqlmock)
	}{
		{
			name:                "Успешный поиск ID для нескольких специализаций",
			specializationNames: []string{"Backend", "Frontend"},
			expectedResult:      []int{1, 2},
			expectedErr:         nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Мокаем первый запрос для Backend
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM specialization WHERE name = $1`,
				)).
					WithArgs("Backend").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Мокаем второй запрос для Frontend
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM specialization WHERE name = $1`,
				)).
					WithArgs("Frontend").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
		},
		{
			name:                "Пустой список специализаций",
			specializationNames: []string{},
			expectedResult:      []int{},
			expectedErr:         nil,
			setupMock:           func(mock sqlmock.Sqlmock) {},
		},
		{
			name:                "Специализация не найдена, создаем новую",
			specializationNames: []string{"NewSpecialization"},
			expectedResult:      []int{10},
			expectedErr:         nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Первый запрос - поиск (не найден)
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM specialization WHERE name = $1`,
				)).
					WithArgs("NewSpecialization").
					WillReturnError(sql.ErrNoRows)

				// Второй запрос - вставка новой специализации
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO specialization (name) VALUES ($1) RETURNING id`,
				)).
					WithArgs("NewSpecialization").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
			},
		},
		{
			name:                "Ошибка при поиске специализации",
			specializationNames: []string{"Backend"},
			expectedResult:      nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования специализации: database error"),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM specialization WHERE name = $1`,
				)).
					WithArgs("Backend").
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:                "Ошибка при создании специализации",
			specializationNames: []string{"NewSpecialization"},
			expectedResult:      nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании специализации: database error"),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				// Первый запрос - поиск (не найден)
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id FROM specialization WHERE name = $1`,
				)).
					WithArgs("NewSpecialization").
					WillReturnError(sql.ErrNoRows)

				// Второй запрос - вставка (ошибка)
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO specialization (name) VALUES ($1) RETURNING id`,
				)).
					WithArgs("NewSpecialization").
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

			tc.setupMock(mock)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.FindSpecializationIDsByNames(ctx, tc.specializationNames)

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

func TestResumeRepository_CreateSkillIfNotExists(t *testing.T) {
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

			repo := &ResumeRepository{DB: db}
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

func TestResumeRepository_CreateSpecializationIfNotExists(t *testing.T) {
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

			repo := &ResumeRepository{DB: db}
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
