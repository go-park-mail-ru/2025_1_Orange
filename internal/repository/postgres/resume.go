package postgres

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/middleware"
	"ResuMatch/internal/repository"
	l "ResuMatch/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type ResumeRepository struct {
	DB *sql.DB
}

// Замечание 10 - Добавление коннектора
// Изменен конструктор для использования готового соединения с БД
func NewResumeRepository(db *sql.DB) (repository.ResumeRepository, error) {
	return &ResumeRepository{DB: db}, nil
}

func (r *ResumeRepository) Create(ctx context.Context, resume *entity.Resume) (*entity.Resume, error) {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на создание резюме Create")

	query := `
		INSERT INTO resume (
			applicant_id, about_me, specialization_id, education, 
			educational_institution, graduation_year, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, applicant_id, about_me, specialization_id, education, 
				  educational_institution, graduation_year, created_at, updated_at
	`

	var createdResume entity.Resume
	err := r.DB.QueryRowContext(
		ctx,
		query,
		resume.ApplicantID,
		resume.AboutMe,
		resume.SpecializationID,
		resume.Education,
		resume.EducationalInstitution,
		resume.GraduationYear,
	).Scan(
		&createdResume.ID,
		&createdResume.ApplicantID,
		&createdResume.AboutMe,
		&createdResume.SpecializationID,
		&createdResume.Education,
		&createdResume.EducationalInstitution,
		&createdResume.GraduationYear,
		&createdResume.CreatedAt,
		&createdResume.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation:
				return nil, entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("резюме с такими параметрами уже существует"),
				)
			case entity.PSQLNotNullViolation:
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле отсутствует"),
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
		}).Error("ошибка при создании резюме")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при создании резюме: %w", err),
		)
	}

	return &createdResume, nil
}

func (r *ResumeRepository) AddSkills(ctx context.Context, resumeID int, skillIDs []int) error {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на добавление навыков к резюме AddSkills")

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
		INSERT INTO resume_skill (resume_id, skill_id)
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
		_, err = stmt.ExecContext(ctx, resumeID, skillID)
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
				"resumeID":  resumeID,
				"skillID":   skillID,
				"error":     err,
			}).Error("ошибка при добавлении навыка к резюме")

			return entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении навыка к резюме: %w", err),
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

func (r *ResumeRepository) AddWorkExperience(ctx context.Context, workExperience *entity.WorkExperience) (*entity.WorkExperience, error) {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на добавление опыта работы AddWorkExperience")

	query := `
		INSERT INTO work_experience (
			resume_id, employer_name, position, duties, 
			achievements, start_date, end_date, until_now, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING id, resume_id, employer_name, position, duties, 
				  achievements, start_date, end_date, until_now, updated_at
	`

	var endDate sql.NullTime
	if !workExperience.UntilNow && !workExperience.EndDate.IsZero() {
		endDate = sql.NullTime{
			Time:  workExperience.EndDate,
			Valid: true,
		}
	}

	var createdWorkExperience entity.WorkExperience
	err := r.DB.QueryRowContext(
		ctx,
		query,
		workExperience.ResumeID,
		workExperience.EmployerName,
		workExperience.Position,
		workExperience.Duties,
		workExperience.Achievements,
		workExperience.StartDate,
		endDate,
		workExperience.UntilNow,
	).Scan(
		&createdWorkExperience.ID,
		&createdWorkExperience.ResumeID,
		&createdWorkExperience.EmployerName,
		&createdWorkExperience.Position,
		&createdWorkExperience.Duties,
		&createdWorkExperience.Achievements,
		&createdWorkExperience.StartDate,
		&endDate,
		&createdWorkExperience.UntilNow,
		&createdWorkExperience.UpdatedAt,
	)

	if endDate.Valid {
		createdWorkExperience.EndDate = endDate.Time
	}

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation:
				return nil, entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("опыт работы с такими параметрами уже существует"),
				)
			case entity.PSQLNotNullViolation:
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле отсутствует"),
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
		}).Error("ошибка при создании опыта работы")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при создании опыта работы: %w", err),
		)
	}

	return &createdWorkExperience, nil
}

func (r *ResumeRepository) GetByID(ctx context.Context, id int) (*entity.Resume, error) {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на получение резюме по ID GetByID")

	query := `
		SELECT id, applicant_id, about_me, specialization_id, education, 
			   educational_institution, graduation_year, created_at, updated_at
		FROM resume
		WHERE id = $1
	`

	var resume entity.Resume
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&resume.ID,
		&resume.ApplicantID,
		&resume.AboutMe,
		&resume.SpecializationID,
		&resume.Education,
		&resume.EducationalInstitution,
		&resume.GraduationYear,
		&resume.CreatedAt,
		&resume.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=%d не найдено", id),
			)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"id":        id,
			"error":     err,
		}).Error("не удалось найти резюме по id")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить резюме по id=%d", id),
		)
	}

	return &resume, nil
}

func (r *ResumeRepository) GetSkillsByResumeID(ctx context.Context, resumeID int) ([]entity.Skill, error) {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на получение навыков резюме GetSkillsByResumeID")

	query := `
		SELECT s.id, s.name
		FROM skill s
		JOIN resume_skill rs ON s.id = rs.skill_id
		WHERE rs.resume_id = $1
	`

	rows, err := r.DB.QueryContext(ctx, query, resumeID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"resumeID":  resumeID,
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
				"resumeID":  resumeID,
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
			"resumeID":  resumeID,
			"error":     err,
		}).Error("ошибка при итерации по навыкам")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по навыкам: %w", err),
		)
	}

	return skills, nil
}

func (r *ResumeRepository) GetWorkExperienceByResumeID(ctx context.Context, resumeID int) ([]entity.WorkExperience, error) {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на получение опыта работы GetWorkExperienceByResumeID")

	query := `
		SELECT id, resume_id, employer_name, position, duties, 
			   achievements, start_date, end_date, until_now, updated_at
		FROM work_experience
		WHERE resume_id = $1
		ORDER BY start_date DESC
	`

	rows, err := r.DB.QueryContext(ctx, query, resumeID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"resumeID":  resumeID,
			"error":     err,
		}).Error("ошибка при получении опыта работы")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении опыта работы: %w", err),
		)
	}
	defer rows.Close()

	var experiences []entity.WorkExperience
	for rows.Next() {
		var experience entity.WorkExperience
		var endDate sql.NullTime

		if err := rows.Scan(
			&experience.ID,
			&experience.ResumeID,
			&experience.EmployerName,
			&experience.Position,
			&experience.Duties,
			&experience.Achievements,
			&experience.StartDate,
			&endDate,
			&experience.UntilNow,
			&experience.UpdatedAt,
		); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"resumeID":  resumeID,
				"error":     err,
			}).Error("ошибка при сканировании опыта работы")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании опыта работы: %w", err),
			)
		}

		if endDate.Valid {
			experience.EndDate = endDate.Time
		}

		experiences = append(experiences, experience)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"resumeID":  resumeID,
			"error":     err,
		}).Error("ошибка при итерации по опыту работы")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по опыту работы: %w", err),
		)
	}

	return experiences, nil
}

func (r *ResumeRepository) AddSpecializations(ctx context.Context, resumeID int, specializationIDs []int) error {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на добавление специализаций AddSpecializations")

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при начале транзакции для добавления специализаций")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при начале транзакции для добавления специализаций: %w", err),
		)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				l.Log.WithFields(logrus.Fields{
					"requestID": requestID,
					"error":     rollbackErr,
				}).Error("ошибка при откате транзакции добавления специализаций")
			}
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO resume_specialization (resume_id, specialization_id)
		VALUES ($1, $2)
	`)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при подготовке запроса для добавления специализаций")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при подготовке запроса для добавления специализаций: %w", err),
		)
	}
	defer stmt.Close()

	for _, specializationID := range specializationIDs {
		_, err = stmt.ExecContext(ctx, resumeID, specializationID)
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
				"requestID":        requestID,
				"resumeID":         resumeID,
				"specializationID": specializationID,
				"error":            err,
			}).Error("ошибка при добавлении специализации к резюме")

			return entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении специализации к резюме: %w", err),
			)
		}
	}

	if err = tx.Commit(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при коммите транзакции добавления специализаций")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при коммите транзакции добавления специализаций: %w", err),
		)
	}

	return nil
}

func (r *ResumeRepository) GetSpecializationsByResumeID(ctx context.Context, resumeID int) ([]entity.Specialization, error) {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на получение специализаций GetSpecializationsByResumeID")

	query := `
		SELECT s.id, s.name
		FROM specialization s
		JOIN resume_specialization rs ON s.id = rs.specialization_id
		WHERE rs.resume_id = $1
	`

	rows, err := r.DB.QueryContext(ctx, query, resumeID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"resumeID":  resumeID,
			"error":     err,
		}).Error("ошибка при получении специализаций резюме")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении специализаций резюме: %w", err),
		)
	}
	defer rows.Close()

	var specializations []entity.Specialization
	for rows.Next() {
		var specialization entity.Specialization
		if err := rows.Scan(&specialization.ID, &specialization.Name); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"resumeID":  resumeID,
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
			"resumeID":  resumeID,
			"error":     err,
		}).Error("ошибка при итерации по специализациям")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по специализациям: %w", err),
		)
	}

	return specializations, nil
}

func (r *ResumeRepository) Update(ctx context.Context, resume *entity.Resume) (*entity.Resume, error) {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на обновление резюме Update")

	query := `
		UPDATE resume
		SET 
			about_me = $1,
			specialization_id = $2,
			education = $3,
			educational_institution = $4,
			graduation_year = $5,
			updated_at = NOW()
		WHERE id = $6 AND applicant_id = $7
		RETURNING id, applicant_id, about_me, specialization_id, education, 
				  educational_institution, graduation_year, created_at, updated_at
	`

	var updatedResume entity.Resume
	err := r.DB.QueryRowContext(
		ctx,
		query,
		resume.AboutMe,
		resume.SpecializationID,
		resume.Education,
		resume.EducationalInstitution,
		resume.GraduationYear,
		resume.ID,
		resume.ApplicantID,
	).Scan(
		&updatedResume.ID,
		&updatedResume.ApplicantID,
		&updatedResume.AboutMe,
		&updatedResume.SpecializationID,
		&updatedResume.Education,
		&updatedResume.EducationalInstitution,
		&updatedResume.GraduationYear,
		&updatedResume.CreatedAt,
		&updatedResume.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			l.Log.WithFields(logrus.Fields{
				"requestID":   requestID,
				"resumeID":    resume.ID,
				"applicantID": resume.ApplicantID,
				"error":       err,
			}).Error("резюме не найдено или не принадлежит указанному соискателю")

			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=%d не найдено или не принадлежит указанному соискателю", resume.ID),
			)
		}

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation:
				return nil, entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("резюме с такими параметрами уже существует"),
				)
			case entity.PSQLNotNullViolation:
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле отсутствует"),
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
			"requestID":   requestID,
			"resumeID":    resume.ID,
			"applicantID": resume.ApplicantID,
			"error":       err,
		}).Error("ошибка при обновлении резюме")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при обновлении резюме: %w", err),
		)
	}

	return &updatedResume, nil
}

func (r *ResumeRepository) Delete(ctx context.Context, id int) error {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на удаление резюме Delete")

	query := `
		DELETE FROM resume
		WHERE id = $1
	`

	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"resumeID":  id,
			"error":     err,
		}).Error("ошибка при удалении резюме")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при удалении резюме: %w", err),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"resumeID":  id,
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
			"resumeID":  id,
		}).Error("резюме не найдено")

		return entity.NewError(
			entity.ErrNotFound,
			fmt.Errorf("резюме с id=%d не найдено", id),
		)
	}

	return nil
}

// DeleteSkills удаляет все навыки резюме
func (r *ResumeRepository) DeleteSkills(ctx context.Context, resumeID int) error {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на удаление навыков резюме DeleteSkills")

	query := `
		DELETE FROM resume_skill
		WHERE resume_id = $1
	`

	_, err := r.DB.ExecContext(ctx, query, resumeID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"resumeID":  resumeID,
			"error":     err,
		}).Error("ошибка при удалении навыков резюме")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при удалении навыков резюме: %w", err),
		)
	}

	return nil
}

// DeleteSpecializations удаляет все специализации резюме
func (r *ResumeRepository) DeleteSpecializations(ctx context.Context, resumeID int) error {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на удаление специализаций резюме DeleteSpecializations")

	query := `
		DELETE FROM resume_specialization
		WHERE resume_id = $1
	`

	_, err := r.DB.ExecContext(ctx, query, resumeID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"resumeID":  resumeID,
			"error":     err,
		}).Error("ошибка при удалении специализаций резюме")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при удалении специализаций резюме: %w", err),
		)
	}

	return nil
}

// DeleteWorkExperiences удаляет весь опыт работы резюме
func (r *ResumeRepository) DeleteWorkExperiences(ctx context.Context, resumeID int) error {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на удаление опыта работы резюме DeleteWorkExperiences")

	query := `
		DELETE FROM work_experience
		WHERE resume_id = $1
	`

	_, err := r.DB.ExecContext(ctx, query, resumeID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"resumeID":  resumeID,
			"error":     err,
		}).Error("ошибка при удалении опыта работы резюме")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при удалении опыта работы резюме: %w", err),
		)
	}

	return nil
}

// UpdateWorkExperience обновляет запись об опыте работы
func (r *ResumeRepository) UpdateWorkExperience(ctx context.Context, workExperience *entity.WorkExperience) (*entity.WorkExperience, error) {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на обновление опыта работы UpdateWorkExperience")

	query := `
		UPDATE work_experience
		SET 
			employer_name = $1,
			position = $2,
			duties = $3,
			achievements = $4,
			start_date = $5,
			end_date = $6,
			until_now = $7,
			updated_at = NOW()
		WHERE id = $8 AND resume_id = $9
		RETURNING id, resume_id, employer_name, position, duties, 
				  achievements, start_date, end_date, until_now, updated_at
	`

	var endDate sql.NullTime
	if !workExperience.UntilNow && !workExperience.EndDate.IsZero() {
		endDate = sql.NullTime{
			Time:  workExperience.EndDate,
			Valid: true,
		}
	}

	var updatedWorkExperience entity.WorkExperience
	err := r.DB.QueryRowContext(
		ctx,
		query,
		workExperience.EmployerName,
		workExperience.Position,
		workExperience.Duties,
		workExperience.Achievements,
		workExperience.StartDate,
		endDate,
		workExperience.UntilNow,
		workExperience.ID,
		workExperience.ResumeID,
	).Scan(
		&updatedWorkExperience.ID,
		&updatedWorkExperience.ResumeID,
		&updatedWorkExperience.EmployerName,
		&updatedWorkExperience.Position,
		&updatedWorkExperience.Duties,
		&updatedWorkExperience.Achievements,
		&updatedWorkExperience.StartDate,
		&endDate,
		&updatedWorkExperience.UntilNow,
		&updatedWorkExperience.UpdatedAt,
	)

	if endDate.Valid {
		updatedWorkExperience.EndDate = endDate.Time
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			l.Log.WithFields(logrus.Fields{
				"requestID":        requestID,
				"workExperienceID": workExperience.ID,
				"resumeID":         workExperience.ResumeID,
				"error":            err,
			}).Error("запись об опыте работы не найдена или не принадлежит указанному резюме")

			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("запись об опыте работы с id=%d не найдена или не принадлежит указанному резюме", workExperience.ID),
			)
		}

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case entity.PSQLUniqueViolation:
				return nil, entity.NewError(
					entity.ErrAlreadyExists,
					fmt.Errorf("запись об опыте работы с такими параметрами уже существует"),
				)
			case entity.PSQLNotNullViolation:
				return nil, entity.NewError(
					entity.ErrBadRequest,
					fmt.Errorf("обязательное поле отсутствует"),
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
			"requestID":        requestID,
			"workExperienceID": workExperience.ID,
			"resumeID":         workExperience.ResumeID,
			"error":            err,
		}).Error("ошибка при обновлении записи об опыте работы")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при обновлении записи об опыте работы: %w", err),
		)
	}

	return &updatedWorkExperience, nil
}

// DeleteWorkExperience удаляет запись об опыте работы
func (r *ResumeRepository) DeleteWorkExperience(ctx context.Context, id int) error {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на удаление записи об опыте работы DeleteWorkExperience")

	query := `
		DELETE FROM work_experience
		WHERE id = $1
	`

	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID":        requestID,
			"workExperienceID": id,
			"error":            err,
		}).Error("ошибка при удалении записи об опыте работы")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при удалении записи об опыте работы: %w", err),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID":        requestID,
			"workExperienceID": id,
			"error":            err,
		}).Error("ошибка при получении количества затронутых строк")

		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении количества затронутых строк: %w", err),
		)
	}

	if rowsAffected == 0 {
		l.Log.WithFields(logrus.Fields{
			"requestID":        requestID,
			"workExperienceID": id,
		}).Error("запись об опыте работы не найдена")

		return entity.NewError(
			entity.ErrNotFound,
			fmt.Errorf("запись об опыте работы с id=%d не найдена", id),
		)
	}

	return nil
}

// GetAll получает список всех резюме
func (r *ResumeRepository) GetAll(ctx context.Context) ([]entity.Resume, error) {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на получение всех резюме GetAll")

	// Исправление 8: Добавлен лимит для безопасной работы с большим количеством резюме
	query := `
		SELECT id, applicant_id, about_me, specialization_id, education, 
			   educational_institution, graduation_year, created_at, updated_at
		FROM resume
		ORDER BY updated_at DESC
		LIMIT 100
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при получении списка резюме")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при получении списка резюме: %w", err),
		)
	}
	defer rows.Close()

	var resumes []entity.Resume
	for rows.Next() {
		var resume entity.Resume
		err := rows.Scan(
			&resume.ID,
			&resume.ApplicantID,
			&resume.AboutMe,
			&resume.SpecializationID,
			&resume.Education,
			&resume.EducationalInstitution,
			&resume.GraduationYear,
			&resume.CreatedAt,
			&resume.UpdatedAt,
		)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("ошибка при сканировании резюме")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании резюме: %w", err),
			)
		}
		resumes = append(resumes, resume)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при итерации по резюме")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по резюме: %w", err),
		)
	}

	return resumes, nil
}

// FindSkillIDsByNames находит ID навыков по их названиям
func (r *ResumeRepository) FindSkillIDsByNames(ctx context.Context, skillNames []string) ([]int, error) {
	requestID := middleware.GetRequestID(ctx)

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

// FindSpecializationIDByName находит ID специализации по её названию
func (r *ResumeRepository) FindSpecializationIDByName(ctx context.Context, specializationName string) (int, error) {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на поиск ID специализации по названию FindSpecializationIDByName")

	query := `
		SELECT id
		FROM specialization
		WHERE name = $1
	`

	var id int
	err := r.DB.QueryRowContext(ctx, query, specializationName).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("специализация с названием '%s' не найдена", specializationName),
			)
		}

		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"name":      specializationName,
			"error":     err,
		}).Error("ошибка при поиске ID специализации по названию")

		return 0, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при поиске ID специализации по названию: %w", err),
		)
	}

	return id, nil
}

// FindSpecializationIDsByNames находит ID специализаций по их названиям
func (r *ResumeRepository) FindSpecializationIDsByNames(ctx context.Context, specializationNames []string) ([]int, error) {
	requestID := middleware.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на поиск ID специализаций по названиям FindSpecializationIDsByNames")

	if len(specializationNames) == 0 {
		return []int{}, nil
	}

	// Создаем параметры для запроса
	params := make([]interface{}, len(specializationNames))
	placeholders := make([]string, len(specializationNames))
	for i, name := range specializationNames {
		params[i] = name
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT id
		FROM specialization
		WHERE name IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := r.DB.QueryContext(ctx, query, params...)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при поиске ID специализаций по названиям")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при поиске ID специализаций по названиям: %w", err),
		)
	}
	defer rows.Close()

	var specializationIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("ошибка при сканировании ID специализации")

			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании ID специализации: %w", err),
			)
		}
		specializationIDs = append(specializationIDs, id)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при итерации по ID специализаций")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при итерации по ID специализаций: %w", err),
		)
	}

	return specializationIDs, nil
}
