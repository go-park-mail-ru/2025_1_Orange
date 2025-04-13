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
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type mockResultWithError struct{}

func (m *mockResultWithError) LastInsertId() (int64, error) {
	return 0, nil
}

func (m *mockResultWithError) RowsAffected() (int64, error) {
	return 0, errors.New("ошибка rowsAffected")
}

func TestApplicantRepository_CreateApplicant(t *testing.T) {
	t.Parallel()

	createTestApplicant := func(id int, email, firstName, lastName string, hash, salt []byte) *entity.Applicant {
		return &entity.Applicant{
			ID:           id,
			Email:        email,
			FirstName:    firstName,
			LastName:     lastName,
			PasswordHash: hash,
			PasswordSalt: salt,
		}
	}

	testQuery := `
        INSERT INTO applicant (email, password_hashed, password_salt, first_name, last_name)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, email, password_hashed, password_salt, first_name, last_name
	`

	testCases := []struct {
		name           string
		applicant      *entity.Applicant
		expectedResult *entity.Applicant
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, applicant *entity.Applicant, query string)
	}{
		{
			name:           "Успешное создание соискателя",
			applicant:      createTestApplicant(1, "test@example.com", "Николай", "Иванов", []byte("hash"), []byte("salt")),
			expectedResult: createTestApplicant(1, "test@example.com", "Николай", "Иванов", []byte("hash"), []byte("salt")),
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, applicant *entity.Applicant, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(applicant.Email, applicant.PasswordHash, applicant.PasswordSalt, applicant.FirstName, applicant.LastName).
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hashed", "password_salt", "first_name", "last_name"}).
						AddRow(1, applicant.Email, applicant.PasswordHash, applicant.PasswordSalt, applicant.FirstName, applicant.LastName))
			},
		},
		{
			name:           "Email уже занят",
			applicant:      createTestApplicant(1, "existing@example.com", "Николай", "Иванов", []byte("hash"), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				errors.New("соискатель с таким email уже зарегистрирован"),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicant *entity.Applicant, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(applicant.Email, applicant.PasswordHash, applicant.PasswordSalt, applicant.FirstName, applicant.LastName).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
			},
		},
		{
			name:           "Отсутствует обязательное поле",
			applicant:      createTestApplicant(1, "", "Николай", "Иванов", []byte("hash"), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				errors.New("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicant *entity.Applicant, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(applicant.Email, applicant.PasswordHash, applicant.PasswordSalt, applicant.FirstName, applicant.LastName).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
			},
		},
		{
			name:           "Проверка CHECK ограничения",
			applicant:      createTestApplicant(1, "existing@example.com", "Николай", "Иванов", []byte("hash"), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicant *entity.Applicant, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(applicant.Email, applicant.PasswordHash, applicant.PasswordSalt, applicant.FirstName, applicant.LastName).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
			},
		},
		{
			name:           "Неверный формат данных",
			applicant:      createTestApplicant(1, "@user.mail.ru", "Николай", "Иванов", []byte("очень много байтов для хеша пароля, так что будет ошибка..."), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicant *entity.Applicant, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(applicant.Email, applicant.PasswordHash, applicant.PasswordSalt, applicant.FirstName, applicant.LastName).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
			},
		},
		{
			name:           "Неизвестная ошибка сервера",
			applicant:      createTestApplicant(1, "@user.mail.ru", "Николай", "Иванов", []byte("hash"), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				errors.New("неизвестная ошибка при создании соискателя err=pq: test pq error"),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicant *entity.Applicant, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(applicant.Email, applicant.PasswordHash, applicant.PasswordSalt, applicant.FirstName, applicant.LastName).
					WillReturnError(&pq.Error{
						Code:    "12345",
						Message: "test pq error",
					})
			},
		},
		{
			name:           "Обычная ошибка (не PostgreSQL)",
			applicant:      createTestApplicant(1, "test@example.com", "Николай", "Иванов", []byte("hash"), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				errors.New("ошибка при создании соискателя: test non-pq error"),
			),
			setupMock: func(mock sqlmock.Sqlmock, applicant *entity.Applicant, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(applicant.Email, applicant.PasswordHash, applicant.PasswordSalt, applicant.FirstName, applicant.LastName).
					WillReturnError(errors.New("test non-pq error"))
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

			repo := &ApplicantRepository{DB: db}

			tc.setupMock(mock, tc.applicant, testQuery)

			ctx := context.Background()
			result, err := repo.CreateApplicant(
				ctx,
				tc.applicant.Email,
				tc.applicant.FirstName,
				tc.applicant.LastName,
				tc.applicant.PasswordHash,
				tc.applicant.PasswordSalt,
			)

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

func TestApplicantRepository_GetApplicantByID(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2025, 4, 2, 12, 0, 0, 0, time.UTC)

	query := `
		SELECT id, first_name, last_name, middle_name, city_id, 
		       birth_date, sex, email, status, quote, vk,
		       telegram, facebook, avatar_id,
		       password_hashed, password_salt, created_at, updated_at
		FROM applicant WHERE id = $1
	`

	columns := []string{
		"id", "first_name", "last_name", "middle_name", "city_id",
		"birth_date", "sex", "email", "status", "quote", "vk",
		"telegram", "facebook", "avatar_id",
		"password_hashed", "password_salt", "created_at", "updated_at",
	}

	testCases := []struct {
		name           string
		id             int
		expectedResult *entity.Applicant
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name: "Успешное получение соискателя по ID",
			id:   5,
			expectedResult: &entity.Applicant{
				ID:           5,
				FirstName:    "Николай",
				LastName:     "Иванов",
				Email:        "ivan@example.com",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
				PasswordHash: []byte("hash"),
				PasswordSalt: []byte("salt"),
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows(columns).AddRow(
						5,
						"Николай",
						"Иванов",
						sql.NullString{},
						sql.NullInt64{},
						sql.NullTime{},
						sql.NullString{},
						"ivan@example.com",
						sql.NullString{},
						sql.NullString{},
						sql.NullString{},
						sql.NullString{},
						sql.NullString{},
						sql.NullInt64{},
						[]byte("hash"),
						[]byte("salt"),
						sql.NullTime{Time: fixedTime, Valid: true},
						sql.NullTime{Time: fixedTime, Valid: true},
					))
			},
		},
		{
			name: "Корректный статус",
			id:   5,
			expectedResult: &entity.Applicant{
				ID:           5,
				FirstName:    "Николай",
				LastName:     "Иванов",
				Email:        "ivan@example.com",
				Status:       entity.StatusActivelySearching,
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
				PasswordHash: []byte("hash"),
				PasswordSalt: []byte("salt"),
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows(columns).AddRow(
						5,
						"Николай",
						"Иванов",
						sql.NullString{},
						sql.NullInt64{},
						sql.NullTime{},
						sql.NullString{},
						"ivan@example.com",
						"actively_searching",
						sql.NullString{},
						sql.NullString{},
						sql.NullString{},
						sql.NullString{},
						sql.NullInt64{},
						[]byte("hash"),
						[]byte("salt"),
						sql.NullTime{Time: fixedTime, Valid: true},
						sql.NullTime{Time: fixedTime, Valid: true},
					))
			},
		},
		{
			name:           "Соискатель не найден по ID",
			id:             777,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("соискатель с id=777 не найден"),
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
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить соискателя по id=5"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnError(errors.New("database connection error"))
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

			repo := &ApplicantRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetApplicantByID(ctx, tc.id)

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

func TestApplicantRepository_GetApplicantByEmail(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2025, 4, 2, 12, 0, 0, 0, time.UTC)

	query := `
		SELECT id, first_name, last_name, middle_name, city_id, 
		       birth_date, sex, email, status, quote, vk,
		       telegram, facebook, avatar_id,
		       password_hashed, password_salt, created_at, updated_at
		FROM applicant WHERE email = $1
	`

	columns := []string{
		"id", "first_name", "last_name", "middle_name", "city_id",
		"birth_date", "sex", "email", "status", "quote", "vk",
		"telegram", "facebook", "avatar_id",
		"password_hashed", "password_salt", "created_at", "updated_at",
	}

	testCases := []struct {
		name           string
		email          string
		expectedResult *entity.Applicant
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, email string)
	}{
		{
			name:  "Успешное получение соискателя по Email",
			email: "ivan@example.com",
			expectedResult: &entity.Applicant{
				ID:           5,
				FirstName:    "Николай",
				LastName:     "Иванов",
				Email:        "ivan@example.com",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
				PasswordHash: []byte("hash"),
				PasswordSalt: []byte("salt"),
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(email).
					WillReturnRows(sqlmock.NewRows(columns).AddRow(
						5,
						"Николай",
						"Иванов",
						sql.NullString{},
						sql.NullInt64{},
						sql.NullTime{},
						sql.NullString{},
						"ivan@example.com",
						sql.NullString{},
						sql.NullString{},
						sql.NullString{},
						sql.NullString{},
						sql.NullString{},
						sql.NullInt64{},
						[]byte("hash"),
						[]byte("salt"),
						sql.NullTime{Time: fixedTime, Valid: true},
						sql.NullTime{Time: fixedTime, Valid: true},
					))
			},
		},
		{
			name:  "Корректный статус",
			email: "ivan@example.com",
			expectedResult: &entity.Applicant{
				ID:           5,
				FirstName:    "Николай",
				LastName:     "Иванов",
				Email:        "ivan@example.com",
				Status:       entity.StatusActivelySearching,
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
				PasswordHash: []byte("hash"),
				PasswordSalt: []byte("salt"),
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(email).
					WillReturnRows(sqlmock.NewRows(columns).AddRow(
						5,
						"Николай",
						"Иванов",
						sql.NullString{},
						sql.NullInt64{},
						sql.NullTime{},
						sql.NullString{},
						"ivan@example.com",
						"actively_searching",
						sql.NullString{},
						sql.NullString{},
						sql.NullString{},
						sql.NullString{},
						sql.NullInt64{},
						[]byte("hash"),
						[]byte("salt"),
						sql.NullTime{Time: fixedTime, Valid: true},
						sql.NullTime{Time: fixedTime, Valid: true},
					))
			},
		},
		{
			name:           "Соискатель не найден по Email",
			email:          "unknown@example.com",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("соискатель с email=unknown@example.com не найден"),
			),
			setupMock: func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(email).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка PostgreSQL",
			email:          "ivan@example.com",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось найти соискателя с email=ivan@example.com"),
			),
			setupMock: func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(email).
					WillReturnError(errors.New("database connection error"))
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

			tc.setupMock(mock, tc.email)

			repo := &ApplicantRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetApplicantByEmail(ctx, tc.email)

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

func TestApplicantRepository_UpdateApplicant(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		userID      int
		fields      map[string]interface{}
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{})
	}{
		{
			name:   "Успешное обновление информации соискателя",
			userID: 1,
			fields: map[string]interface{}{
				"first_name": "Николай",
				"last_name":  "Петров",
				"quote":      "Новый статус",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE applicant SET first_name = $1, last_name = $2, quote = $3 WHERE id = $4"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["first_name"],
						fields["last_name"],
						fields["quote"],
						userID,
					).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:   "Ошибка NOT NULL для обязательного поля",
			userID: 3,
			fields: map[string]interface{}{
				"first_name": nil,
			},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE applicant SET first_name = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["first_name"],
						userID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
			},
		},

		{
			name:   "Ошибка DATA TYPE",
			userID: 3,
			fields: map[string]interface{}{
				"birth_date": "29 августа",
			},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE applicant SET birth_date = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["birth_date"],
						userID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
			},
		},
		{
			name:   "Нарушение CHECK ограничения длины имени",
			userID: 1,
			fields: map[string]interface{}{
				"first_name": strings.Repeat("a", 50),
			},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE applicant SET first_name = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["first_name"],
						userID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
			},
		},
		{
			name:   "Нарушение уникальности VK",
			userID: 1,
			fields: map[string]interface{}{
				"vk": "https://vk.com/existing",
			},
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("ошибка уникальности"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE applicant SET vk = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["vk"],
						userID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
			},
		},
		{
			name:   "Неизвестная ошибка PostgreSQL",
			userID: 1,
			fields: map[string]interface{}{
				"quote": "Новая цитата",
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("неизвестная ошибка при обновлении соискателя err=pq: unknown error"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE applicant SET quote = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["quote"],
						userID,
					).
					WillReturnError(&pq.Error{
						Code:    "99999",
						Message: "unknown error",
					})
			},
		},
		{
			name:   "Обычная ошибка (не PostgreSQL)",
			userID: 8,
			fields: map[string]interface{}{
				"telegram": "@newtelegram",
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось обновить соискателя с id=8"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE applicant SET telegram = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["telegram"],
						userID,
					).
					WillReturnError(errors.New("connection error"))
			},
		},
		{
			name:   "Ошибка при вызове RowsAffected()",
			userID: 1,
			fields: map[string]interface{}{
				"first_name": "Артем",
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить обновленные строки при обновлении соискателя с id=1"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE applicant SET first_name = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["first_name"],
						userID,
					).
					WillReturnResult(&mockResultWithError{})
			},
		},
		{
			name:   "Ничего не обновилось — rowsAffected == 0",
			userID: 1,
			fields: map[string]interface{}{
				"quote": "Не обновилось",
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось найти при обновлении соискателя с id=1"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE applicant SET quote = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["quote"],
						userID,
					).
					WillReturnResult(sqlmock.NewResult(0, 0))
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

			repo := &ApplicantRepository{DB: db}
			tc.setupMock(mock, tc.userID, tc.fields)

			ctx := context.Background()
			err = repo.UpdateApplicant(ctx, tc.userID, tc.fields)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
