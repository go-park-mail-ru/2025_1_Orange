package postgres

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type EmployerRepository struct {
	DB *sql.DB
}

func NewEmployerRepository(cfg config.PostgresConfig) (*EmployerRepository, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось установить соединение с PostgreSQL из EmployerRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить соединение PostgreSQL из EmployerRepository: %w", err),
		)
	}

	if err := db.Ping(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось выполнить ping PostgreSQL из EmployerRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось выполнить ping PostgreSQL из EmployerRepository: %w", err),
		)
	}

	return &EmployerRepository{DB: db}, nil
}

func (r *EmployerRepository) Create(ctx context.Context, employer *entity.Employer) (*entity.Employer, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
		INSERT INTO employer (email, password_hashed, password_salt, company_name, legal_address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, password_hashed, password_salt, company_name, legal_address
	`

	var createdEmployer entity.Employer
	err := r.DB.QueryRowContext(ctx, query,
		employer.Email,
		employer.PasswordHash,
		employer.PasswordSalt,
		employer.CompanyName,
		employer.LegalAddress,
	).Scan(
		&createdEmployer.ID,
		&createdEmployer.Email,
		&createdEmployer.PasswordHash,
		&createdEmployer.PasswordSalt,
		&createdEmployer.CompanyName,
		&createdEmployer.LegalAddress,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation: // Уникальное ограничение
				return nil, entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("работодатель с таким email уже зарегистрирован"),
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
		}).Error("ошибка при создании работодателя")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при создании работодателя: %w", err),
		)
	}

	return &createdEmployer, nil
}

func (r *EmployerRepository) GetByID(ctx context.Context, id int) (*entity.Employer, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
		SELECT id, email, password_hashed, password_salt, company_name, legal_address
		FROM employer
		WHERE id = $1
	`

	var employer entity.Employer
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&employer.ID,
		&employer.Email,
		&employer.PasswordHash,
		&employer.PasswordSalt,
		&employer.CompanyName,
		&employer.LegalAddress,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("работодатель с id=%d не найден", id),
			)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        id,
			"error":     err,
		}).Error("не удалось найти работодателя по id")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить работодателя по id=%d", id),
		)
	}

	return &employer, nil
}

func (r *EmployerRepository) GetByEmail(ctx context.Context, email string) (*entity.Employer, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
		SELECT id, email, password_hashed, password_salt, company_name, legal_address
		FROM employer
		WHERE email = $1
	`

	var employer entity.Employer
	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&employer.ID,
		&employer.Email,
		&employer.PasswordHash,
		&employer.PasswordSalt,
		&employer.CompanyName,
		&employer.LegalAddress,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("работодатель с email=%s не найден", email),
			)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"email":     employer.Email,
			"error":     err,
		}).Error("не удалось найти работодателя по email")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось найти работодателя с email=%s", email),
		)
	}

	return &employer, nil
}

func (r *EmployerRepository) Update(ctx context.Context, employer *entity.Employer) error {
	requestID := utils.GetRequestID(ctx)

	query := `
		UPDATE employer
		SET 
			email = $1,
			password_hashed = $2,
			password_salt = $3,
			company_name = $4,
			legal_address = $5
		WHERE id = $6
	`

	result, err := r.DB.ExecContext(ctx, query,
		employer.Email,
		employer.PasswordHash,
		employer.PasswordSalt,
		employer.CompanyName,
		employer.LegalAddress,
		employer.ID,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation: // Уникальное ограничение
				return entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("работодатель с таким email уже зарегистрирован"),
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
			"id":        employer.ID,
			"error":     err,
		}).Error("не удалось обновить работодателя")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось обновить работодателя с id=%d", employer.ID),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        employer.ID,
			"error":     err,
		}).Error("не удалось получить обновленные строки при обновлении работодателя")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить обновленные строки при обновлении работодателя с id=%d", employer.ID),
		)
	}

	if rowsAffected == 0 {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        employer.ID,
			"error":     err,
		}).Error("не удалось найти при обновлении работодателя")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось найти при обновлении работодателя с id=%d", employer.ID),
		)
	}

	return nil
}
