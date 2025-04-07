package postgres

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/middleware"
	"ResuMatch/internal/repository"
	l "ResuMatch/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type ApplicantRepository struct {
	DB *sql.DB
}

func NewApplicantRepository(cfg config.PostgresConfig) (repository.ApplicantRepository, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось установить соединение с PostgreSQL из ApplicantRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить соединение PostgreSQL из ApplicantRepository: %w", err),
		)
	}

	if err := db.Ping(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось выполнить ping PostgreSQL из ApplicantRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось выполнить ping PostgreSQL из ApplicantRepository: %w", err),
		)
	}
	return &ApplicantRepository{DB: db}, nil
}

func (r *ApplicantRepository) Create(ctx context.Context, applicant *entity.Applicant) (*entity.Applicant, error) {
	requestID := middleware.GetRequestID(ctx)

	query := `
        INSERT INTO applicant (email, password_hashed, password_salt, first_name, last_name)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, email, password_hashed, password_salt, first_name, last_name
    `

	var createdApplicant entity.Applicant
	err := r.DB.QueryRowContext(ctx, query,
		applicant.Email,
		applicant.PasswordHash,
		applicant.PasswordSalt,
		applicant.FirstName,
		applicant.LastName,
	).Scan(
		&createdApplicant.ID,
		&createdApplicant.Email,
		&createdApplicant.PasswordHash,
		&createdApplicant.PasswordSalt,
		&createdApplicant.FirstName,
		&createdApplicant.LastName,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation: // Уникальное ограничение
				return nil, entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("соискатель с таким email уже зарегистрирован"),
				)
			case entity.PSQLNotNullViolation: // NOT NULL ограничение
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле отсутствует"),
				)
			case entity.PSQLDatatypeViolation: // Ошибка типа данных
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неправильный формат данных"),
				)
			case entity.PSQLCheckViolation: // Ошибка constraint
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неправильные данные"),
				)
			}
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при создании соискателя")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при создании соискателя: %w", err),
		)
	}

	return &createdApplicant, nil
}

func (r *ApplicantRepository) GetByID(ctx context.Context, id int) (*entity.Applicant, error) {
	requestID := middleware.GetRequestID(ctx)

	query := `
		SELECT id, email, password_hashed, password_salt, first_name, last_name
		FROM applicant
		WHERE id = $1
	`

	var applicant entity.Applicant
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&applicant.ID,
		&applicant.Email,
		&applicant.PasswordHash,
		&applicant.PasswordSalt,
		&applicant.FirstName,
		&applicant.LastName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("соискатель с id=%d не найден", id),
			)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        id,
			"error":     err,
		}).Error("не удалось найти соискателя по id")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить соискателя по id=%d", id),
		)
	}

	return &applicant, nil
}

func (r *ApplicantRepository) GetByEmail(ctx context.Context, email string) (*entity.Applicant, error) {
	requestID := middleware.GetRequestID(ctx)

	query := `
		SELECT id, email, password_hashed, password_salt, first_name, last_name
		FROM applicant
		WHERE email = $1
	`

	var applicant entity.Applicant
	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&applicant.ID,
		&applicant.Email,
		&applicant.PasswordHash,
		&applicant.PasswordSalt,
		&applicant.FirstName,
		&applicant.LastName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("соискатель с email=%s не найден", email),
			)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"email":     email,
			"error":     err,
		}).Error("не удалось найти соискателя по email")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось найти соискателя с email=%s", email),
		)
	}

	return &applicant, nil
}

func (r *ApplicantRepository) Update(ctx context.Context, applicant *entity.Applicant) error {
	requestID := middleware.GetRequestID(ctx)

	query := `
		UPDATE applicant
		SET 
			email = $1,
			password_hashed = $2,
			password_salt = $3,
			first_name = $4,
			last_name = $5
		WHERE id = $6
	`

	result, err := r.DB.ExecContext(ctx, query,
		applicant.Email,
		applicant.PasswordHash,
		applicant.PasswordSalt,
		applicant.FirstName,
		applicant.LastName,
		applicant.ID,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation: // Уникальное ограничение
				return entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("соискатель с таким email уже зарегистрирован"),
				)
			case entity.PSQLNotNullViolation: // NOT NULL ограничение
				return entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле отсутствует"),
				)
			case entity.PSQLDatatypeViolation: // Ошибка типа данных
				return entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неправильный формат данных"),
				)
			case entity.PSQLCheckViolation: // Ошибка constraint
				return entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неправильные данные"),
				)
			}
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        applicant.ID,
			"error":     err,
		}).Error("не удалось обновить соискателя")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось обновить соискателя с id=%d", applicant.ID),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        applicant.ID,
			"error":     err,
		}).Error("не удалось получить обновленные строки при обновлении соискателя")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить обновленные строки при обновлении соискателя с id=%d", applicant.ID),
		)
	}

	if rowsAffected == 0 {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        applicant.ID,
			"error":     err,
		}).Error("не удалось найти при обновлении соискателя")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось найти при обновлении соискателя с id=%d", applicant.ID),
		)
	}

	return nil
}
