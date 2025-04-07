package postgres

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/middleware"
	l "ResuMatch/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type VacancyRepository struct {
	DB *sql.DB
}

func NewVacancyRepository(cfg config.PostgresConfig) (*VacancyRepository, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось установить соединение с PostgreSQL из VacancyRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить соединение PostgreSQL из VacancyRepository: %w", err),
		)
	}

	if err := db.Ping(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось выполнить ping PostgreSQL из VacancyRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось выполнить ping PostgreSQL из VacancyRepository: %w", err),
		)
	}
	return &VacancyRepository{DB: db}, nil
}

func (r *VacancyRepository) Create(ctx context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error) {
	requestID := middleware.GetRequestID(ctx)

	query := `
        INSERT INTO vacancy (
            employer_id,
            title,
            is_active,
            specialization_id,
            work_format,
            employment,
            schedule,
            working_hours,
            salary_from,
            salary_to,
            taxes_included,
            experience,
            description,
            tasks,
            requirements,
            optional_requirements
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
        RETURNING 
            id,
            employer_id,
            title,
            is_active,
            specialization_id,
            work_format,
            employment,
            schedule,
            working_hours,
            salary_from,
            salary_to,
            taxes_included,
            experience,
            description,
            tasks,
            requirements,
            optional_requirements
    `

	var createdVacancy entity.Vacancy
	err := r.DB.QueryRowContext(ctx, query,
		vacancy.EmployerID,
		vacancy.Title,
		vacancy.IsActive,
		vacancy.SpecializationID,
		vacancy.WorkFormat,
		vacancy.Employment,
		vacancy.Schedule,
		vacancy.WorkingHours,
		vacancy.SalaryFrom,
		vacancy.SalaryTo,
		vacancy.TaxesIncluded,
		vacancy.Experience,
		vacancy.Description,
		vacancy.Tasks,
		vacancy.Requirements,
		vacancy.OptionalRequirements,
	).Scan(
		&createdVacancy.ID,
		&createdVacancy.EmployerID,
		&createdVacancy.Title,
		&createdVacancy.IsActive,
		&createdVacancy.SpecializationID,
		&createdVacancy.WorkFormat,
		&createdVacancy.Employment,
		&createdVacancy.Schedule,
		&createdVacancy.WorkingHours,
		&createdVacancy.SalaryFrom,
		&createdVacancy.SalaryTo,
		&createdVacancy.TaxesIncluded,
		&createdVacancy.Experience,
		&createdVacancy.Description,
		&createdVacancy.Tasks,
		&createdVacancy.Requirements,
		&createdVacancy.OptionalRequirements,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // Уникальное ограничение
				return nil, entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("вакансия с такими параметрами уже существует"),
				)
			case "23503": // Ошибка внешнего ключа
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("работодатель или специализация с указанным ID не существует"),
				)
			case "23502": // NOT NULL ограничение
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле отсутствует"),
				)
			case "22P02": // Ошибка типа данных
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неправильный формат данных"),
				)
			case "23514": // Ошибка constraint
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неправильные данные (например, salary_from > salary_to)"),
				)
			}
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при создании вакансии")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при создании вакансии: %w", err),
		)
	}

	return &createdVacancy, nil
}

func (r *VacancyRepository) GetByID(ctx context.Context, id int) (*entity.Vacancy, error) {
	requestID := middleware.GetRequestID(ctx)

	query := `
        SELECT 
            id,
            title,
            is_active,
            employer_id,
            specialization_id,
            work_format,
            employment,
            schedule,
            working_hours,
            salary_from,
            salary_to,
            taxes_included,
            experience,
            description,
            tasks,
            requirements,
            optional_requirements,
        FROM vacancy
        WHERE id = $1
    `

	var vacancy entity.Vacancy
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&vacancy.ID,
		&vacancy.Title,
		&vacancy.IsActive,
		&vacancy.EmployerID,
		&vacancy.SpecializationID,
		&vacancy.WorkFormat,
		&vacancy.Employment,
		&vacancy.Schedule,
		&vacancy.WorkingHours,
		&vacancy.SalaryFrom,
		&vacancy.SalaryTo,
		&vacancy.TaxesIncluded,
		&vacancy.Experience,
		&vacancy.Description,
		&vacancy.Tasks,
		&vacancy.Requirements,
		&vacancy.OptionalRequirements,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("вакансия с id=%d не найдена", id),
			)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        id,
			"error":     err,
		}).Error("не удалось найти вакансию по id")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить вакансию по id=%d", id),
		)
	}

	return &vacancy, nil
}

func (r *VacancyRepository) Update(ctx context.Context, vacancy *entity.Vacancy) error {
	requestID := middleware.GetRequestID(ctx)

	query := `
        UPDATE vacancy
        SET 
            title = $1,
            is_active = $2,
            employer_id = $3,
            specialization_id = $4,
            work_format = $5,
            employment = $6,
            schedule = $7,
            working_hours = $8,
            salary_from = $9,
            salary_to = $10,
            taxes_included = $11,
            experience = $12,
            description = $13,
            tasks = $14,
            requirements = $15,
            optional_requirements = $16,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $17
    `

	result, err := r.DB.ExecContext(ctx, query,
		vacancy.Title,
		vacancy.IsActive,
		vacancy.EmployerID,
		vacancy.SpecializationID,
		vacancy.WorkFormat,
		vacancy.Employment,
		vacancy.Schedule,
		vacancy.WorkingHours,
		vacancy.SalaryFrom,
		vacancy.SalaryTo,
		vacancy.TaxesIncluded,
		vacancy.Experience,
		vacancy.Description,
		vacancy.Tasks,
		vacancy.Requirements,
		vacancy.OptionalRequirements,
		vacancy.ID,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // Уникальное ограничение
				return entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("конфликт уникальных данных вакансии"),
				)
			case "23503": // Ошибка внешнего ключа
				return entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("работодатель или специализация с указанным ID не существует"),
				)
			case "23502": // NOT NULL ограничение
				return entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле отсутствует"),
				)
			case "22P02": // Ошибка типа данных
				return entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неправильный формат данных"),
				)
			case "23514": // Ошибка constraint
				return entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неправильные данные (например, salary_from > salary_to)"),
				)
			}
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        vacancy.ID,
			"error":     err,
		}).Error("не удалось обновить вакансию")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось обновить вакансию с id=%d", vacancy.ID),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        vacancy.ID,
			"error":     err,
		}).Error("не удалось получить обновленные строки при обновлении вакансии")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить обновленные строки при обновлении вакансии с id=%d", vacancy.ID),
		)
	}

	if rowsAffected == 0 {
		return entity.NewError(
			entity.ErrNotFound,
			fmt.Errorf("вакансия с id=%d не найдена", vacancy.ID),
		)
	}

	return nil
}

func (r *VacancyRepository) GetAll(ctx context.Context) ([]*entity.Vacancy, error) {
	requestID := middleware.GetRequestID(ctx)

	query := `
        SELECT 
            id,
            title,
            is_active,
            employer_id,
            specialization_id,
            work_format,
            employment,
            schedule,
            working_hours,
            salary_from,
            salary_to,
            taxes_included,
            experience,
            description,
            tasks,
            requirements,
            optional_requirements,
        FROM vacancy
    `

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("не удалось получить список вакансий")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить список вакансий: %w", err),
		)
	}
	defer rows.Close()

	var vacancies []*entity.Vacancy
	for rows.Next() {
		var vacancy entity.Vacancy
		err := rows.Scan(
			&vacancy.ID,
			&vacancy.Title,
			&vacancy.IsActive,
			&vacancy.EmployerID,
			&vacancy.SpecializationID,
			&vacancy.WorkFormat,
			&vacancy.Employment,
			&vacancy.Schedule,
			&vacancy.WorkingHours,
			&vacancy.SalaryFrom,
			&vacancy.SalaryTo,
			&vacancy.TaxesIncluded,
			&vacancy.Experience,
			&vacancy.Description,
			&vacancy.Tasks,
			&vacancy.Requirements,
			&vacancy.OptionalRequirements,
		)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("ошибка сканирования вакансии")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка обработки данных вакансии: %w", err),
			)
		}
		vacancies = append(vacancies, &vacancy)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при обработке результатов запроса")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при обработке результатов запроса вакансий: %w", err),
		)
	}

	if len(vacancies) == 0 {
		return nil, entity.NewError(
			entity.ErrNotFound,
			fmt.Errorf("вакансии не найдены"),
		)
	}

	return vacancies, nil
}

func (r *VacancyRepository) Delete(ctx context.Context, employerID, vacancyID int) error {
	requestID := middleware.GetRequestID(ctx)

	query := `
        DELETE FROM vacancy
        WHERE id = $1
        AND employer_id = $2
    `

	result, err := r.DB.ExecContext(ctx, query, vacancyID, employerID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID":  requestID,
			"employerID": employerID,
			"vacancyID":  vacancyID,
			"error":      err,
		}).Error("не удалось удалить вакансию")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить вакансию с id=%d для работодателя id=%d", vacancyID, employerID),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID":  requestID,
			"employerID": employerID,
			"vacancyID":  vacancyID,
			"error":      err,
		}).Error("не удалось получить количество удаленных строк")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось проверить удаление вакансии с id=%d для работодателя id=%d", vacancyID, employerID),
		)
	}

	if rowsAffected == 0 {
		l.Log.WithFields(logrus.Fields{
			"requestID":  requestID,
			"employerID": employerID,
			"vacancyID":  vacancyID,
		}).Warn("попытка удаления несуществующей вакансии")

		return entity.NewError(
			entity.ErrNotFound,
			fmt.Errorf("вакансия с id=%d для работодателя id=%d не найдена", vacancyID, employerID),
		)
	}

	return nil
}
