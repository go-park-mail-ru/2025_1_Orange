package postgres

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type VacancyRepository struct {
	DB *sql.DB
}

func NewVacancyRepository(db *sql.DB) (repository.VacancyRepository, error) {
	return &VacancyRepository{DB: db}, nil
}

func (r *VacancyRepository) Create(ctx context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error) {
	requestID := utils.GetRequestID(ctx)

	if vacancy.Title == "" || vacancy.SpecializationID == 0 {
		return nil, entity.NewError(entity.ErrBadRequest, fmt.Errorf("обязательное поле отсутствует"))
	}

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на создание резюме Create")

	query := `
	INSERT INTO vacancy (
        	employer_id,
            title,
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
			city,
			created_at,
			updated_at
	)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
        RETURNING id, employer_id, title, is_active, specialization_id, work_format,
            employment, schedule, working_hours, salary_from, salary_to,
            taxes_included, experience, description, tasks,
            requirements, optional_requirements, city, created_at, updated_at
    `
	var createdVacancy entity.Vacancy
	err := r.DB.QueryRowContext(ctx, query,
		vacancy.EmployerID,
		vacancy.Title,
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
		vacancy.City,
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
		&createdVacancy.City,
		&createdVacancy.CreatedAt,
		&createdVacancy.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLNotNullViolation:
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле отсутствует"),
				)
			case entity.PSQLUniqueViolation:
				return nil, entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("вакансия с такими параметрами уже существует"),
				)
			case entity.PSQLDatatypeViolation:
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неправильный формат данных"),
				)
			case entity.PSQLCheckViolation:
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("неправильные данные"),
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

func (r *VacancyRepository) AddSkills(ctx context.Context, vacancyID int, skillIDs []int) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на добавление навыков к вакансии AddSkills")

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при начале транзакции для добавления навыков")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при начале транзакции для добавления навыков: %w", err),
		)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				l.Log.WithFields(logrus.Fields{
					"requestID": requestID,
					"error":     rollbackErr,
				}).Error("ошибка при откате транзакции добавления навыков")
			}
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO vacancy_skill (vacancy_id, skill_id)
		VALUES ($1, $2)
	`)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при подготовке запроса для добавления навыков")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при подготовке запроса для добавления навыков: %w", err),
		)
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть statement: %v", err)
		}
	}(stmt)

	for _, skillID := range skillIDs {
		_, err = stmt.ExecContext(ctx, vacancyID, skillID)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) {
				switch pqErr.Code {
				case entity.PSQLUniqueViolation:
					continue // Пропускаем дубликаты
				case entity.PSQLNotNullViolation:
					return entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("обязательное поле отсутствует"),
					)
				case entity.PSQLDatatypeViolation:
					return entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("неправильный формат данных"),
					)
				case entity.PSQLCheckViolation:
					return entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("неправильные данные"),
					)
				}
			}

			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"vacancyID": vacancyID,
				"skillID":   skillID,
				"error":     err,
			}).Error("ошибка при добавлении навыка к вакансии")

			return entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении навыка к вакансии: %w", err),
			)
		}
	}

	if err = tx.Commit(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при коммите транзакции добавления навыков")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при коммите транзакции добавления навыков: %w", err),
		)
	}

	return nil
}

func (r *VacancyRepository) AddCity(ctx context.Context, vacancyID int, cityIDs []int) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на добавление города к вакансии AddSkills")

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при начале транзакции для добавления городов")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при начале транзакции для добавления городов: %w", err),
		)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				l.Log.WithFields(logrus.Fields{
					"requestID": requestID,
					"error":     rollbackErr,
				}).Error("ошибка при откате транзакции добавления городов")
			}
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO vacancy_city (vacancy_id, city_id)
		VALUES ($1, $2)
	`)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при подготовке запроса для добавления городов")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при подготовке запроса для добавления городов: %w", err),
		)
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть statement: %v", err)
		}
	}(stmt)

	for _, cityID := range cityIDs {
		_, err = stmt.ExecContext(ctx, vacancyID, cityID)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) {
				switch pqErr.Code {
				case entity.PSQLUniqueViolation:
					continue
				case entity.PSQLNotNullViolation:
					return entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("обязательное поле отсутствует"),
					)
				case entity.PSQLDatatypeViolation:
					return entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("неправильный формат данных"),
					)
				case entity.PSQLCheckViolation:
					return entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("неправильные данные"),
					)
				}
			}

			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"vacancyID": vacancyID,
				"cityID":    cityID,
				"error":     err,
			}).Error("ошибка при добавлении города к вакансии")

			return entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении города к вакансии: %w", err),
			)
		}
	}

	if err = tx.Commit(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при коммите транзакции добавления городов")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при коммите транзакции добавления городов: %w", err),
		)
	}

	return nil
}

func (r *VacancyRepository) CreateSkillIfNotExists(ctx context.Context, skillName string) (int, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"skillName": skillName,
	}).Info("sql-запрос в БД на создание навыка, если он не существует CreateSkillIfNotExists")

	// Сначала проверяем, существует ли навык
	var id int
	query := `
        SELECT id
        FROM skill
        WHERE name = $1
    `
	err := r.DB.QueryRowContext(ctx, query, skillName).Scan(&id)
	if err == nil {
		// Навык уже существует, возвращаем его ID
		return id, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		// Произошла ошибка, отличная от "запись не найдена"
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"skillName": skillName,
			"error":     err,
		}).Error("ошибка при проверке существования навыка")

		return 0, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при проверке существования навыка: %w", err),
		)
	}
	// Навык не существует, создаем его
	query = `
        INSERT INTO skill (name)
        VALUES ($1)
        RETURNING id
    `
	err = r.DB.QueryRowContext(ctx, query, skillName).Scan(&id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation:
				// Возможно, навык был создан другим запросом между нашими проверками
				// Попробуем получить его ID еще раз
				query = `
                    SELECT id
                    FROM skill
                    WHERE name = $1
                `
				err = r.DB.QueryRowContext(ctx, query, skillName).Scan(&id)
				if err != nil {
					l.Log.WithFields(logrus.Fields{
						"requestID": requestID,
						"skillName": skillName,
						"error":     err,
					}).Error("ошибка при получении ID навыка после конфликта")

					return 0, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении ID навыка после конфликта: %w", err),
					)
				}
				return id, nil
			default:
				l.Log.WithFields(logrus.Fields{
					"requestID": requestID,
					"skillName": skillName,
					"error":     err,
				}).Error("ошибка при создании навыка")

				return 0, entity.NewError(
					entity.ErrInternal,
					fmt.Errorf("ошибка при создании навыка: %w", err),
				)
			}
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"skillName": skillName,
			"error":     err,
		}).Error("ошибка при создании навыка")

		return 0, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при создании навыка: %w", err),
		)
	}

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"skillName": skillName,
		"skillID":   id,
	}).Info("навык успешно создан")

	return id, nil
}

func (r *VacancyRepository) GetSpecializationByResumeID(ctx context.Context, vacancyID int) ([]entity.Specialization, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на получение специализаций GetSpecializationsByResumeID")

	query := `
		SELECT s.id, s.name
		FROM specialization s
		JOIN vacancy_specialization rs ON s.id = rs.specialization_id
		WHERE rs.vacancy_id = $1
	`

	rows, err := r.DB.QueryContext(ctx, query, vacancyID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("ошибка при получении специализаций вакансии")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении специализаций вакансии: %w", err),
		)
	}
	defer rows.Close()

	var specializations []entity.Specialization
	for rows.Next() {
		var specialization entity.Specialization
		if err := rows.Scan(&specialization.ID, &specialization.Name); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"vacancyID": vacancyID,
				"error":     err,
			}).Error("ошибка при сканировании специализации")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании специализации: %w", err),
			)
		}
		specializations = append(specializations, specialization)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("ошибка при итерации по специализациям")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по специализациям: %w", err),
		)
	}

	return specializations, nil
}

func (r *VacancyRepository) GetByID(ctx context.Context, id int) (*entity.Vacancy, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
        SELECT 
            id,
            title,
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
			city,
			created_at,
			updated_at
        FROM vacancy
        WHERE id = $1
    `

	var vacancy entity.Vacancy
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&vacancy.ID,
		&vacancy.Title,
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
		&vacancy.City,
		&vacancy.CreatedAt,
		&vacancy.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"vacancyID": id,
			}).Debug("вакансия не найдена")

			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("вакансия с id=%d не найдена", id),
			)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": id,
			"error":     err,
		}).Error("ошибка при получении вакансии")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении вакансии: %w", err),
		)
	}

	return &vacancy, nil
}

func (r *VacancyRepository) Update(ctx context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на обновление вакансии Update")

	query := `
        UPDATE vacancy
        SET 
            title = $1,
            is_active = $2,
            specialization_id = $3,
            work_format = $4,
            employment = $5,
            schedule = $6,
            working_hours = $7,
            salary_from = $8,
            salary_to = $9,
            taxes_included = $10,
            experience = $11,
            description = $12,
            tasks = $13,
            requirements = $14,
            optional_requirements = $15,
			city = $16,
			created_at = NOW(),
            updated_at = NOW()
        WHERE id = $17 AND employer_id = $18
		RETURNING id, employer_id, title, is_active, specialization_id, work_format,
		 employment, schedule, working_hours, salary_from, salary_to, taxes_included,
		 experience, description, tasks, requirements, optional_requirements, city, created_at, updated_at
    `
	var updatedVacancy entity.Vacancy
	err := r.DB.QueryRowContext(ctx, query,
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
		vacancy.City,
		vacancy.ID,
		vacancy.EmployerID,
	).Scan(
		&updatedVacancy.ID,
		&updatedVacancy.EmployerID,
		&updatedVacancy.Title,
		&updatedVacancy.IsActive,
		&updatedVacancy.SpecializationID,
		&updatedVacancy.WorkFormat,
		&updatedVacancy.Employment,
		&updatedVacancy.Schedule,
		&updatedVacancy.WorkingHours,
		&updatedVacancy.SalaryFrom,
		&updatedVacancy.SalaryTo,
		&updatedVacancy.TaxesIncluded,
		&updatedVacancy.Experience,
		&updatedVacancy.Description,
		&updatedVacancy.Tasks,
		&updatedVacancy.Requirements,
		&updatedVacancy.OptionalRequirements,
		&updatedVacancy.City,
		&updatedVacancy.CreatedAt,
		&updatedVacancy.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // Уникальное ограничение
				return nil, entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("конфликт уникальных данных вакансии"),
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
			"id":        vacancy.ID,
			"error":     err,
		}).Error("не удалось обновить вакансию")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось обновить вакансию с id=%d", vacancy.ID),
		)
	}
	fmt.Println(updatedVacancy.SpecializationID)
	return &updatedVacancy, nil
}

func (r *VacancyRepository) GetAll(ctx context.Context, limit int, offset int) ([]*entity.Vacancy, error) {
	requestID := utils.GetRequestID(ctx)

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
			city,
			created_at,
			updated_at
        FROM vacancy
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
		`
	rows, err := r.DB.QueryContext(ctx, query, limit, offset)
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

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

	vacancies := make([]*entity.Vacancy, 0)
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
			&vacancy.City,
			&vacancy.CreatedAt,
			&vacancy.UpdatedAt,
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

	// // if len(vacancies) == 0 {
	// // 	return nil, entity.NewError(
	// // 		entity.ErrNotFound,
	// // 		fmt.Errorf("вакансии не найдены"),
	// // 	)
	// }
	return vacancies, nil
}

func (r *VacancyRepository) Delete(ctx context.Context, vacancyID int) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на удаление вакансии Delete")

	query := `
        DELETE FROM vacancy
        WHERE id = $1
    `

	result, err := r.DB.ExecContext(ctx, query, vacancyID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("не удалось удалить вакансию")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить вакансию с id=%d", vacancyID),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("ошибка при получении количества затронутых строк")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении количества затронутых строк: %w", err),
		)
	}

	if rowsAffected == 0 {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
		}).Warn("попытка удаления несуществующей вакансии")

		return entity.NewError(
			entity.ErrNotFound,
			fmt.Errorf("вакансия с id=%d не найдена", vacancyID),
		)
	}
	return nil
}

func (r *VacancyRepository) GetSkillsByVacancyID(ctx context.Context, vacancyID int) ([]entity.Skill, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на получение навыков вакансии GetSkillsByVacancyID")

	query := `
		SELECT s.id, s.name
		FROM skill s
		JOIN vacancy_skill vs ON s.id = vs.skill_id
		WHERE vs.vacancy_id = $1
	`

	rows, err := r.DB.QueryContext(ctx, query, vacancyID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("ошибка при получении навыков резюме")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении навыков резюме: %w", err),
		)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

	var skills []entity.Skill
	for rows.Next() {
		var skill entity.Skill
		if err := rows.Scan(&skill.ID, &skill.Name); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"vacancyID": vacancyID,
				"error":     err,
			}).Error("ошибка при сканировании навыка")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании навыка: %w", err),
			)
		}
		skills = append(skills, skill)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("ошибка при итерации по навыкам")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по навыкам: %w", err),
		)
	}

	return skills, nil
}

func (r *VacancyRepository) GetCityByVacancyID(ctx context.Context, vacancyID int) ([]entity.City, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на получение городов вакансии GetSkillsByVacancyID")

	query := `
		SELECT c.id, c.name
		FROM city c
		JOIN vacancy_city vc ON c.id = vc.city_id
		WHERE vc.vacancy_id = $1
	`

	rows, err := r.DB.QueryContext(ctx, query, vacancyID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("ошибка при получении городов резюме")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении городов резюме: %w", err),
		)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

	var cities []entity.City
	for rows.Next() {
		var skill entity.City
		if err := rows.Scan(&skill.ID, &skill.Name); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"vacancyID": vacancyID,
				"error":     err,
			}).Error("ошибка при сканировании навыка")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании навыка: %w", err),
			)
		}
		cities = append(cities, skill)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"resumeID":  vacancyID,
			"error":     err,
		}).Error("ошибка при итерации по навыкам")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по навыкам: %w", err),
		)
	}

	return cities, nil
}

func (r *VacancyRepository) DeleteSkills(ctx context.Context, vacancyID int) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на удаление навыков вакансии DeleteSkills")

	query := `
		DELETE FROM vacancy_skill
		WHERE vacancy_id = $1
	`

	_, err := r.DB.ExecContext(ctx, query, vacancyID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("ошибка при удалении навыков вакансии")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при удалении навыков вакансии: %w", err),
		)
	}

	return nil
}

func (r *VacancyRepository) DeleteCity(ctx context.Context, vacancyID int) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на удаление городов вакансии DeleteSkills")

	query := `
		DELETE FROM vacancy_city
		WHERE vacancy_id = $1
	`

	_, err := r.DB.ExecContext(ctx, query, vacancyID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("ошибка при удалении городов вакансии")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при удалении городов вакансии: %w", err),
		)
	}

	return nil
}

func (r *VacancyRepository) FindSkillIDsByNames(ctx context.Context, skillNames []string) ([]int, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на поиск ID навыков по названиям FindSkillIDsByNames")

	if len(skillNames) == 0 {
		return []int{}, nil
	}

	var skillIDs []int

	// Для каждого навыка проверяем его существование и создаем при необходимости
	for _, name := range skillNames {
		id, err := r.CreateSkillIfNotExists(ctx, name)
		if err != nil {
			return nil, err
		}
		skillIDs = append(skillIDs, id)
	}

	return skillIDs, nil
}

func (r *VacancyRepository) FindCityIDsByNames(ctx context.Context, cityNames []string) ([]int, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на поиск ID городов по названиям FindCityIDsByNames")

	if len(cityNames) == 0 {
		return []int{}, nil
	}

	params := make([]interface{}, len(cityNames))
	placeholders := make([]string, len(cityNames))
	for i, name := range cityNames {
		params[i] = name
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
        SELECT id
        FROM city
        WHERE name IN (%s)
    `, strings.Join(placeholders, ", "))

	rows, err := r.DB.QueryContext(ctx, query, params...)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при поиске ID городов по названиям")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при поиске ID городов по названиям: %w", err),
		)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

	var cityIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("ошибка при сканировании ID города")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании ID города: %w", err),
			)
		}
		cityIDs = append(cityIDs, id)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при итерации по ID городов")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по ID городов: %w", err),
		)
	}

	return cityIDs, nil
}

func (r *VacancyRepository) ResponseExists(ctx context.Context, vacancyID, applicantID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM vacancy_response WHERE vacancy_id = $1 AND applicant_id = $2)`
	var exists bool
	err := r.DB.QueryRowContext(ctx, query, vacancyID, applicantID).Scan(&exists)
	return exists, err
}

func (r *VacancyRepository) CreateResponse(ctx context.Context, vacancyID, applicantID int) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":   requestID,
		"vacancyID":   vacancyID,
		"applicantID": applicantID,
	}).Info("Creating vacancy response")

	// Получаем последнее резюме соискателя
	var resumeID sql.NullInt32
	err := r.DB.QueryRowContext(ctx,
		`SELECT id FROM resume WHERE applicant_id = $1 ORDER BY created_at DESC LIMIT 1`,
		applicantID,
	).Scan(&resumeID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.NewError(entity.ErrNotFound,
				fmt.Errorf("no active resumes found for applicant"))
		}
		return fmt.Errorf("failed to get applicant resume: %w", err)
	}

	query := `
        INSERT INTO vacancy_response (
            vacancy_id, 
            applicant_id,
            resume_id, 
            applied_at
        ) VALUES ($1, $2, $3, NOW())
    `

	_, err = r.DB.ExecContext(ctx, query, vacancyID, applicantID, resumeID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23503": // foreign key violation
				return entity.NewError(entity.ErrBadRequest,
					fmt.Errorf("vacancy or applicant does not exist"))
			case "23505": // unique violation
				return entity.NewError(entity.ErrAlreadyExists,
					fmt.Errorf("response already exists"))
			}
		}
		return fmt.Errorf("failed to create vacancy response: %w", err)
	}

	return nil
}

func (r *VacancyRepository) FindSpecializationIDByName(ctx context.Context, specializationName string) (int, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на поиск ID специализации по названию FindSpecializationIDByName")

	return r.CreateSpecializationIfNotExists(ctx, specializationName)
}

func (r *VacancyRepository) CreateSpecializationIfNotExists(ctx context.Context, specializationName string) (int, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":          requestID,
		"specializationName": specializationName,
	}).Info("sql-запрос в БД на создание специализации, если она не существует CreateSpecializationIfNotExists")

	// Сначала проверяем, существует ли специализация
	var id int
	query := `
        SELECT id
        FROM specialization
        WHERE name = $1
    `
	err := r.DB.QueryRowContext(ctx, query, specializationName).Scan(&id)
	if err == nil {
		// Специализация уже существует, возвращаем её ID
		return id, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		// Произошла ошибка, отличная от "запись не найдена"
		l.Log.WithFields(logrus.Fields{
			"requestID":          requestID,
			"specializationName": specializationName,
			"error":              err,
		}).Error("ошибка при проверке существования специализации")

		return 0, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при проверке существования специализации: %w", err),
		)
	}

	// Специализация не существует, создаем её
	query = `
        INSERT INTO specialization (name)
        VALUES ($1)
        RETURNING id
    `
	err = r.DB.QueryRowContext(ctx, query, specializationName).Scan(&id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation:
				// Возможно, специализация была создана другим запросом между нашими проверками
				// Попробуем получить её ID еще раз
				query = `
                    SELECT id
                    FROM specialization
                    WHERE name = $1
                `
				err = r.DB.QueryRowContext(ctx, query, specializationName).Scan(&id)
				if err != nil {
					l.Log.WithFields(logrus.Fields{
						"requestID":          requestID,
						"specializationName": specializationName,
						"error":              err,
					}).Error("ошибка при получении ID специализации после конфликта")

					return 0, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении ID специализации после конфликта: %w", err),
					)
				}
				return id, nil
			default:
				l.Log.WithFields(logrus.Fields{
					"requestID":          requestID,
					"specializationName": specializationName,
					"error":              err,
				}).Error("ошибка при создании специализации")

				return 0, entity.NewError(
					entity.ErrInternal,
					fmt.Errorf("ошибка при создании специализации: %w", err),
				)
			}
		}

		l.Log.WithFields(logrus.Fields{
			"requestID":          requestID,
			"specializationName": specializationName,
			"error":              err,
		}).Error("ошибка при создании специализации")

		return 0, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при создании специализации: %w", err),
		)
	}

	l.Log.WithFields(logrus.Fields{
		"requestID":          requestID,
		"specializationName": specializationName,
		"specializationID":   id,
	}).Info("специализация успешно создана")

	return id, nil
}

func (r *VacancyRepository) GetActiveVacanciesByEmployerID(ctx context.Context, employerID int, limit int, offset int) ([]*entity.Vacancy, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
        SELECT id, title, employer_id, specialization_id, work_format, employment, 
               schedule, working_hours, salary_from, salary_to, taxes_included, experience, 
               description, tasks, requirements, optional_requirements, city, created_at, updated_at
        FROM vacancy
        WHERE employer_id = $1 AND is_active = TRUE
        ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3;
    `

	rows, err := r.DB.QueryContext(ctx, query, employerID, limit, offset)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID":  requestID,
			"employerID": employerID,
			"error":      err,
		}).Error("Ошибка при получении активных вакансий работодателя")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении активных вакансий работодателя: %w", err),
		)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

	var vacancies []*entity.Vacancy
	for rows.Next() {
		var vacancy entity.Vacancy
		err := rows.Scan(
			&vacancy.ID, &vacancy.Title, &vacancy.EmployerID, &vacancy.SpecializationID,
			&vacancy.WorkFormat, &vacancy.Employment, &vacancy.Schedule, &vacancy.WorkingHours,
			&vacancy.SalaryFrom, &vacancy.SalaryTo, &vacancy.TaxesIncluded, &vacancy.Experience,
			&vacancy.Description, &vacancy.Tasks, &vacancy.Requirements, &vacancy.OptionalRequirements,
			&vacancy.City, &vacancy.CreatedAt, &vacancy.UpdatedAt,
		)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("Ошибка при сканировании вакансии")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка обработки данных вакансии: %w", err),
			)
		}
		vacancies = append(vacancies, &vacancy)
	}

	return vacancies, nil
}

func (r *VacancyRepository) GetVacanciesByApplicantID(ctx context.Context, applicantID int, limit int, offset int) ([]*entity.Vacancy, error) {
	requestID := utils.GetRequestID(ctx)

	query := `
        SELECT v.id, v.title, v.employer_id, v.specialization_id, v.work_format, 
               v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to, 
               v.taxes_included, v.experience, v.description, v.tasks, v.requirements, 
               v.optional_requirements, v.city, v.created_at, v.updated_at
        FROM vacancy v
        JOIN vacancy_response vr ON v.id = vr.vacancy_id
        WHERE vr.applicant_id = $1
        ORDER BY vr.applied_at DESC
		LIMIT $2 OFFSET $3
    `
	rows, err := r.DB.QueryContext(ctx, query, applicantID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID":   requestID,
			"applicantID": applicantID,
			"error":       err,
		}).Error("Ошибка при получении вакансий, на которые откликнулся соискатель")

		return nil, entity.NewError(entity.ErrInternal, fmt.Errorf("ошибка при получении списка вакансий: %w", err))
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

	var vacancies []*entity.Vacancy
	for rows.Next() {
		var vacancy entity.Vacancy
		err := rows.Scan(
			&vacancy.ID, &vacancy.Title, &vacancy.EmployerID, &vacancy.SpecializationID,
			&vacancy.WorkFormat, &vacancy.Employment, &vacancy.Schedule, &vacancy.WorkingHours,
			&vacancy.SalaryFrom, &vacancy.SalaryTo, &vacancy.TaxesIncluded, &vacancy.Experience,
			&vacancy.Description, &vacancy.Tasks, &vacancy.Requirements, &vacancy.OptionalRequirements,
			&vacancy.City, &vacancy.CreatedAt, &vacancy.UpdatedAt,
		)
		if err != nil {
			return nil, entity.NewError(entity.ErrInternal, fmt.Errorf("ошибка обработки данных вакансии: %w", err))
		}
		vacancies = append(vacancies, &vacancy)
	}

	if len(vacancies) == 0 {
		return []*entity.Vacancy{}, nil
	}

	return vacancies, nil
}

// SearchVacancies ищет вакансии по заданному запросу во всех вакансиях
func (r *VacancyRepository) SearchVacancies(ctx context.Context, searchQuery string, limit int, offset int) ([]*entity.Vacancy, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"query":     searchQuery,
	}).Info("sql-запрос в БД на поиск вакансий SearchVacancies")

	query := `
        SELECT v.id, v.title, v.is_active, v.employer_id, v.specialization_id, v.work_format, 
               v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to, 
               v.taxes_included, v.experience, v.description, v.tasks, v.requirements, 
               v.optional_requirements, v.city, v.created_at, v.updated_at
        FROM vacancy v
        JOIN employer e ON v.employer_id = e.id
        JOIN specialization s ON v.specialization_id = s.id
        WHERE v.title ILIKE $1 
           OR s.name ILIKE $1 
           OR e.company_name ILIKE $1
        ORDER BY v.updated_at DESC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.DB.QueryContext(ctx, query, "%"+searchQuery+"%", limit, offset)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при поиске вакансий")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при поиске вакансий: %w", err),
		)
	}
	// defer rows.Close()
	defer func() {
		if err := rows.Close(); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}()

	vacancies := make([]*entity.Vacancy, 0)
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
			&vacancy.City,
			&vacancy.CreatedAt,
			&vacancy.UpdatedAt,
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

	return vacancies, nil
}

// SearchVacanciesByEmployerID ищет вакансии по заданному запросу для конкретного работодателя
func (r *VacancyRepository) SearchVacanciesByEmployerID(ctx context.Context, employerID int, searchQuery string, limit int, offset int) ([]*entity.Vacancy, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"employerID": employerID,
		"query":      searchQuery,
	}).Info("sql-запрос в БД на поиск вакансий работодателя SearchVacanciesByEmployerID")

	query := `
        SELECT v.id, v.title, v.is_active, v.employer_id, v.specialization_id, v.work_format, 
               v.employment, v.schedule, v.working_hours, v.salary_from, v.salary_to, 
               v.taxes_included, v.experience, v.description, v.tasks, v.requirements, 
               v.optional_requirements, v.city, v.created_at, v.updated_at
        FROM vacancy v
        JOIN specialization s ON v.specialization_id = s.id
        WHERE v.employer_id = $1 
          AND (v.title ILIKE $2 OR s.name ILIKE $2)
        ORDER BY v.updated_at DESC
        LIMIT $3 OFFSET $4
    `

	rows, err := r.DB.QueryContext(ctx, query, employerID, "%"+searchQuery+"%", limit, offset)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при поиске вакансий работодателя")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при поиске вакансий работодателя: %w", err),
		)
	}
	// defer rows.Close()
	defer func() {
		if err := rows.Close(); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}()

	vacancies := make([]*entity.Vacancy, 0)
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
			&vacancy.City,
			&vacancy.CreatedAt,
			&vacancy.UpdatedAt,
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

	return vacancies, nil
}
