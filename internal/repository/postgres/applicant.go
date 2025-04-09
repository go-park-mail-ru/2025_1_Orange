package postgres

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"strings"
)

type ApplicantRepository struct {
	DB *sql.DB
}

type ScanApplicant struct {
	ID           int
	FirstName    string
	LastName     string
	MiddleName   sql.NullString
	Email        string
	CityID       sql.NullInt64
	BirthDate    sql.NullTime
	Sex          sql.NullString
	Status       sql.NullString
	Quote        sql.NullString
	AvatarID     sql.NullInt64
	PasswordHash []byte
	PasswordSalt []byte
	CreatedAt    sql.NullTime
	UpdatedAt    sql.NullTime
}

func (a *ScanApplicant) GetEntity() *entity.Applicant {
	applicant := &entity.Applicant{
		ID:           a.ID,
		FirstName:    a.FirstName,
		LastName:     a.LastName,
		MiddleName:   a.MiddleName.String,
		BirthDate:    a.BirthDate.Time,
		Email:        a.Email,
		CityID:       int(a.CityID.Int64),
		Sex:          a.Sex.String,
		Status:       toApplicantStatus(a.Status),
		Quote:        a.Quote.String,
		AvatarID:     int(a.AvatarID.Int64),
		PasswordHash: a.PasswordHash,
		PasswordSalt: a.PasswordSalt,
		CreatedAt:    a.CreatedAt.Time,
		UpdatedAt:    a.UpdatedAt.Time,
	}
	return applicant
}

func toApplicantStatus(status sql.NullString) entity.ApplicantStatus {
	if status.Valid {
		return entity.ApplicantStatus(status.String)
	}
	return ""
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

func (r *ApplicantRepository) CreateApplicant(
	ctx context.Context, email, firstName, lastName string, passwordHash, passwordSalt []byte) (*entity.Applicant, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
        INSERT INTO applicant (email, password_hashed, password_salt, first_name, last_name)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, email, password_hashed, password_salt, first_name, last_name
    `

	var createdApplicant entity.Applicant
	err := r.DB.QueryRowContext(ctx, query,
		email,
		passwordHash,
		passwordSalt,
		firstName,
		lastName,
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
			default:
				return nil, entity.NewError(
					entity.ErrInternal,
					fmt.Errorf("неизвестная ошибка при создании соискателя err=%w", err),
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

func (r *ApplicantRepository) GetApplicantByID(ctx context.Context, id int) (*entity.Applicant, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
		SELECT id, first_name, last_name, middle_name, city_id, 
		       birth_date, sex, email, status, quote, avatar_id,
		       password_hashed, password_salt, created_at, updated_at
		FROM applicant WHERE id = $1
	`

	scanApplicant := ScanApplicant{}
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&scanApplicant.ID,
		&scanApplicant.FirstName,
		&scanApplicant.LastName,
		&scanApplicant.MiddleName,
		&scanApplicant.CityID,
		&scanApplicant.BirthDate,
		&scanApplicant.Sex,
		&scanApplicant.Email,
		&scanApplicant.Status,
		&scanApplicant.Quote,
		&scanApplicant.AvatarID,
		&scanApplicant.PasswordHash,
		&scanApplicant.PasswordSalt,
		&scanApplicant.CreatedAt,
		&scanApplicant.UpdatedAt,
	)

	applicant := scanApplicant.GetEntity()

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

	return applicant, nil
}

func (r *ApplicantRepository) GetApplicantByEmail(ctx context.Context, email string) (*entity.Applicant, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
		SELECT id, first_name, last_name, middle_name, city_id, 
		       birth_date, sex, email, status, quote, avatar_id,
		       password_hashed, password_salt, created_at, updated_at
		FROM applicant WHERE email = $1
	`

	scanApplicant := ScanApplicant{}
	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&scanApplicant.ID,
		&scanApplicant.FirstName,
		&scanApplicant.LastName,
		&scanApplicant.MiddleName,
		&scanApplicant.CityID,
		&scanApplicant.BirthDate,
		&scanApplicant.Sex,
		&scanApplicant.Email,
		&scanApplicant.Status,
		&scanApplicant.Quote,
		&scanApplicant.AvatarID,
		&scanApplicant.PasswordHash,
		&scanApplicant.PasswordSalt,
		&scanApplicant.CreatedAt,
		&scanApplicant.UpdatedAt,
	)

	applicant := scanApplicant.GetEntity()

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

	return applicant, nil
}

func (r *ApplicantRepository) UpdateApplicant(ctx context.Context, userID int, fields map[string]interface{}) error {
	requestID := utils.GetRequestID(ctx)

	query := "UPDATE applicant SET "
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

	_, err := r.DB.ExecContext(ctx, query, args...)

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
			"id":        userID,
			"error":     err,
		}).Error("не удалось обновить соискателя")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось обновить соискателя с id=%d", userID),
		)
	}

	return nil
}
