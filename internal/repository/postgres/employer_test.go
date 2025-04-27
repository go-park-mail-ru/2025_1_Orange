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

func TestEmployerRepository_CreateEmployer(t *testing.T) {
	t.Parallel()

	createTestEmployer := func(id int, email, companyName, legalAddress string, hash, salt []byte) *entity.Employer {
		return &entity.Employer{
			ID:           id,
			Email:        email,
			CompanyName:  companyName,
			LegalAddress: legalAddress,
			PasswordHash: hash,
			PasswordSalt: salt,
		}
	}

	testQuery := `
        INSERT INTO employer (email, password_hashed, password_salt, company_name, legal_address)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, email, password_hashed, password_salt, company_name, legal_address
	`

	testCases := []struct {
		name           string
		employer       *entity.Employer
		expectedResult *entity.Employer
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, employer *entity.Employer, query string)
	}{
		{
			name:           "Успешное создание работодателя",
			employer:       createTestEmployer(1, "test@example.com", "Технопарк", "МГТУ им. Н.Э. Баумана", []byte("hash"), []byte("salt")),
			expectedResult: createTestEmployer(1, "test@example.com", "Технопарк", "МГТУ им. Н.Э. Баумана", []byte("hash"), []byte("salt")),
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, employer *entity.Employer, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(employer.Email, employer.PasswordHash, employer.PasswordSalt, employer.CompanyName, employer.LegalAddress).
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hashed", "password_salt", "first_name", "last_name"}).
						AddRow(1, employer.Email, employer.PasswordHash, employer.PasswordSalt, employer.CompanyName, employer.LegalAddress))
			},
		},
		{
			name:           "Email уже занят",
			employer:       createTestEmployer(1, "existing@example.com", "Технопарк", "МГТУ им. Н.Э. Баумана", []byte("hash"), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("работодатель с таким email уже зарегистрирован"),
			),
			setupMock: func(mock sqlmock.Sqlmock, employer *entity.Employer, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(employer.Email, employer.PasswordHash, employer.PasswordSalt, employer.CompanyName, employer.LegalAddress).
					WillReturnError(&pq.Error{Code: entity.PSQLUniqueViolation})
			},
		},
		{
			name:           "Отсутствует обязательное поле",
			employer:       createTestEmployer(1, "", "Технопарк", "МГТУ им. Н.Э. Баумана", []byte("hash"), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				errors.New("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, employer *entity.Employer, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(employer.Email, employer.PasswordHash, employer.PasswordSalt, employer.CompanyName, employer.LegalAddress).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
			},
		},
		{
			name:           "Проверка CHECK ограничения",
			employer:       createTestEmployer(1, "existing@example.com", "Технопарк", "МГТУ им. Н.Э. Баумана", []byte("hash"), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, employer *entity.Employer, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(employer.Email, employer.PasswordHash, employer.PasswordSalt, employer.CompanyName, employer.LegalAddress).
					WillReturnError(&pq.Error{Code: entity.PSQLCheckViolation})
			},
		},
		{
			name:           "Неверный формат данных",
			employer:       createTestEmployer(1, "@user.mail.ru", "Технопарк", "МГТУ им. Н.Э. Баумана", []byte("очень много байтов для хеша пароля, так что будет ошибка..."), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, employer *entity.Employer, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(employer.Email, employer.PasswordHash, employer.PasswordSalt, employer.CompanyName, employer.LegalAddress).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
			},
		},
		{
			name:           "Неизвестная ошибка сервера",
			employer:       createTestEmployer(1, "@user.mail.ru", "Технопарк", "МГТУ им. Н.Э. Баумана", []byte("hash"), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("неизвестная ошибка при создании работодателя err=pq: test pq error"),
			),
			setupMock: func(mock sqlmock.Sqlmock, employer *entity.Employer, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(employer.Email, employer.PasswordHash, employer.PasswordSalt, employer.CompanyName, employer.LegalAddress).
					WillReturnError(&pq.Error{
						Code:    "12345",
						Message: "test pq error",
					})
			},
		},
		{
			name:           "Обычная ошибка (не PostgreSQL)",
			employer:       createTestEmployer(1, "test@example.com", "Технопарк", "МГТУ им. Н.Э. Баумана", []byte("hash"), []byte("salt")),
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании работодателя: test non-pq error"),
			),
			setupMock: func(mock sqlmock.Sqlmock, employer *entity.Employer, query string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(employer.Email, employer.PasswordHash, employer.PasswordSalt, employer.CompanyName, employer.LegalAddress).
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

			defer func() {
				err := db.Close()
				require.NoError(t, err)
			}()

			repo := &EmployerRepository{DB: db}

			tc.setupMock(mock, tc.employer, testQuery)

			ctx := context.Background()
			result, err := repo.CreateEmployer(
				ctx,
				tc.employer.Email,
				tc.employer.CompanyName,
				tc.employer.LegalAddress,
				tc.employer.PasswordHash,
				tc.employer.PasswordSalt,
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

func TestEmployerRepository_GetEmployerByID(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2023, 2, 2, 1, 0, 0, 0, time.UTC)

	query := `
		SELECT id, email, password_hashed, password_salt, company_name,
		       legal_address, vk, telegram, facebook, slogan, 
		       website, description, logo_id, created_at, updated_at
		FROM employer
		WHERE id = $1
	`

	columns := []string{
		"id", "email", "password_hashed", "password_salt", "company_name",
		"legal_address", "vk", "telegram", "facebook", "slogan",
		"website", "description", "logo_id", "created_at", "updated_at",
	}

	testCases := []struct {
		name           string
		id             int
		expectedResult *entity.Employer
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name: "Успешное получение работодателя по ID",
			id:   1,
			expectedResult: &entity.Employer{
				ID:           1,
				Email:        "technopark_vk@mail.ru",
				CompanyName:  "Технопарк ВК",
				LegalAddress: "Москва, МГТУ им. Н.Э. Баумана",
				Vk:           "vk.com/technopark",
				Telegram:     "t.me/technopark",
				Facebook:     "fb.com/technopark",
				Slogan:       "Создаем крутых программистов",
				Website:      "https://technopark.com",
				Description:  "Образовательный центр ВК Технопарк",
				LogoID:       1,
				PasswordHash: []byte("hash123"),
				PasswordSalt: []byte("salt123"),
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows(columns).AddRow(
						1,
						"technopark_vk@mail.ru",
						[]byte("hash123"),
						[]byte("salt123"),
						"Технопарк ВК",
						"Москва, МГТУ им. Н.Э. Баумана",
						sql.NullString{String: "vk.com/technopark", Valid: true},
						sql.NullString{String: "t.me/technopark", Valid: true},
						sql.NullString{String: "fb.com/technopark", Valid: true},
						sql.NullString{String: "Создаем крутых программистов", Valid: true},
						sql.NullString{String: "https://technopark.com", Valid: true},
						sql.NullString{String: "Образовательный центр ВК Технопарк", Valid: true},
						sql.NullInt64{Int64: 1, Valid: true},
						sql.NullTime{Time: fixedTime, Valid: true},
						sql.NullTime{Time: fixedTime, Valid: true},
					))
			},
		},
		{
			name:           "Работодатель не найден по ID",
			id:             12345678,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("работодатель с id=12345678 не найден"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка PostgreSQL",
			id:             1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить работодателя по id=1"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnError(errors.New("unexpected connection error"))
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

			repo := &EmployerRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetEmployerByID(ctx, tc.id)

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

func TestEmployerRepository_GetEmployerByEmail(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2023, 2, 2, 1, 0, 0, 0, time.UTC)

	query := `
		SELECT id, email, password_hashed, password_salt, company_name,
		       legal_address, vk, telegram, facebook, slogan, 
		       website, description, logo_id, created_at, updated_at
		FROM employer
		WHERE email = $1
	`

	columns := []string{
		"id", "email", "password_hashed", "password_salt", "company_name",
		"legal_address", "vk", "telegram", "facebook", "slogan",
		"website", "description", "logo_id", "created_at", "updated_at",
	}

	testCases := []struct {
		name           string
		email          string
		expectedResult *entity.Employer
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, email string)
	}{
		{
			name:  "Успешное получение работодателя по Email",
			email: "technopark_vk@mail.ru",
			expectedResult: &entity.Employer{
				ID:           1,
				Email:        "technopark_vk@mail.ru",
				CompanyName:  "Технопарк ВК",
				LegalAddress: "Москва, МГТУ им. Н.Э. Баумана",
				Vk:           "vk.com/technopark",
				Telegram:     "t.me/technopark",
				Facebook:     "fb.com/technopark",
				Slogan:       "Создаем крутых программистов",
				Website:      "https://technopark.com",
				Description:  "Образовательный центр ВК Технопарк",
				LogoID:       1,
				PasswordHash: []byte("hash123"),
				PasswordSalt: []byte("salt123"),
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(email).
					WillReturnRows(sqlmock.NewRows(columns).AddRow(
						1,
						"technopark_vk@mail.ru",
						[]byte("hash123"),
						[]byte("salt123"),
						"Технопарк ВК",
						"Москва, МГТУ им. Н.Э. Баумана",
						sql.NullString{String: "vk.com/technopark", Valid: true},
						sql.NullString{String: "t.me/technopark", Valid: true},
						sql.NullString{String: "fb.com/technopark", Valid: true},
						sql.NullString{String: "Создаем крутых программистов", Valid: true},
						sql.NullString{String: "https://technopark.com", Valid: true},
						sql.NullString{String: "Образовательный центр ВК Технопарк", Valid: true},
						sql.NullInt64{Int64: 1, Valid: true},
						sql.NullTime{Time: fixedTime, Valid: true},
						sql.NullTime{Time: fixedTime, Valid: true},
					))
			},
		},
		{
			name:           "Работодатель не найден по Email",
			email:          "unknown@mail.ru",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("работодатель с email=unknown@mail.ru не найден"),
			),
			setupMock: func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(email).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка PostgreSQL",
			email:          "technopark_vk@mail.ru",
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось найти работодателя с email=technopark_vk@mail.ru"),
			),
			setupMock: func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(email).
					WillReturnError(errors.New("unexpected connection error"))
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

			tc.setupMock(mock, tc.email)

			repo := &EmployerRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetEmployerByEmail(ctx, tc.email)

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

func TestEmployerRepository_UpdateEmployer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		userID      int
		fields      map[string]interface{}
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{})
	}{
		{
			name:   "Успешное обновление информации работодателя",
			userID: 1,
			fields: map[string]interface{}{
				"slogan":        "Мужество, Воля, Труд и Упорства",
				"legal_address": "Москва, МГТУ им. Н.Э. Баумана",
				"website":       "https://website.com",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE employer SET slogan = $1, legal_address = $2, website = $3 WHERE id = $4"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["slogan"],
						fields["legal_address"],
						fields["website"],
						userID,
					).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:   "Ошибка NOT NULL для обязательного поля",
			userID: 3,
			fields: map[string]interface{}{
				"company_name": nil,
			},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("обязательное поле отсутствует"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE employer SET company_name = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["company_name"],
						userID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLNotNullViolation})
			},
		},

		{
			name:   "Ошибка DATA TYPE",
			userID: 3,
			fields: map[string]interface{}{
				"logo_id": "строковый тип",
			},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE employer SET logo_id = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["logo_id"],
						userID,
					).
					WillReturnError(&pq.Error{Code: entity.PSQLDatatypeViolation})
			},
		},
		{
			name:   "Нарушение CHECK ограничения почты",
			userID: 1,
			fields: map[string]interface{}{
				"email": "invalid.mail.ru",
			},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильные данные"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE employer SET email = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["email"],
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
				query := "UPDATE employer SET vk = $1 WHERE id = $2"

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
				"logo_id": 12,
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("неизвестная ошибка при обновлении работодателя err=pq: unknown error"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE employer SET logo_id = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["logo_id"],
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
				fmt.Errorf("не удалось обновить работодателя с id=8"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE employer SET telegram = $1 WHERE id = $2"

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
				"slogan": "У нас самые крутые образовательные курсы",
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить обновленные строки при обновлении работодателя с id=1"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE employer SET slogan = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["slogan"],
						userID,
					).
					WillReturnResult(&mockResultWithError{})
			},
		},
		{
			name:   "Ничего не обновилось — rowsAffected == 0",
			userID: 1,
			fields: map[string]interface{}{
				"slogan": "Не обновилось",
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось найти при обновлении работодателя с id=1"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, fields map[string]interface{}) {
				query := "UPDATE employer SET slogan = $1 WHERE id = $2"

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(
						fields["slogan"],
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

			defer func() {
				err := db.Close()
				require.NoError(t, err)
			}()

			repo := &EmployerRepository{DB: db}
			tc.setupMock(mock, tc.userID, tc.fields)

			ctx := context.Background()
			err = repo.UpdateEmployer(ctx, tc.userID, tc.fields)

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
