package postgres

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/metrics"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"strings"
)

type EmployerRepository struct {
	DB *sql.DB
}

type ScanEmployer struct {
	ID           int
	CompanyName  string
	LegalAddress string
	Email        string
	Slogan       sql.NullString
	Website      sql.NullString
	Description  sql.NullString
	Vk           sql.NullString
	Telegram     sql.NullString
	Facebook     sql.NullString
	LogoID       sql.NullInt64
	PasswordHash []byte
	PasswordSalt []byte
	CreatedAt    sql.NullTime
	UpdatedAt    sql.NullTime
}

func (e *ScanEmployer) GetEntity() *entity.Employer {
	employer := &entity.Employer{
		ID:           e.ID,
		CompanyName:  e.CompanyName,
		LegalAddress: e.LegalAddress,
		Email:        e.Email,
		Slogan:       e.Slogan.String,
		Website:      e.Website.String,
		Description:  e.Description.String,
		Vk:           e.Vk.String,
		Telegram:     e.Telegram.String,
		Facebook:     e.Facebook.String,
		LogoID:       int(e.LogoID.Int64),
		PasswordHash: e.PasswordHash,
		PasswordSalt: e.PasswordSalt,
		CreatedAt:    e.CreatedAt.Time,
		UpdatedAt:    e.UpdatedAt.Time,
	}
	return employer
}

func NewEmployerRepository(db *sql.DB) (repository.EmployerRepository, error) {
	return &EmployerRepository{DB: db}, nil
}

func (r *EmployerRepository) CreateEmployer(ctx context.Context, email, companyName, legalAddress string, passwordHash, passwordSalt []byte) (*entity.Employer, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("выполнение sql-запроса создания работодателя CreateEmployer")

	query := `
		INSERT INTO employer (email, password_hashed, password_salt, company_name, legal_address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, password_hashed, password_salt, company_name, legal_address
	`

	var createdEmployer entity.Employer
	err := r.DB.QueryRowContext(ctx, query,
		email,
		passwordHash,
		passwordSalt,
		companyName,
		legalAddress,
	).Scan(
		&createdEmployer.ID,
		&createdEmployer.Email,
		&createdEmployer.PasswordHash,
		&createdEmployer.PasswordSalt,
		&createdEmployer.CompanyName,
		&createdEmployer.LegalAddress,
	)

	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Employer Repository", "CreateEmployer").Inc()
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation: // Уникальное ограничение
				return nil, entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("такой работодатель уже зарегистрирован"),
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
			default:
				return nil, entity.NewError(
					entity.ErrInternal,
					fmt.Errorf("неизвестная ошибка при создании работодателя err=%w", err),
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

func (r *EmployerRepository) GetEmployerByID(ctx context.Context, id int) (*entity.Employer, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"id":        id,
	}).Info("выполнение sql-запроса получения работодателя по ID GetEmployerByID")

	query := `
		SELECT id, email, password_hashed, password_salt, company_name,
		       legal_address, vk, telegram, facebook, slogan, 
		       website, description, logo_id, created_at, updated_at
		FROM employer
		WHERE id = $1
	`

	scanEmployer := ScanEmployer{}
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&scanEmployer.ID,
		&scanEmployer.Email,
		&scanEmployer.PasswordHash,
		&scanEmployer.PasswordSalt,
		&scanEmployer.CompanyName,
		&scanEmployer.LegalAddress,
		&scanEmployer.Vk,
		&scanEmployer.Telegram,
		&scanEmployer.Facebook,
		&scanEmployer.Slogan,
		&scanEmployer.Website,
		&scanEmployer.Description,
		&scanEmployer.LogoID,
		&scanEmployer.CreatedAt,
		&scanEmployer.UpdatedAt,
	)

	employer := scanEmployer.GetEntity()
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Employer Repository", "GetEmployerByID").Inc()
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

	return employer, nil
}

func (r *EmployerRepository) GetEmployerByEmail(ctx context.Context, email string) (*entity.Employer, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"email":     email,
	}).Info("выполнение sql-запроса получения работодателя по почте GetEmployerByEmail")

	query := `
		SELECT id, email, password_hashed, password_salt, company_name,
		       legal_address, vk, telegram, facebook, slogan,
		       website, description, logo_id, created_at, updated_at
		FROM employer
		WHERE email = $1
	`

	scanEmployer := ScanEmployer{}
	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&scanEmployer.ID,
		&scanEmployer.Email,
		&scanEmployer.PasswordHash,
		&scanEmployer.PasswordSalt,
		&scanEmployer.CompanyName,
		&scanEmployer.LegalAddress,
		&scanEmployer.Vk,
		&scanEmployer.Telegram,
		&scanEmployer.Facebook,
		&scanEmployer.Slogan,
		&scanEmployer.Website,
		&scanEmployer.Description,
		&scanEmployer.LogoID,
		&scanEmployer.CreatedAt,
		&scanEmployer.UpdatedAt,
	)

	employer := scanEmployer.GetEntity()
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Employer Repository", "GetEmployerByEmail").Inc()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("работодатель с email=%s не найден", email),
			)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"email":     email,
			"error":     err,
		}).Error("не удалось найти работодателя по email")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось найти работодателя с email=%s", email),
		)
	}

	return employer, nil
}

func (r *EmployerRepository) UpdateEmployer(ctx context.Context, userID int, fields map[string]interface{}) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("выполнение sql-запроса обновления информации работодателя UpdateEmployer")

	query := "UPDATE employer SET "
	setParts := make([]string, 0, len(fields))
	args := make([]interface{}, 0, len(fields)+1)
	i := 1

	for field, value := range fields {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, i))
		args = append(args, value)
		i++
	}

	query += strings.Join(setParts, ", ")
	query += fmt.Sprintf(" WHERE id = $%d", i)
	args = append(args, userID)

	result, err := r.DB.ExecContext(ctx, query, args...)

	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Employer Repository", "UpdateEmployer").Inc()
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation: // Уникальное ограничение
				return entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("ошибка уникальности"),
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
			default:
				return entity.NewError(
					entity.ErrInternal,
					fmt.Errorf("неизвестная ошибка при обновлении работодателя err=%w", err),
				)
			}
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        userID,
			"error":     err,
		}).Error("не удалось обновить работодателя")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось обновить работодателя с id=%d", userID),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Employer Repository", "UpdateEmployer").Inc()
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        userID,
			"error":     err,
		}).Error("не удалось получить обновленные строки при обновлении работодателя")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить обновленные строки при обновлении работодателя с id=%d", userID),
		)
	}

	if rowsAffected == 0 {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        userID,
			"error":     err,
		}).Error("не удалось найти при обновлении работодателя")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось найти при обновлении работодателя с id=%d", userID),
		)
	}

	return nil
}
