package postgres

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type ApplicantDB struct {
	DB *sql.DB
}

func NewApplicantRepository(cfg config.PostgresDBConfig) (repository.ApplicantRepository, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, entity.NewClientError("failed to connect to PostgreSQL", entity.ErrPostgres)
	}

	if err := db.Ping(); err != nil {
		return nil, entity.NewClientError("failed to ping PostgreSQL", entity.ErrPostgres)
	}
	return &ApplicantDB{DB: db}, nil
}

func (r *ApplicantDB) Create(ctx context.Context, applicant *entity.Applicant) (*entity.Applicant, error) {
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
			case "23505": // Уникальное ограничение
				return nil, entity.NewClientError("Соискатель с таким email уже зарегистрирован", entity.ErrAlreadyExists)
			case "23502": // NOT NULL ограничение
				return nil, entity.NewClientError("Обязательное поле отсутствует", entity.ErrBadRequest)
			case "22P02": // Ошибка типа данных
				return nil, entity.NewClientError("Некорректный формат данных", entity.ErrBadRequest)
			}
		}

		return nil, entity.NewClientError("Ошибка создания соискателя", entity.ErrPostgres)
	}

	return &createdApplicant, nil
}

func (r *ApplicantDB) GetByID(ctx context.Context, id int) (*entity.Applicant, error) {
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
			return nil, entity.NewClientError(fmt.Sprintf("applicant with id=%d not found", id), entity.ErrNotFound)
		}
		return nil, entity.NewClientError(fmt.Sprintf("failed to get Applicant with id=%d", id), entity.ErrPostgres)
	}

	return &applicant, nil
}

func (r *ApplicantDB) GetByEmail(ctx context.Context, email string) (*entity.Applicant, error) {
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
			return nil, entity.NewClientError(fmt.Sprintf("applicant with email=%s not found", email), entity.ErrNotFound)
		}
		return nil, entity.NewClientError(fmt.Sprintf("failed to get Applicant with email=%s", email), entity.ErrPostgres)
	}

	return &applicant, nil
}

func (r *ApplicantDB) Update(ctx context.Context, applicant *entity.Applicant) error {
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
			case "23505": // Уникальное ограничение
				return entity.NewClientError("Соискатель с таким email уже зарегистрирован", entity.ErrAlreadyExists)
			case "23502": // NOT NULL ограничение
				return entity.NewClientError("Обязательное поле отсутствует", entity.ErrBadRequest)
			case "22P02": // Ошибка типа данных
				return entity.NewClientError("Некорректный формат данных", entity.ErrBadRequest)
			}
		}
		return entity.NewClientError(fmt.Sprintf("failed to update Applicant with id=%d", applicant.ID), entity.ErrPostgres)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return entity.NewClientError(fmt.Sprintf("failed to get rows affected while updating applicant with id=%d", applicant.ID), entity.ErrPostgres)
	}

	if rowsAffected == 0 {
		return entity.NewClientError(fmt.Sprintf("failed to find applicant for update with id=%d", applicant.ID), entity.ErrPostgres)
	}

	return nil
}
