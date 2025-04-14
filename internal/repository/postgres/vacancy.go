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
			created_at
			updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
        RETURNING id
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
		&createdVacancy.CreatedAt,
		&createdVacancy.UpdatedAt,
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
	defer stmt.Close()

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

func (r *VacancyRepository) AddApplicant(ctx context.Context, vacancyID, applicantID int) error {
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
	defer stmt.Close()

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

func (r *VacancyRepository) GetByID(ctx context.Context, id int) (*entity.Vacancy, error) {
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
        FROM vacancy
        WHERE id = $1
    `

	var vacancy entity.Vacancy
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&vacancy.ID,
		&vacancy.EmployerID,
		&vacancy.Title,
		&vacancy.IsActive,
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
	var updatedVacancy entity.Vacancy
	err := r.DB.QueryRowContext(ctx, query,
		vacancy.ID,
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
	return &updatedVacancy, nil
}

func (r *VacancyRepository) GetAll(ctx context.Context) ([]*entity.Vacancy, error) {
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
			&vacancy.EmployerID,
			&vacancy.Title,
			&vacancy.IsActive,
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

func (r *VacancyRepository) Delete(ctx context.Context, vacancyID int) error {
	requestID := utils.GetRequestID(ctx)

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
		}).Error("не удалось получить количество удаленных строк")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось проверить удаление вакансии с id=%d", vacancyID),
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
	defer rows.Close()

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
		JOIN vacancy_city vc ON c.id = vc.skill_id
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
	defer rows.Close()

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

func (r *VacancyRepository) GetVacancyResponsesByVacancyID(ctx context.Context, vacancyID int) ([]entity.VacancyResponses, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"vacancyID": vacancyID,
	}).Info("SQL запрос на получение откликов по вакансии")

	query := `
		SELECT 
			r.id,
			r.vacancy_id,
			r.applicant_id,
			r.applied_at,
		FROM vacancy_response
        WHERE vacancy_id = $1
        ORDER BY applied_at DESC
	`
	rows, err := r.DB.QueryContext(ctx, query, vacancyID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("ошибка при получении откликов на вакансию")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении откликов на вакансию: %w", err),
		)
	}
	defer rows.Close()

	var responses []entity.VacancyResponses
	for rows.Next() {
		var response entity.VacancyResponses
		if err := rows.Scan(
			&response.ID,
			&response.VacancyID,
			&response.ApplicantID,
			&response.AppliedAt,
		); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"vacancyID": vacancyID,
				"error":     err,
			}).Error("ошибка при сканировании отклика")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании отклика: %w", err),
			)
		}
		responses = append(responses, response)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("ошибка при итерации по откликам")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по откликам: %w", err),
		)
	}
	return responses, nil
}

func (r *VacancyRepository) GetVacancyLikesByVacancyID(ctx context.Context, vacancyID int) ([]entity.VacancyLike, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"vacancyID": vacancyID,
	}).Info("SQL запрос на получение лайкнутых вакансий")

	query := `
        SELECT 
            id,
            vacancy_id,
            applicant_id,
            liked_at
        FROM vacancy_like
        WHERE vacancy_id = $1
        ORDER BY liked_at DESC
    `

	rows, err := r.DB.QueryContext(ctx, query, vacancyID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"vacancyID": vacancyID,
			"error":     err,
		}).Error("ошибка при получении лайкнутых вакансий")

		return nil, fmt.Errorf("failed to get liked vacancies: %w", err)
	}
	defer rows.Close()

	var likes []entity.VacancyLike
	for rows.Next() {
		var like entity.VacancyLike
		if err := rows.Scan(
			&like.ID,
			&like.VacancyID,
			&like.ApplicantID,
			&like.LikedAt,
		); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"vacancyID": vacancyID,
				"error":     err,
			}).Error("ошибка при сканировании лайка")

			return nil, fmt.Errorf("ошибка при сканировании лайка: %w", err)
		}
		likes = append(likes, like)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID":   requestID,
			"applicantID": vacancyID,
			"error":       err,
		}).Error("ошибка при итерации по лайкам")

		return nil, fmt.Errorf("error iterating vacancy likes: %w", err)
	}

	return likes, nil
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

	// Создаем параметры для запроса
	params := make([]interface{}, len(skillNames))
	placeholders := make([]string, len(skillNames))
	for i, name := range skillNames {
		params[i] = name
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT id
		FROM skill
		WHERE name IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := r.DB.QueryContext(ctx, query, params...)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при поиске ID навыков по названиям")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при поиске ID навыков по названиям: %w", err),
		)
	}
	defer rows.Close()

	var skillIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("ошибка при сканировании ID навыка")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании ID навыка: %w", err),
			)
		}
		skillIDs = append(skillIDs, id)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при итерации по ID навыков")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по ID навыков: %w", err),
		)
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
	defer rows.Close()

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

func (r *VacancyRepository) CreateResponse(ctx context.Context, vacancyID, applicantID, resumeID int) error {
	query := `
        INSERT INTO vacancy_response (
            vacancy_id, 
            applicant_id,
            resume_id, 
            applied_at
        ) VALUES ($1, $2, $3, NOW())
    `

	var err error
	if resumeID != -1 {
		_, err = r.DB.ExecContext(ctx, query, vacancyID, applicantID, resumeID)
	} else {
		_, err = r.DB.ExecContext(ctx, query, vacancyID, applicantID, nil)
	}

	if err != nil {
		return fmt.Errorf("Ошибка в создании отклика на вакансию: %w", err)
	}
	return nil
}
