package postgres

import (
	"ResuMatch/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	// "regexp"
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
        INSERT INTO applicant \(email, password_hashed, password_salt, first_name, last_name\)
        VALUES \(\$1, \$2, \$3, \$4, \$5\)
        RETURNING id, email, password_hashed, password_salt, first_name, last_name
    `

	testCases := []struct {
		name           string
		email          string
		firstName      string
		lastName       string
		passwordHash   []byte
		passwordSalt   []byte
		expectedResult *entity.Applicant
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock)
	}{
		{
			name:         "Успешное создание соискателя",
			email:        "test@example.com",
			firstName:    "Николай",
			lastName:     "Иванов",
			passwordHash: []byte("hash"),
			passwordSalt: []byte("salt"),
			expectedResult: createTestApplicant(1, "test@example.com", "Николай", "Иванов",
				[]byte("hash"), []byte("salt")),
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(testQuery).
					WithArgs("test@example.com", []byte("hash"), []byte("salt"), "Николай", "Иванов").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hashed", "password_salt", "first_name", "last_name"}).
						AddRow(1, "test@example.com", []byte("hash"), []byte("salt"), "Николай", "Иванов"))
				mock.ExpectClose()
			},
		},
		{
			name:           "Email уже занят",
			email:          "existing@example.com",
			firstName:      "Николай",
			lastName:       "Иванов",
			passwordHash:   []byte("hash"),
			passwordSalt:   []byte("salt"),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				errors.New("соискатель с таким email уже зарегистрирован"),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(testQuery).
					WithArgs("existing@example.com", []byte("hash"), []byte("salt"), "Николай", "Иванов").
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectClose()
			},
		},
		{
			name:           "Отсутствует обязательное поле",
			email:          "",
			firstName:      "Николай",
			lastName:       "Иванов",
			passwordHash:   []byte("hash"),
			passwordSalt:   []byte("salt"),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				errors.New("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(testQuery).
					WithArgs("", []byte("hash"), []byte("salt"), "Николай", "Иванов").
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
				mock.ExpectClose()
			},
		},
		{
			name:           "Проверка CHECK ограничения",
			email:          "existing@example.com",
			firstName:      "Николай",
			lastName:       "Иванов",
			passwordHash:   []byte("hash"),
			passwordSalt:   []byte("salt"),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(testQuery).
					WithArgs("existing@example.com", []byte("hash"), []byte("salt"), "Николай", "Иванов").
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
				mock.ExpectClose()
			},
		},
		{
			name:           "Неверный формат данных",
			email:          "@user.mail.ru",
			firstName:      "Николай",
			lastName:       "Иванов",
			passwordHash:   []byte("очень много байтов для хеша пароля, так что будет ошибка..."),
			passwordSalt:   []byte("salt"),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(testQuery).
					WithArgs("@user.mail.ru", []byte("очень много байтов для хеша пароля, так что будет ошибка..."), []byte("salt"), "Николай", "Иванов").
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
				mock.ExpectClose()
			},
		},
		{
			name:           "Неизвестная ошибка сервера",
			email:          "@user.mail.ru",
			firstName:      "Николай",
			lastName:       "Иванов",
			passwordHash:   []byte("hash"),
			passwordSalt:   []byte("salt"),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				errors.New("неизвестная ошибка при создании соискателя err=pq: test pq error"),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(testQuery).
					WithArgs("@user.mail.ru", []byte("hash"), []byte("salt"), "Николай", "Иванов").
					WillReturnError(&pq.Error{
						Code:    "12345",
						Message: "test pq error",
					})
				mock.ExpectClose()
			},
		},
		{
			name:           "Обычная ошибка (не PostgreSQL)",
			email:          "test@example.com",
			firstName:      "Николай",
			lastName:       "Иванов",
			passwordHash:   []byte("hash"),
			passwordSalt:   []byte("salt"),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				errors.New("ошибка при создании соискателя: test non-pq error"),
			),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(testQuery).
					WithArgs("test@example.com", []byte("hash"), []byte("salt"), "Николай", "Иванов").
					WillReturnError(errors.New("test non-pq error"))
				mock.ExpectClose()
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			repo := &ApplicantRepository{DB: db}

			tc.setupMock(mock)

			ctx := context.Background()
			result, err := repo.CreateApplicant(
				ctx,
				tc.email,
				tc.firstName,
				tc.lastName,
				tc.passwordHash,
				tc.passwordSalt,
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

			require.NoError(t, db.Close())
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
        FROM applicant WHERE id = \$1
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
		setupMock      func(mock sqlmock.Sqlmock)
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
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(5).
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
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(5).
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
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(777).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(5).
					WillReturnError(errors.New("database connection error"))
				mock.ExpectClose()
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			repo := &ApplicantRepository{DB: db}

			tc.setupMock(mock)

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

			require.NoError(t, db.Close())
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
        FROM applicant WHERE email = \$1
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
		setupMock      func(mock sqlmock.Sqlmock)
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
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs("ivan@example.com").
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
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs("ivan@example.com").
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
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs("unknown@example.com").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs("ivan@example.com").
					WillReturnError(errors.New("database connection error"))
				mock.ExpectClose()
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			repo := &ApplicantRepository{DB: db}

			tc.setupMock(mock)

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

			require.NoError(t, db.Close())
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
		setupMock   func(mock sqlmock.Sqlmock)
	}{

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
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `UPDATE applicant SET first_name = \$1 WHERE id = \$2`
				mock.ExpectExec(query).
					WithArgs(nil, 3).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `UPDATE applicant SET birth_date = \$1 WHERE id = \$2`
				mock.ExpectExec(query).
					WithArgs("29 августа", 3).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `UPDATE applicant SET first_name = \$1 WHERE id = \$2`
				mock.ExpectExec(query).
					WithArgs(strings.Repeat("a", 50), 1).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `UPDATE applicant SET vk = \$1 WHERE id = \$2`
				mock.ExpectExec(query).
					WithArgs("https://vk.com/existing", 1).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `UPDATE applicant SET quote = \$1 WHERE id = \$2`
				mock.ExpectExec(query).
					WithArgs("Новая цитата", 1).
					WillReturnError(&pq.Error{
						Code:    "99999",
						Message: "unknown error",
					})
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `UPDATE applicant SET telegram = \$1 WHERE id = \$2`
				mock.ExpectExec(query).
					WithArgs("@newtelegram", 8).
					WillReturnError(errors.New("connection error"))
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `UPDATE applicant SET first_name = \$1 WHERE id = \$2`
				mock.ExpectExec(query).
					WithArgs("Артем", 1).
					WillReturnResult(&mockResultWithError{})
				mock.ExpectClose()
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
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `UPDATE applicant SET quote = \$1 WHERE id = \$2`
				mock.ExpectExec(query).
					WithArgs("Не обновилось", 1).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectClose()
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			repo := &ApplicantRepository{DB: db}

			tc.setupMock(mock)

			ctx := context.Background()
			err = repo.UpdateApplicant(ctx, tc.userID, tc.fields)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, db.Close())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
