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
	now := time.Now()

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year",
		"profession", "created_at", "updated_at",
	}

	query := regexp.QuoteMeta(`
		INSERT INTO resume (
			applicant_id, about_me, specialization_id, education, 
			educational_institution, graduation_year, profession, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, applicant_id, about_me, specialization_id, education, 
				  educational_institution, graduation_year, profession, created_at, updated_at
	`)

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
				Profession:             "Программист",
			},
			expectedResult: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
				CreatedAt:              now,
				UpdatedAt:              now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
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
								resume.Profession,
								now,
								now,
							),
					)
			},
		},
		{
			name: "Успешное создание резюме без профессии",
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
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
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
								resume.Profession,
								now,
								now,
							),
					)
			},
		},
		{
			name: "Ошибка - нарушение уникальности",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("резюме с такими параметрами уже существует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
			},
		},
		{
			name: "Ошибка - обязательное поле отсутствует",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
			},
		},
		{
			name: "Ошибка - неверный тип данных",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
			},
		},
		{
			name: "Ошибка - нарушение проверки данных",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
			},
		},
		{
			name: "Ошибка - внутренняя ошибка базы данных",
			inputResume: &entity.Resume{
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.ApplicantID,
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
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

			tc.setupMock(mock, tc.inputResume)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.Create(ctx, tc.inputResume)

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
				require.Equal(t, tc.expectedResult.ApplicantID, result.ApplicantID)
				require.Equal(t, tc.expectedResult.AboutMe, result.AboutMe)
				require.Equal(t, tc.expectedResult.SpecializationID, result.SpecializationID)
				require.Equal(t, tc.expectedResult.Education, result.Education)
				require.Equal(t, tc.expectedResult.EducationalInstitution, result.EducationalInstitution)
				require.Equal(t, tc.expectedResult.GraduationYear, result.GraduationYear)
				require.Equal(t, tc.expectedResult.Profession, result.Profession)
				require.False(t, result.CreatedAt.IsZero())
				require.False(t, result.UpdatedAt.IsZero())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_AddSkills(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		INSERT INTO resume_skill (resume_id, skill_id)
		VALUES ($1, $2)
	`)

	testCases := []struct {
		name        string
		resumeID    int
		skillIDs    []int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int)
	}{
		{
			name:     "Успешное добавление навыков",
			resumeID: 1,
			skillIDs: []int{1, 2, 3},
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				for _, skillID := range skillIDs {
					mock.ExpectExec(query).
						WithArgs(resumeID, skillID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
				mock.ExpectCommit()
			},
		},
		{
			name:     "Успешное добавление с дубликатами",
			resumeID: 1,
			skillIDs: []int{1, 1, 2},
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectExec(query).
					WithArgs(resumeID, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:     "Ошибка - начало транзакции",
			resumeID: 1,
			skillIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при начале транзакции для добавления навыков: %w", errors.New("transaction error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin().WillReturnError(errors.New("transaction error"))
			},
		},
		{
			name:     "Ошибка - подготовка запроса",
			resumeID: 1,
			skillIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при подготовке запроса для добавления навыков: %w", errors.New("prepare error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).WillReturnError(errors.New("prepare error"))
				mock.ExpectRollback()
			},
		},
		{
			name:     "Ошибка - обязательное поле отсутствует",
			resumeID: 1,
			skillIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:     "Ошибка - неверный формат данных",
			resumeID: 1,
			skillIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:     "Ошибка - неверные данные",
			resumeID: 1,
			skillIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:     "Ошибка - внутренняя ошибка при выполнении запроса",
			resumeID: 1,
			skillIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении навыка к резюме: %w", errors.New("exec error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnError(errors.New("exec error"))
				mock.ExpectRollback()
			},
		},
		{
			name:     "Ошибка - коммит транзакции",
			resumeID: 1,
			skillIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при коммите транзакции добавления навыков: %w", errors.New("commit error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, skillIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				for _, skillID := range skillIDs {
					mock.ExpectExec(query).
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
			defer db.Close()

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

	startDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	columns := []string{
		"id", "resume_id", "employer_name", "position", "duties",
		"achievements", "start_date", "end_date", "until_now", "updated_at",
	}

	query := regexp.QuoteMeta(`
		INSERT INTO work_experience (
			resume_id, employer_name, position, duties, 
			achievements, start_date, end_date, until_now, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING id, resume_id, employer_name, position, duties, 
				  achievements, start_date, end_date, until_now, updated_at
	`)

	testCases := []struct {
		name                string
		inputWorkExperience *entity.WorkExperience
		expectedResult      *entity.WorkExperience
		expectedErr         error
		setupMock           func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience)
	}{
		{
			name: "Успешное добавление опыта работы с датой окончания",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Разработчик",
				Duties:       "Разработка ПО",
				Achievements: "Внедрил систему CI/CD",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: &entity.WorkExperience{
				ID:           1,
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Разработчик",
				Duties:       "Разработка ПО",
				Achievements: "Внедрил систему CI/CD",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
				UpdatedAt:    now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.ResumeID,
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								1,
								workExp.ResumeID,
								workExp.EmployerName,
								workExp.Position,
								workExp.Duties,
								workExp.Achievements,
								workExp.StartDate,
								endDateNullTime,
								workExp.UntilNow,
								now,
							),
					)
			},
		},
		{
			name: "Успешное добавление опыта работы с until_now=true",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:     1,
				EmployerName: "Startup Inc",
				Position:     "Ведущий разработчик",
				Duties:       "Лидерство в команде",
				Achievements: "Запуск продукта",
				StartDate:    startDate,
				UntilNow:     true,
			},
			expectedResult: &entity.WorkExperience{
				ID:           2,
				ResumeID:     1,
				EmployerName: "Startup Inc",
				Position:     "Ведущий разработчик",
				Duties:       "Лидерство в команде",
				Achievements: "Запуск продукта",
				StartDate:    startDate,
				UntilNow:     true,
				UpdatedAt:    now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Valid: false}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.ResumeID,
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								2,
								workExp.ResumeID,
								workExp.EmployerName,
								workExp.Position,
								workExp.Duties,
								workExp.Achievements,
								workExp.StartDate,
								endDateNullTime,
								workExp.UntilNow,
								now,
							),
					)
			},
		},
		{
			name: "Ошибка - нарушение уникальности",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Разработчик",
				Duties:       "Разработка ПО",
				Achievements: "Внедрил систему CI/CD",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("опыт работы с такими параметрами уже существует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.ResumeID,
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
			},
		},
		{
			name: "Ошибка - обязательное поле отсутствует",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Разработчик",
				Duties:       "Разработка ПО",
				Achievements: "Внедрил систему CI/CD",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.ResumeID,
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
			},
		},
		{
			name: "Ошибка - неверный формат данных",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Разработчик",
				Duties:       "Разработка ПО",
				Achievements: "Внедрил систему CI/CD",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.ResumeID,
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
			},
		},
		{
			name: "Ошибка - неверные данные",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Разработчик",
				Duties:       "Разработка ПО",
				Achievements: "Внедрил систему CI/CD",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.ResumeID,
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
			},
		},
		{
			name: "Ошибка - внутренняя ошибка базы данных",
			inputWorkExperience: &entity.WorkExperience{
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Разработчик",
				Duties:       "Разработка ПО",
				Achievements: "Внедрил систему CI/CD",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании опыта работы: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.ResumeID,
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
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

			tc.setupMock(mock, tc.inputWorkExperience)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.AddWorkExperience(ctx, tc.inputWorkExperience)

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
				require.Equal(t, tc.expectedResult.ResumeID, result.ResumeID)
				require.Equal(t, tc.expectedResult.EmployerName, result.EmployerName)
				require.Equal(t, tc.expectedResult.Position, result.Position)
				require.Equal(t, tc.expectedResult.Duties, result.Duties)
				require.Equal(t, tc.expectedResult.Achievements, result.Achievements)
				require.Equal(t, tc.expectedResult.StartDate, result.StartDate)
				require.Equal(t, tc.expectedResult.EndDate, result.EndDate)
				require.Equal(t, tc.expectedResult.UntilNow, result.UntilNow)
				require.False(t, result.UpdatedAt.IsZero())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_GetByID(t *testing.T) {
	t.Parallel()

	graduationDate := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year",
		"profession", "created_at", "updated_at",
	}

	query := regexp.QuoteMeta(`
		SELECT id, applicant_id, about_me, specialization_id, education, 
			   educational_institution, graduation_year, profession, created_at, updated_at
		FROM resume
		WHERE id = $1
	`)

	testCases := []struct {
		name           string
		resumeID       int
		expectedResult *entity.Resume
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, resumeID int)
	}{
		{
			name:     "Успешное получение резюме",
			resumeID: 1,
			expectedResult: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Опытный разработчик",
				SpecializationID:       2,
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
				CreatedAt:              now,
				UpdatedAt:              now,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								1,
								1,
								"Опытный разработчик",
								2,
								string(entity.Higher),
								"МГУ",
								graduationDate,
								"Программист",
								now,
								now,
							),
					)
			},
		},
		{
			name:           "Ошибка - резюме не найдено",
			resumeID:       999,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=%d не найдено", 999),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка - внутренняя ошибка базы данных",
			resumeID:       1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить резюме по id=%d", 1),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectQuery(query).
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
			defer db.Close()

			tc.setupMock(mock, tc.resumeID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetByID(ctx, tc.resumeID)

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
				require.Equal(t, tc.expectedResult.ApplicantID, result.ApplicantID)
				require.Equal(t, tc.expectedResult.AboutMe, result.AboutMe)
				require.Equal(t, tc.expectedResult.SpecializationID, result.SpecializationID)
				require.Equal(t, tc.expectedResult.Education, result.Education)
				require.Equal(t, tc.expectedResult.EducationalInstitution, result.EducationalInstitution)
				require.Equal(t, tc.expectedResult.GraduationYear, result.GraduationYear)
				require.Equal(t, tc.expectedResult.Profession, result.Profession)
				require.Equal(t, tc.expectedResult.CreatedAt, result.CreatedAt)
				require.Equal(t, tc.expectedResult.UpdatedAt, result.UpdatedAt)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_GetSkillsByResumeID(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		SELECT s.id, s.name
		FROM skill s
		JOIN resume_skill rs ON s.id = rs.skill_id
		WHERE rs.resume_id = $1
	`)

	columns := []string{"id", "name"}

	testCases := []struct {
		name           string
		resumeID       int
		expectedResult []entity.Skill
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, resumeID int)
	}{
		{
			name:     "Успешное получение навыков",
			resumeID: 1,
			expectedResult: []entity.Skill{
				{ID: 1, Name: "Go"},
				{ID: 2, Name: "SQL"},
				{ID: 3, Name: "Docker"},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Go").
					AddRow(2, "SQL").
					AddRow(3, "Docker")
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение - нет навыков",
			resumeID:       2,
			expectedResult: []entity.Skill{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - внутренняя ошибка при выполнении запроса",
			resumeID:       1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении навыков резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка - ошибка при сканировании",
			resumeID:       1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании навыка: %w", errors.New("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns).
					AddRow("invalid", "Go") // Некорректное значение для id (строка вместо числа)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при итерации",
			resumeID:       1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по навыкам: %w", errors.New("iteration error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Go").
					AddRow(2, "SQL")
				mock.ExpectQuery(query).
					WithArgs(resumeID).
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
			defer db.Close()

			tc.setupMock(mock, tc.resumeID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetSkillsByResumeID(ctx, tc.resumeID)

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

func TestResumeRepository_GetWorkExperienceByResumeID(t *testing.T) {
	t.Parallel()

	startDate1 := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate1 := time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC)
	startDate2 := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	columns := []string{
		"id", "resume_id", "employer_name", "position", "duties",
		"achievements", "start_date", "end_date", "until_now", "updated_at",
	}

	query := regexp.QuoteMeta(`
		SELECT id, resume_id, employer_name, position, duties, 
			   achievements, start_date, end_date, until_now, updated_at
		FROM work_experience
		WHERE resume_id = $1
		ORDER BY start_date DESC
	`)

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
					EmployerName: "Tech Corp",
					Position:     "Разработчик",
					Duties:       "Разработка ПО",
					Achievements: "Внедрил CI/CD",
					StartDate:    startDate1,
					EndDate:      endDate1,
					UntilNow:     false,
					UpdatedAt:    now,
				},
				{
					ID:           2,
					ResumeID:     1,
					EmployerName: "Startup Inc",
					Position:     "Инженер",
					Duties:       "Поддержка систем",
					Achievements: "Оптимизация процессов",
					StartDate:    startDate2,
					UntilNow:     true,
					UpdatedAt:    now,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						1,
						"Tech Corp",
						"Разработчик",
						"Разработка ПО",
						"Внедрил CI/CD",
						startDate1,
						sql.NullTime{Time: endDate1, Valid: true},
						false,
						now,
					).
					AddRow(
						2,
						1,
						"Startup Inc",
						"Инженер",
						"Поддержка систем",
						"Оптимизация процессов",
						startDate2,
						sql.NullTime{Valid: false},
						true,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение - нет опыта работы",
			resumeID:       2,
			expectedResult: []entity.WorkExperience{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - внутренняя ошибка при выполнении запроса",
			resumeID:       1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении опыта работы: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка - ошибка при сканировании",
			resumeID:       1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании опыта работы: %w", errors.New("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						"invalid", // Некорректное значение для id
						1,
						"Tech Corp",
						"Разработчик",
						"Разработка ПО",
						"Внедрил CI/CD",
						startDate1,
						sql.NullTime{Time: endDate1, Valid: true},
						false,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при итерации",
			resumeID:       1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по опыту работы: %w", errors.New("iteration error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						1,
						"Tech Corp",
						"Разработчик",
						"Разработка ПО",
						"Внедрил CI/CD",
						startDate1,
						sql.NullTime{Time: endDate1, Valid: true},
						false,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
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
			defer db.Close()

			tc.setupMock(mock, tc.resumeID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetWorkExperienceByResumeID(ctx, tc.resumeID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))
				for i, expectedExp := range tc.expectedResult {
					require.Equal(t, expectedExp.ID, result[i].ID)
					require.Equal(t, expectedExp.ResumeID, result[i].ResumeID)
					require.Equal(t, expectedExp.EmployerName, result[i].EmployerName)
					require.Equal(t, expectedExp.Position, result[i].Position)
					require.Equal(t, expectedExp.Duties, result[i].Duties)
					require.Equal(t, expectedExp.Achievements, result[i].Achievements)
					require.Equal(t, expectedExp.StartDate, result[i].StartDate)
					require.Equal(t, expectedExp.EndDate, result[i].EndDate)
					require.Equal(t, expectedExp.UntilNow, result[i].UntilNow)
					require.Equal(t, expectedExp.UpdatedAt, result[i].UpdatedAt)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_AddSpecializations(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		INSERT INTO resume_specialization (resume_id, specialization_id)
		VALUES ($1, $2)
	`)

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
			specializationIDs: []int{1, 2, 3},
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				for _, specializationID := range specializationIDs {
					mock.ExpectExec(query).
						WithArgs(resumeID, specializationID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
				mock.ExpectCommit()
			},
		},
		{
			name:              "Успешное добавление с дубликатами",
			resumeID:          1,
			specializationIDs: []int{1, 1, 2},
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectExec(query).
					WithArgs(resumeID, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:              "Ошибка - начало транзакции",
			resumeID:          1,
			specializationIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при начале транзакции для добавления специализаций: %w", errors.New("transaction error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin().WillReturnError(errors.New("transaction error"))
			},
		},
		{
			name:              "Ошибка - подготовка запроса",
			resumeID:          1,
			specializationIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при подготовке запроса для добавления специализаций: %w", errors.New("prepare error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query).WillReturnError(errors.New("prepare error"))
				mock.ExpectRollback()
			},
		},
		{
			name:              "Ошибка - обязательное поле отсутствует",
			resumeID:          1,
			specializationIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:              "Ошибка - неверный формат данных",
			resumeID:          1,
			specializationIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:              "Ошибка - неверные данные",
			resumeID:          1,
			specializationIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
				mock.ExpectRollback()
			},
		},
		{
			name:              "Ошибка - внутренняя ошибка при выполнении запроса",
			resumeID:          1,
			specializationIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении специализации к резюме: %w", errors.New("exec error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(resumeID, 1).
					WillReturnError(errors.New("exec error"))
				mock.ExpectRollback()
			},
		},
		{
			name:              "Ошибка - коммит транзакции",
			resumeID:          1,
			specializationIDs: []int{1, 2},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при коммите транзакции добавления специализаций: %w", errors.New("commit error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int, specializationIDs []int) {
				mock.ExpectBegin()
				mock.ExpectPrepare(query)
				for _, specializationID := range specializationIDs {
					mock.ExpectExec(query).
						WithArgs(resumeID, specializationID).
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
			defer db.Close()

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

	query := regexp.QuoteMeta(`
		SELECT s.id, s.name
		FROM specialization s
		JOIN resume_specialization rs ON s.id = rs.specialization_id
		WHERE rs.resume_id = $1
	`)

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
				{ID: 1, Name: "Backend Development"},
				{ID: 2, Name: "DevOps"},
				{ID: 3, Name: "Database Administration"},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Backend Development").
					AddRow(2, "DevOps").
					AddRow(3, "Database Administration")
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение - нет специализаций",
			resumeID:       2,
			expectedResult: []entity.Specialization{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - внутренняя ошибка при выполнении запроса",
			resumeID:       1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении специализаций резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка - ошибка при сканировании",
			resumeID:       1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании специализации: %w", errors.New("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns).
					AddRow("invalid", "Backend Development") // Некорректное значение для id
				mock.ExpectQuery(query).
					WithArgs(resumeID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при итерации",
			resumeID:       1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по специализациям: %w", errors.New("iteration error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(1, "Backend Development").
					AddRow(2, "DevOps")
				mock.ExpectQuery(query).
					WithArgs(resumeID).
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
			defer db.Close()

			tc.setupMock(mock, tc.resumeID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetSpecializationsByResumeID(ctx, tc.resumeID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(result))
				for i, expectedSpec := range tc.expectedResult {
					require.Equal(t, expectedSpec.ID, result[i].ID)
					require.Equal(t, expectedSpec.Name, result[i].Name)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_Update(t *testing.T) {
	t.Parallel()

	graduationDate := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year",
		"profession", "created_at", "updated_at",
	}

	query := regexp.QuoteMeta(`
		UPDATE resume
		SET 
			about_me = $1,
			specialization_id = $2,
			education = $3,
			educational_institution = $4,
			graduation_year = $5,
			profession = $6,
			updated_at = NOW()
		WHERE id = $7 AND applicant_id = $8
		RETURNING id, applicant_id, about_me, specialization_id, education, 
				  educational_institution, graduation_year, profession, created_at, updated_at
	`)

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
				AboutMe:                "Обновленный разработчик",
				SpecializationID:       3,
				Education:              entity.Higher,
				EducationalInstitution: "МГТУ",
				GraduationYear:         graduationDate,
				Profession:             "Ведущий программист",
			},
			expectedResult: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Обновленный разработчик",
				SpecializationID:       3,
				Education:              entity.Higher,
				EducationalInstitution: "МГТУ",
				GraduationYear:         graduationDate,
				Profession:             "Ведущий программист",
				CreatedAt:              createdAt,
				UpdatedAt:              updatedAt,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
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
								resume.Profession,
								createdAt,
								updatedAt,
							),
					)
			},
		},
		{
			name: "Ошибка - резюме не найдено",
			inputResume: &entity.Resume{
				ID:                     999,
				ApplicantID:            1,
				AboutMe:                "Не найденный разработчик",
				SpecializationID:       3,
				Education:              entity.Higher,
				EducationalInstitution: "МГТУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=%d не найдено или не принадлежит указанному соискателю", 999),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
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
				AboutMe:                "Обновленный разработчик",
				SpecializationID:       3,
				Education:              entity.Higher,
				EducationalInstitution: "МГТУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("резюме с такими параметрами уже существует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
						resume.ID,
						resume.ApplicantID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
			},
		},
		{
			name: "Ошибка - обязательное поле отсутствует",
			inputResume: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Обновленный разработчик",
				SpecializationID:       3,
				Education:              entity.Higher,
				EducationalInstitution: "МГТУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
						resume.ID,
						resume.ApplicantID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
			},
		},
		{
			name: "Ошибка - неверный формат данных",
			inputResume: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Обновленный разработчик",
				SpecializationID:       3,
				Education:              entity.Higher,
				EducationalInstitution: "МГТУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
						resume.ID,
						resume.ApplicantID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
			},
		},
		{
			name: "Ошибка - неверные данные",
			inputResume: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Обновленный разработчик",
				SpecializationID:       3,
				Education:              entity.Higher,
				EducationalInstitution: "МГТУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
						resume.ID,
						resume.ApplicantID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
			},
		},
		{
			name: "Ошибка - внутренняя ошибка базы данных",
			inputResume: &entity.Resume{
				ID:                     1,
				ApplicantID:            1,
				AboutMe:                "Обновленный разработчик",
				SpecializationID:       3,
				Education:              entity.Higher,
				EducationalInstitution: "МГТУ",
				GraduationYear:         graduationDate,
				Profession:             "Программист",
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обновлении резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resume *entity.Resume) {
				mock.ExpectQuery(query).
					WithArgs(
						resume.AboutMe,
						resume.SpecializationID,
						string(resume.Education),
						resume.EducationalInstitution,
						resume.GraduationYear,
						resume.Profession,
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
			defer db.Close()

			tc.setupMock(mock, tc.inputResume)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.Update(ctx, tc.inputResume)

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
				require.Equal(t, tc.expectedResult.ApplicantID, result.ApplicantID)
				require.Equal(t, tc.expectedResult.AboutMe, result.AboutMe)
				require.Equal(t, tc.expectedResult.SpecializationID, result.SpecializationID)
				require.Equal(t, tc.expectedResult.Education, result.Education)
				require.Equal(t, tc.expectedResult.EducationalInstitution, result.EducationalInstitution)
				require.Equal(t, tc.expectedResult.GraduationYear, result.GraduationYear)
				require.Equal(t, tc.expectedResult.Profession, result.Profession)
				require.Equal(t, tc.expectedResult.CreatedAt, result.CreatedAt)
				require.False(t, result.UpdatedAt.IsZero())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_Delete(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		DELETE FROM resume
		WHERE id = $1
	`)

	testCases := []struct {
		name        string
		resumeID    int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, resumeID int)
	}{
		{
			name:     "Успешное удаление резюме",
			resumeID: 1,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:     "Ошибка - резюме не найдено",
			resumeID: 999,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=%d не найдено", 999),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:     "Ошибка - внутренняя ошибка при выполнении запроса",
			resumeID: 1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:     "Ошибка - ошибка при получении количества затронутых строк",
			resumeID: 1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении количества затронутых строк: %w", errors.New("rows affected error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectExec(query).
					WithArgs(resumeID).
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
			defer db.Close()

			tc.setupMock(mock, tc.resumeID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			err = repo.Delete(ctx, tc.resumeID)

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

	query := regexp.QuoteMeta(`
		DELETE FROM resume_skill
		WHERE resume_id = $1
	`)

	testCases := []struct {
		name        string
		resumeID    int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, resumeID int)
	}{
		{
			name:     "Успешное удаление навыков",
			resumeID: 1,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:     "Успешное удаление - навыки отсутствуют",
			resumeID: 2,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:     "Ошибка - внутренняя ошибка при выполнении запроса",
			resumeID: 1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении навыков резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
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
			defer db.Close()

			tc.setupMock(mock, tc.resumeID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

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

	query := regexp.QuoteMeta(`
		DELETE FROM resume_specialization
		WHERE resume_id = $1
	`)

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
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:     "Успешное удаление - специализации отсутствуют",
			resumeID: 2,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:     "Ошибка - внутренняя ошибка при выполнении запроса",
			resumeID: 1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении специализаций резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
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
			defer db.Close()

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

	query := regexp.QuoteMeta(`
		DELETE FROM work_experience
		WHERE resume_id = $1
	`)

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
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:     "Успешное удаление - опыт работы отсутствует",
			resumeID: 2,
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
				mock.ExpectExec(query).
					WithArgs(resumeID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:     "Ошибка - внутренняя ошибка при выполнении запроса",
			resumeID: 1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении опыта работы резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, resumeID int) {
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
			defer db.Close()

			tc.setupMock(mock, tc.resumeID)

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
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestResumeRepository_UpdateWorkExperience(t *testing.T) {
	t.Parallel()

	startDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Now()

	columns := []string{
		"id", "resume_id", "employer_name", "position", "duties",
		"achievements", "start_date", "end_date", "until_now", "updated_at",
	}

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

	testCases := []struct {
		name                string
		inputWorkExperience *entity.WorkExperience
		expectedResult      *entity.WorkExperience
		expectedErr         error
		setupMock           func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience)
	}{
		{
			name: "Успешное обновление опыта работы с датой окончания",
			inputWorkExperience: &entity.WorkExperience{
				ID:           1,
				ResumeID:     1,
				EmployerName: "Updated Corp",
				Position:     "Senior Developer",
				Duties:       "Leading development",
				Achievements: "Improved performance",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: &entity.WorkExperience{
				ID:           1,
				ResumeID:     1,
				EmployerName: "Updated Corp",
				Position:     "Senior Developer",
				Duties:       "Leading development",
				Achievements: "Improved performance",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
				UpdatedAt:    updatedAt,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
						workExp.ID,
						workExp.ResumeID,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								workExp.ID,
								workExp.ResumeID,
								workExp.EmployerName,
								workExp.Position,
								workExp.Duties,
								workExp.Achievements,
								workExp.StartDate,
								endDateNullTime,
								workExp.UntilNow,
								updatedAt,
							),
					)
			},
		},
		{
			name: "Успешное обновление опыта работы с until_now=true",
			inputWorkExperience: &entity.WorkExperience{
				ID:           2,
				ResumeID:     1,
				EmployerName: "Startup Inc",
				Position:     "Tech Lead",
				Duties:       "Team leadership",
				Achievements: "Product launch",
				StartDate:    startDate,
				UntilNow:     true,
			},
			expectedResult: &entity.WorkExperience{
				ID:           2,
				ResumeID:     1,
				EmployerName: "Startup Inc",
				Position:     "Tech Lead",
				Duties:       "Team leadership",
				Achievements: "Product launch",
				StartDate:    startDate,
				UntilNow:     true,
				UpdatedAt:    updatedAt,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Valid: false}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
						workExp.ID,
						workExp.ResumeID,
					).
					WillReturnRows(
						sqlmock.NewRows(columns).
							AddRow(
								workExp.ID,
								workExp.ResumeID,
								workExp.EmployerName,
								workExp.Position,
								workExp.Duties,
								workExp.Achievements,
								workExp.StartDate,
								endDateNullTime,
								workExp.UntilNow,
								updatedAt,
							),
					)
			},
		},
		{
			name: "Ошибка - запись об опыте работы не найдена",
			inputWorkExperience: &entity.WorkExperience{
				ID:           999,
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Developer",
				Duties:       "Development",
				Achievements: "CI/CD implementation",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("запись об опыте работы с id=%d не найдена или не принадлежит указанному резюме", 999),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
						workExp.ID,
						workExp.ResumeID,
					).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name: "Ошибка - нарушение уникальности",
			inputWorkExperience: &entity.WorkExperience{
				ID:           1,
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Developer",
				Duties:       "Development",
				Achievements: "CI/CD implementation",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("запись об опыте работы с такими параметрами уже существует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
						workExp.ID,
						workExp.ResumeID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
			},
		},
		{
			name: "Ошибка - обязательное поле отсутствует",
			inputWorkExperience: &entity.WorkExperience{
				ID:           1,
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Developer",
				Duties:       "Development",
				Achievements: "CI/CD implementation",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
						workExp.ID,
						workExp.ResumeID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
			},
		},
		{
			name: "Ошибка - неверный формат данных",
			inputWorkExperience: &entity.WorkExperience{
				ID:           1,
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Developer",
				Duties:       "Development",
				Achievements: "CI/CD implementation",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
						workExp.ID,
						workExp.ResumeID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
			},
		},
		{
			name: "Ошибка - неверные данные",
			inputWorkExperience: &entity.WorkExperience{
				ID:           1,
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Developer",
				Duties:       "Development",
				Achievements: "CI/CD implementation",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
						workExp.ID,
						workExp.ResumeID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
			},
		},
		{
			name: "Ошибка - внутренняя ошибка базы данных",
			inputWorkExperience: &entity.WorkExperience{
				ID:           1,
				ResumeID:     1,
				EmployerName: "Tech Corp",
				Position:     "Developer",
				Duties:       "Development",
				Achievements: "CI/CD implementation",
				StartDate:    startDate,
				EndDate:      endDate,
				UntilNow:     false,
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обновлении записи об опыте работы: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExp *entity.WorkExperience) {
				endDateNullTime := sql.NullTime{Time: workExp.EndDate, Valid: true}
				mock.ExpectQuery(query).
					WithArgs(
						workExp.EmployerName,
						workExp.Position,
						workExp.Duties,
						workExp.Achievements,
						workExp.StartDate,
						endDateNullTime,
						workExp.UntilNow,
						workExp.ID,
						workExp.ResumeID,
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

			tc.setupMock(mock, tc.inputWorkExperience)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.UpdateWorkExperience(ctx, tc.inputWorkExperience)

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
				require.Equal(t, tc.expectedResult.ResumeID, result.ResumeID)
				require.Equal(t, tc.expectedResult.EmployerName, result.EmployerName)
				require.Equal(t, tc.expectedResult.Position, result.Position)
				require.Equal(t, tc.expectedResult.Duties, result.Duties)
				require.Equal(t, tc.expectedResult.Achievements, result.Achievements)
				require.Equal(t, tc.expectedResult.StartDate, result.StartDate)
				require.Equal(t, tc.expectedResult.EndDate, result.EndDate)
				require.Equal(t, tc.expectedResult.UntilNow, result.UntilNow)
				require.False(t, result.UpdatedAt.IsZero())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_DeleteWorkExperience(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		DELETE FROM work_experience
		WHERE id = $1
	`)

	testCases := []struct {
		name        string
		workExpID   int
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, workExpID int)
	}{
		{
			name:      "Успешное удаление записи об опыте работы",
			workExpID: 1,
			setupMock: func(mock sqlmock.Sqlmock, workExpID int) {
				mock.ExpectExec(query).
					WithArgs(workExpID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:      "Ошибка - запись об опыте работы не найдена",
			workExpID: 999,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("запись об опыте работы с id=%d не найдена", 999),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExpID int) {
				mock.ExpectExec(query).
					WithArgs(workExpID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
		{
			name:      "Ошибка - внутренняя ошибка при выполнении запроса",
			workExpID: 1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении записи об опыте работы: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExpID int) {
				mock.ExpectExec(query).
					WithArgs(workExpID).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:      "Ошибка - ошибка при получении количества затронутых строк",
			workExpID: 1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении количества затронутых строк: %w", errors.New("rows affected error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, workExpID int) {
				mock.ExpectExec(query).
					WithArgs(workExpID).
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
			defer db.Close()

			tc.setupMock(mock, tc.workExpID)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			err = repo.DeleteWorkExperience(ctx, tc.workExpID)

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

	graduationDate := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Now().Add(-48 * time.Hour)
	updatedAt := time.Now()

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year",
		"profession", "created_at", "updated_at",
	}

	query := regexp.QuoteMeta(`
		SELECT id, applicant_id, about_me, specialization_id, education, 
			   educational_institution, graduation_year, profession, created_at, updated_at
		FROM resume
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`)

	testCases := []struct {
		name           string
		limit          int
		offset         int
		expectedResult []entity.Resume
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, limit, offset int)
	}{
		{
			name:   "Успешное получение списка резюме",
			limit:  2,
			offset: 0,
			expectedResult: []entity.Resume{
				{
					ID:                     1,
					ApplicantID:            1,
					AboutMe:                "Опытный разработчик",
					SpecializationID:       2,
					Education:              entity.Higher,
					EducationalInstitution: "МГУ",
					GraduationYear:         graduationDate,
					Profession:             "Программист",
					CreatedAt:              createdAt,
					UpdatedAt:              updatedAt,
				},
				{
					ID:                     2,
					ApplicantID:            2,
					AboutMe:                "Младший разработчик",
					SpecializationID:       3,
					Education:              entity.Higher,
					EducationalInstitution: "МГТУ",
					GraduationYear:         graduationDate,
					Profession:             "Младший программист",
					CreatedAt:              createdAt,
					UpdatedAt:              updatedAt,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						1,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Программист",
						createdAt,
						updatedAt,
					).
					AddRow(
						2,
						2,
						"Младший разработчик",
						3,
						string(entity.Higher),
						"МГТУ",
						graduationDate,
						"Младший программист",
						createdAt,
						updatedAt,
					)
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение - пустой список",
			limit:          2,
			offset:         10,
			expectedResult: []entity.Resume{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - внутренняя ошибка при выполнении запроса",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка - ошибка при сканировании",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании резюме: %w", errors.New("sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"invalid\") to a int: invalid syntax")),
			),
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						"invalid", // Некорректное значение для id
						1,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Программист",
						createdAt,
						updatedAt,
					)
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка - ошибка при итерации",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по резюме: %w", errors.New("iteration error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						1,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Программист",
						createdAt,
						updatedAt,
					)
				mock.ExpectQuery(query).
					WithArgs(limit, offset).
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
			defer db.Close()

			tc.setupMock(mock, tc.limit, tc.offset)

			repo := &ResumeRepository{DB: db}
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
				for i, expectedResume := range tc.expectedResult {
					require.Equal(t, expectedResume.ID, result[i].ID)
					require.Equal(t, expectedResume.ApplicantID, result[i].ApplicantID)
					require.Equal(t, expectedResume.AboutMe, result[i].AboutMe)
					require.Equal(t, expectedResume.SpecializationID, result[i].SpecializationID)
					require.Equal(t, expectedResume.Education, result[i].Education)
					require.Equal(t, expectedResume.EducationalInstitution, result[i].EducationalInstitution)
					require.Equal(t, expectedResume.GraduationYear, result[i].GraduationYear)
					require.Equal(t, expectedResume.Profession, result[i].Profession)
					require.Equal(t, expectedResume.CreatedAt, result[i].CreatedAt)
					require.Equal(t, expectedResume.UpdatedAt, result[i].UpdatedAt)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestResumeRepository_GetAllResumesByApplicantID(t *testing.T) {
	t.Parallel()

	graduationDate := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year",
		"profession", "created_at", "updated_at",
	}

	query := regexp.QuoteMeta(`
		SELECT id, applicant_id, about_me, specialization_id, education, 
			   educational_institution, graduation_year, profession, created_at, updated_at
		FROM resume
		WHERE applicant_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`)

	testCases := []struct {
		name           string
		applicantID    int
		limit          int
		offset         int
		expectedResult []entity.Resume
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, applicantID, limit, offset int)
	}{
		{
			name:        "Успешное получение списка резюме",
			applicantID: 1,
			limit:       2,
			offset:      0,
			expectedResult: []entity.Resume{
				{
					ID:                     1,
					ApplicantID:            1,
					AboutMe:                "Опытный разработчик",
					SpecializationID:       2,
					Education:              entity.Higher,
					EducationalInstitution: "МГУ",
					GraduationYear:         graduationDate,
					Profession:             "Программист",
					CreatedAt:              now,
					UpdatedAt:              now,
				},
				{
					ID:                     2,
					ApplicantID:            1,
					AboutMe:                "Начинающий разработчик",
					SpecializationID:       3,
					Education:              entity.SecondarySchool,
					EducationalInstitution: "Школа №123",
					GraduationYear:         graduationDate,
					Profession:             "",
					CreatedAt:              now,
					UpdatedAt:              now,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, applicantID, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						applicantID,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Программист",
						now,
						now,
					).
					AddRow(
						2,
						applicantID,
						"Начинающий разработчик",
						3,
						string(entity.SecondarySchool),
						"Школа №123",
						graduationDate,
						"",
						now,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs(applicantID, limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Пустой список резюме",
			applicantID:    1,
			limit:          10,
			offset:         0,
			expectedResult: []entity.Resume{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, applicantID, limit, offset int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(applicantID, limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка базы данных при выполнении запроса",
			applicantID:    1,
			limit:          10,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка резюме: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicantID, limit, offset int) {
				mock.ExpectQuery(query).
					WithArgs(applicantID, limit, offset).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка сканирования строк",
			applicantID:    1,
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по резюме: %w", errors.New("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicantID, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						applicantID,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Программист",
						now,
						now,
					).
					RowError(0, errors.New("scan error"))
				mock.ExpectQuery(query).
					WithArgs(applicantID, limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка итерации по строкам",
			applicantID:    1,
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по резюме: %w", errors.New("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicantID, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						applicantID,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Программист",
						now,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs(applicantID, limit, offset).
					WillReturnRows(rows)
				rows.CloseError(errors.New("rows error"))
			},
		},
		{
			name:           "Ошибка закрытия строк",
			applicantID:    1,
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по резюме: %w", errors.New("close error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicantID, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						applicantID,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Программист",
						now,
						now,
					)
				rows.CloseError(errors.New("close error"))
				mock.ExpectQuery(query).
					WithArgs(applicantID, limit, offset).
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

			tc.setupMock(mock, tc.applicantID, tc.limit, tc.offset)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetAllResumesByApplicantID(ctx, tc.applicantID, tc.limit, tc.offset)

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
					require.Equal(t, expected.ApplicantID, result[i].ApplicantID)
					require.Equal(t, expected.AboutMe, result[i].AboutMe)
					require.Equal(t, expected.SpecializationID, result[i].SpecializationID)
					require.Equal(t, expected.Education, result[i].Education)
					require.Equal(t, expected.EducationalInstitution, result[i].EducationalInstitution)
					require.Equal(t, expected.GraduationYear, result[i].GraduationYear)
					require.Equal(t, expected.Profession, result[i].Profession)
					require.False(t, result[i].CreatedAt.IsZero())
					require.False(t, result[i].UpdatedAt.IsZero())
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestResumeRepository_FindSkillIDsByNames(t *testing.T) {
	t.Parallel()

	querySelect := regexp.QuoteMeta(`SELECT id FROM skill WHERE name = $1`)
	queryInsert := regexp.QuoteMeta(`INSERT INTO skill (name) VALUES ($1) RETURNING id`)

	testCases := []struct {
		name           string
		skillNames     []string
		expectedResult []int
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, skillNames []string)
	}{
		{
			name:           "Пустой список навыков",
			skillNames:     []string{},
			expectedResult: []int{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				// No database queries expected for empty input
			},
		},
		{
			name:           "Успешное получение ID для существующих навыков",
			skillNames:     []string{"Go", "SQL"},
			expectedResult: []int{1, 2},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				for i, name := range skillNames {
					mock.ExpectQuery(querySelect).
						WithArgs(name).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 1))
				}
			},
		},
		{
			name:           "Успешное создание новых навыков",
			skillNames:     []string{"Python", "Docker"},
			expectedResult: []int{3, 4},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				for i, name := range skillNames {
					// Simulate skill not found
					mock.ExpectQuery(querySelect).
						WithArgs(name).
						WillReturnError(sql.ErrNoRows)
					// Simulate skill creation
					mock.ExpectQuery(queryInsert).
						WithArgs(name).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 3))
				}
			},
		},
		{
			name:           "Смешанный случай: существующие и новые навыки",
			skillNames:     []string{"Go", "Kubernetes"},
			expectedResult: []int{1, 5},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				// First skill exists
				mock.ExpectQuery(querySelect).
					WithArgs("Go").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				// Second skill does not exist
				mock.ExpectQuery(querySelect).
					WithArgs("Kubernetes").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(queryInsert).
					WithArgs("Kubernetes").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
			},
		},
		{
			name:           "Ошибка при проверке существования навыка",
			skillNames:     []string{"Go"},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования навыка: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				mock.ExpectQuery(querySelect).
					WithArgs("Go").
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка при создании нового навыка",
			skillNames:     []string{"Rust"},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании навыка: %w", errors.New("insert error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillNames []string) {
				mock.ExpectQuery(querySelect).
					WithArgs("Rust").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(queryInsert).
					WithArgs("Rust").
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
			defer db.Close()

			tc.setupMock(mock, tc.skillNames)

			repo := &ResumeRepository{DB: db}
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
func TestResumeRepository_FindSpecializationIDByName(t *testing.T) {
	t.Parallel()

	querySelect := regexp.QuoteMeta(`SELECT id FROM specialization WHERE name = $1`)
	queryInsert := regexp.QuoteMeta(`INSERT INTO specialization (name) VALUES ($1) RETURNING id`)

	testCases := []struct {
		name               string
		specializationName string
		expectedID         int
		expectedErr        error
		setupMock          func(mock sqlmock.Sqlmock, specializationName string)
	}{
		{
			name:               "Успешное получение ID для существующей специализации",
			specializationName: "Backend Developer",
			expectedID:         1,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name:               "Успешное создание новой специализации",
			specializationName: "Data Scientist",
			expectedID:         2,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				// Simulate specialization not found
				mock.ExpectQuery(querySelect).
					WithArgs(specializationName).
					WillReturnError(sql.ErrNoRows)
				// Simulate specialization creation
				mock.ExpectQuery(queryInsert).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
		},
		{
			name:               "Ошибка при проверке существования специализации",
			specializationName: "Frontend Developer",
			expectedID:         0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования специализации: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(specializationName).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:               "Ошибка при создании новой специализации",
			specializationName: "DevOps Engineer",
			expectedID:         0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании специализации: %w", errors.New("insert error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(specializationName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(queryInsert).
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
			defer db.Close()

			tc.setupMock(mock, tc.specializationName)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.FindSpecializationIDByName(ctx, tc.specializationName)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Equal(t, tc.expectedID, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestResumeRepository_FindSpecializationIDsByNames(t *testing.T) {
	t.Parallel()

	querySelect := regexp.QuoteMeta(`SELECT id FROM specialization WHERE name = $1`)
	queryInsert := regexp.QuoteMeta(`INSERT INTO specialization (name) VALUES ($1) RETURNING id`)

	testCases := []struct {
		name                string
		specializationNames []string
		expectedResult      []int
		expectedErr         error
		setupMock           func(mock sqlmock.Sqlmock, specializationNames []string)
	}{
		{
			name:                "Пустой список специализаций",
			specializationNames: []string{},
			expectedResult:      []int{},
			expectedErr:         nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationNames []string) {
				// No database queries expected for empty input
			},
		},
		{
			name:                "Успешное получение ID для существующих специализаций",
			specializationNames: []string{"Backend Developer", "Frontend Developer"},
			expectedResult:      []int{1, 2},
			expectedErr:         nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationNames []string) {
				for i, name := range specializationNames {
					mock.ExpectQuery(querySelect).
						WithArgs(name).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 1))
				}
			},
		},
		{
			name:                "Успешное создание новых специализаций",
			specializationNames: []string{"Data Scientist", "DevOps Engineer"},
			expectedResult:      []int{3, 4},
			expectedErr:         nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationNames []string) {
				for i, name := range specializationNames {
					// Simulate specialization not found
					mock.ExpectQuery(querySelect).
						WithArgs(name).
						WillReturnError(sql.ErrNoRows)
					// Simulate specialization creation
					mock.ExpectQuery(queryInsert).
						WithArgs(name).
						WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 3))
				}
			},
		},
		{
			name:                "Смешанный случай: существующие и новые специализации",
			specializationNames: []string{"Backend Developer", "Machine Learning Engineer"},
			expectedResult:      []int{1, 5},
			expectedErr:         nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationNames []string) {
				// First specialization exists
				mock.ExpectQuery(querySelect).
					WithArgs("Backend Developer").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				// Second specialization does not exist
				mock.ExpectQuery(querySelect).
					WithArgs("Machine Learning Engineer").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(queryInsert).
					WithArgs("Machine Learning Engineer").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
			},
		},
		{
			name:                "Ошибка при проверке существования специализации",
			specializationNames: []string{"Backend Developer"},
			expectedResult:      nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования специализации: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationNames []string) {
				mock.ExpectQuery(querySelect).
					WithArgs("Backend Developer").
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:                "Ошибка при создании новой специализации",
			specializationNames: []string{"Cloud Architect"},
			expectedResult:      nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании специализации: %w", errors.New("insert error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationNames []string) {
				mock.ExpectQuery(querySelect).
					WithArgs("Cloud Architect").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(queryInsert).
					WithArgs("Cloud Architect").
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
			defer db.Close()

			tc.setupMock(mock, tc.specializationNames)

			repo := &ResumeRepository{DB: db}
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

func TestResumeRepository_CreateSkillIfNotExists(t *testing.T) {
	t.Parallel()

	querySelect := regexp.QuoteMeta(`SELECT id FROM skill WHERE name = $1`)
	queryInsert := regexp.QuoteMeta(`INSERT INTO skill (name) VALUES ($1) RETURNING id`)

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
				mock.ExpectQuery(querySelect).
					WithArgs(skillName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name:        "Навык не существует, успешно создан",
			skillName:   "Python",
			expectedID:  2,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				// Skill does not exist
				mock.ExpectQuery(querySelect).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				// Skill creation
				mock.ExpectQuery(queryInsert).
					WithArgs(skillName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
		},
		{
			name:       "Ошибка при проверке существования навыка",
			skillName:  "SQL",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования навыка: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(skillName).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:       "Ошибка при создании навыка (не уникальное нарушение)",
			skillName:  "Docker",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании навыка: %w", errors.New("insert error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(queryInsert).
					WithArgs(skillName).
					WillReturnError(errors.New("insert error"))
			},
		},
		{
			name:        "Уникальное нарушение, успешное получение ID",
			skillName:   "Kubernetes",
			expectedID:  3,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(queryInsert).
					WithArgs(skillName).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectQuery(querySelect).
					WithArgs(skillName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
			},
		},
		{
			name:       "Уникальное нарушение, ошибка при повторном получении ID",
			skillName:  "Rust",
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении ID навыка после конфликта: %w", errors.New("retry error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, skillName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(skillName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(queryInsert).
					WithArgs(skillName).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectQuery(querySelect).
					WithArgs(skillName).
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
			defer db.Close()

			tc.setupMock(mock, tc.skillName)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.CreateSkillIfNotExists(ctx, tc.skillName)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Equal(t, tc.expectedID, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestResumeRepository_CreateSpecializationIfNotExists(t *testing.T) {
	t.Parallel()

	querySelect := regexp.QuoteMeta(`SELECT id FROM specialization WHERE name = $1`)
	queryInsert := regexp.QuoteMeta(`INSERT INTO specialization (name) VALUES ($1) RETURNING id`)

	testCases := []struct {
		name               string
		specializationName string
		expectedID         int
		expectedErr        error
		setupMock          func(mock sqlmock.Sqlmock, specializationName string)
	}{
		{
			name:               "Специализация уже существует",
			specializationName: "Backend Developer",
			expectedID:         1,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name:               "Специализация не существует, успешно создана",
			specializationName: "Data Scientist",
			expectedID:         2,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				// Specialization does not exist
				mock.ExpectQuery(querySelect).
					WithArgs(specializationName).
					WillReturnError(sql.ErrNoRows)
				// Specialization creation
				mock.ExpectQuery(queryInsert).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			},
		},
		{
			name:               "Ошибка при проверке существования специализации",
			specializationName: "Frontend Developer",
			expectedID:         0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке существования специализации: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(specializationName).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:               "Ошибка при создании специализации (не уникальное нарушение)",
			specializationName: "DevOps Engineer",
			expectedID:         0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании специализации: %w", errors.New("insert error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(specializationName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(queryInsert).
					WithArgs(specializationName).
					WillReturnError(errors.New("insert error"))
			},
		},
		{
			name:               "Уникальное нарушение, успешное получение ID",
			specializationName: "Machine Learning Engineer",
			expectedID:         3,
			expectedErr:        nil,
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(specializationName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(queryInsert).
					WithArgs(specializationName).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectQuery(querySelect).
					WithArgs(specializationName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
			},
		},
		{
			name:               "Уникальное нарушение, ошибка при повторном получении ID",
			specializationName: "Cloud Architect",
			expectedID:         0,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении ID специализации после конфликта: %w", errors.New("retry error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, specializationName string) {
				mock.ExpectQuery(querySelect).
					WithArgs(specializationName).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(queryInsert).
					WithArgs(specializationName).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectQuery(querySelect).
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
			defer db.Close()

			tc.setupMock(mock, tc.specializationName)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.CreateSpecializationIfNotExists(ctx, tc.specializationName)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Equal(t, tc.expectedID, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestResumeRepository_SearchResumesByProfession(t *testing.T) {
	t.Parallel()

	graduationDate := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year",
		"profession", "created_at", "updated_at",
	}

	query := regexp.QuoteMeta(`
        SELECT id, applicant_id, about_me, specialization_id, education, 
               educational_institution, graduation_year, profession, created_at, updated_at
        FROM resume
        WHERE profession ILIKE $1
        ORDER BY updated_at DESC
        LIMIT $2 OFFSET $3
    `)

	testCases := []struct {
		name           string
		profession     string
		limit          int
		offset         int
		expectedResult []entity.Resume
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, profession string, limit, offset int)
	}{
		{
			name:       "Успешное получение списка резюме",
			profession: "Developer",
			limit:      2,
			offset:     0,
			expectedResult: []entity.Resume{
				{
					ID:                     1,
					ApplicantID:            1,
					AboutMe:                "Опытный разработчик",
					SpecializationID:       2,
					Education:              entity.Higher,
					EducationalInstitution: "МГУ",
					GraduationYear:         graduationDate,
					Profession:             "Backend Developer",
					CreatedAt:              now,
					UpdatedAt:              now,
				},
				{
					ID:                     2,
					ApplicantID:            2,
					AboutMe:                "Начинающий разработчик",
					SpecializationID:       3,
					Education:              entity.SecondarySchool,
					EducationalInstitution: "Школа №123",
					GraduationYear:         graduationDate,
					Profession:             "Frontend Developer",
					CreatedAt:              now,
					UpdatedAt:              now,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, profession string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						1,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Backend Developer",
						now,
						now,
					).
					AddRow(
						2,
						2,
						"Начинающий разработчик",
						3,
						string(entity.SecondarySchool),
						"Школа №123",
						graduationDate,
						"Frontend Developer",
						now,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs("%"+profession+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Пустой список резюме",
			profession:     "Analyst",
			limit:          10,
			offset:         0,
			expectedResult: []entity.Resume{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, profession string, limit, offset int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs("%"+profession+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка базы данных при выполнении запроса",
			profession:     "Developer",
			limit:          10,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске резюме по профессии: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, profession string, limit, offset int) {
				mock.ExpectQuery(query).
					WithArgs("%"+profession+"%", limit, offset).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка сканирования строк",
			profession:     "Developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по резюме: %w", errors.New("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, profession string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						1,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Backend Developer",
						now,
						now,
					).
					RowError(0, errors.New("scan error"))
				mock.ExpectQuery(query).
					WithArgs("%"+profession+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка итерации по строкам",
			profession:     "Developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по резюме: %w", errors.New("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, profession string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						1,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Backend Developer",
						now,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs("%"+profession+"%", limit, offset).
					WillReturnRows(rows)
				rows.CloseError(errors.New("rows error"))
			},
		},
		{
			name:           "Ошибка закрытия строк",
			profession:     "Developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по резюме: %w", errors.New("close error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, profession string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						1,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Backend Developer",
						now,
						now,
					)
				rows.CloseError(errors.New("close error"))
				mock.ExpectQuery(query).
					WithArgs("%"+profession+"%", limit, offset).
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

			tc.setupMock(mock, tc.profession, tc.limit, tc.offset)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.SearchResumesByProfession(ctx, tc.profession, tc.limit, tc.offset)

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
					require.Equal(t, expected.ApplicantID, result[i].ApplicantID)
					require.Equal(t, expected.AboutMe, result[i].AboutMe)
					require.Equal(t, expected.SpecializationID, result[i].SpecializationID)
					require.Equal(t, expected.Education, result[i].Education)
					require.Equal(t, expected.EducationalInstitution, result[i].EducationalInstitution)
					require.Equal(t, expected.GraduationYear, result[i].GraduationYear)
					require.Equal(t, expected.Profession, result[i].Profession)
					require.False(t, result[i].CreatedAt.IsZero())
					require.False(t, result[i].UpdatedAt.IsZero())
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestResumeRepository_SearchResumesByProfessionForApplicant(t *testing.T) {
	t.Parallel()

	graduationDate := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	columns := []string{
		"id", "applicant_id", "about_me", "specialization_id",
		"education", "educational_institution", "graduation_year",
		"profession", "created_at", "updated_at",
	}

	query := regexp.QuoteMeta(`
        SELECT id, applicant_id, about_me, specialization_id, education, 
               educational_institution, graduation_year, profession, created_at, updated_at
        FROM resume
        WHERE applicant_id = $1 AND profession ILIKE $2
        ORDER BY updated_at DESC
        LIMIT $3 OFFSET $4
    `)

	testCases := []struct {
		name           string
		applicantID    int
		profession     string
		limit          int
		offset         int
		expectedResult []entity.Resume
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, applicantID int, profession string, limit, offset int)
	}{
		{
			name:        "Успешное получение списка резюме",
			applicantID: 1,
			profession:  "Developer",
			limit:       2,
			offset:      0,
			expectedResult: []entity.Resume{
				{
					ID:                     1,
					ApplicantID:            1,
					AboutMe:                "Опытный разработчик",
					SpecializationID:       2,
					Education:              entity.Higher,
					EducationalInstitution: "МГУ",
					GraduationYear:         graduationDate,
					Profession:             "Backend Developer",
					CreatedAt:              now,
					UpdatedAt:              now,
				},
				{
					ID:                     2,
					ApplicantID:            1,
					AboutMe:                "Начинающий разработчик",
					SpecializationID:       3,
					Education:              entity.SecondarySchool,
					EducationalInstitution: "Школа №123",
					GraduationYear:         graduationDate,
					Profession:             "Frontend Developer",
					CreatedAt:              now,
					UpdatedAt:              now,
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, applicantID int, profession string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						applicantID,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Backend Developer",
						now,
						now,
					).
					AddRow(
						2,
						applicantID,
						"Начинающий разработчик",
						3,
						string(entity.SecondarySchool),
						"Школа №123",
						graduationDate,
						"Frontend Developer",
						now,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs(applicantID, "%"+profession+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Пустой список резюме",
			applicantID:    1,
			profession:     "Analyst",
			limit:          10,
			offset:         0,
			expectedResult: []entity.Resume{},
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, applicantID int, profession string, limit, offset int) {
				rows := sqlmock.NewRows(columns)
				mock.ExpectQuery(query).
					WithArgs(applicantID, "%"+profession+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка базы данных при выполнении запроса",
			applicantID:    1,
			profession:     "Developer",
			limit:          10,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске резюме по профессии для соискателя: %w", errors.New("database error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicantID int, profession string, limit, offset int) {
				mock.ExpectQuery(query).
					WithArgs(applicantID, "%"+profession+"%", limit, offset).
					WillReturnError(errors.New("database error"))
			},
		},
		{
			name:           "Ошибка сканирования строк",
			applicantID:    1,
			profession:     "Developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по резюме: %w", errors.New("scan error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicantID int, profession string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						applicantID,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Backend Developer",
						now,
						now,
					).
					RowError(0, errors.New("scan error"))
				mock.ExpectQuery(query).
					WithArgs(applicantID, "%"+profession+"%", limit, offset).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка итерации по строкам",
			applicantID:    1,
			profession:     "Developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по резюме: %w", errors.New("rows error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicantID int, profession string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						applicantID,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Backend Developer",
						now,
						now,
					)
				mock.ExpectQuery(query).
					WithArgs(applicantID, "%"+profession+"%", limit, offset).
					WillReturnRows(rows)
				rows.CloseError(errors.New("rows error"))
			},
		},
		{
			name:           "Ошибка закрытия строк",
			applicantID:    1,
			profession:     "Developer",
			limit:          2,
			offset:         0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при итерации по резюме: %w", errors.New("close error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicantID int, profession string, limit, offset int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1,
						applicantID,
						"Опытный разработчик",
						2,
						string(entity.Higher),
						"МГУ",
						graduationDate,
						"Backend Developer",
						now,
						now,
					)
				rows.CloseError(errors.New("close error"))
				mock.ExpectQuery(query).
					WithArgs(applicantID, "%"+profession+"%", limit, offset).
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

			tc.setupMock(mock, tc.applicantID, tc.profession, tc.limit, tc.offset)

			repo := &ResumeRepository{DB: db}
			ctx := context.Background()

			result, err := repo.SearchResumesByProfessionForApplicant(ctx, tc.applicantID, tc.profession, tc.limit, tc.offset)

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
					require.Equal(t, expected.ApplicantID, result[i].ApplicantID)
					require.Equal(t, expected.AboutMe, result[i].AboutMe)
					require.Equal(t, expected.SpecializationID, result[i].SpecializationID)
					require.Equal(t, expected.Education, result[i].Education)
					require.Equal(t, expected.EducationalInstitution, result[i].EducationalInstitution)
					require.Equal(t, expected.GraduationYear, result[i].GraduationYear)
					require.Equal(t, expected.Profession, result[i].Profession)
					require.False(t, result[i].CreatedAt.IsZero())
					require.False(t, result[i].UpdatedAt.IsZero())
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
