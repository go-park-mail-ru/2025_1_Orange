package postgres

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type EmployerDB struct {
	DB *sql.DB
}

func NewEmployerDB(cfg config.PostgresConfig) (*EmployerDB, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, entity.NewClientError("failed to connect to PostgreSQL", entity.ErrPostgres)
	}

	if err := db.Ping(); err != nil {
		return nil, entity.NewClientError("failed to ping PostgreSQL", entity.ErrPostgres)
	}

	return &EmployerDB{DB: db}, nil
}

func (r *EmployerDB) Create(ctx context.Context, employer *entity.Employer) (*entity.Employer, error) {
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
			case "23505": // Уникальное ограничение
				return nil, entity.NewClientError("Работодатель с таким email или названием компании уже зарегистрирован", entity.ErrAlreadyExists)
			case "23502": // NOT NULL ограничение
				return nil, entity.NewClientError("Обязательное поле отсутствует", entity.ErrBadRequest)
			case "22P02": // Ошибка типа данных
				return nil, entity.NewClientError("Некорректный формат данных", entity.ErrBadRequest)
			}
		}

		return nil, entity.NewClientError("Ошибка создания работодателя", entity.ErrPostgres)
	}

	return &createdEmployer, nil
}

func (r *EmployerDB) GetByID(ctx context.Context, id int) (*entity.Employer, error) {
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
			return nil, entity.NewClientError(fmt.Sprintf("employer with id=%d not found", id), entity.ErrNotFound)
		}
		return nil, entity.NewClientError(fmt.Sprintf("failed to get Employer with id=%d", id), entity.ErrPostgres)
	}

	return &employer, nil
}

func (r *EmployerDB) GetByEmail(ctx context.Context, email string) (*entity.Employer, error) {
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
			return nil, entity.NewClientError(fmt.Sprintf("employer with email=%s not found", email), entity.ErrNotFound)
		}
		return nil, entity.NewClientError(fmt.Sprintf("failed to get Employer with email=%s", email), entity.ErrPostgres)
	}

	return &employer, nil
}

func (r *EmployerDB) Update(ctx context.Context, employer *entity.Employer) error {
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
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return entity.NewClientError("email или название компании уже используется", entity.ErrAlreadyExists)
		}
		return entity.NewClientError(fmt.Sprintf("failed to update Employer with id=%d", employer.ID), entity.ErrPostgres)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return entity.NewClientError(fmt.Sprintf("failed to get rows affected while updating employer with id=%d", employer.ID), entity.ErrPostgres)
	}

	if rowsAffected == 0 {
		return entity.NewClientError(fmt.Sprintf("failed to find employer for update with id=%d", employer.ID), entity.ErrPostgres)
	}

	return nil
}
