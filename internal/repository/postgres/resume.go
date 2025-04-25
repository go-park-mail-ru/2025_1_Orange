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

	// "strings"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type ResumeRepository struct {
	DB *sql.DB
}

func NewResumeRepository(db *sql.DB) (repository.ResumeRepository, error) {
	return &ResumeRepository{DB: db}, nil
}

func (r *ResumeRepository) Create(ctx context.Context, resume *entity.Resume) (*entity.Resume, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на создание резюме Create")

	query := `
		INSERT INTO resume (
			applicant_id, about_me, specialization_id, education, 
			educational_institution, graduation_year, profession, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, applicant_id, about_me, specialization_id, education, 
				  educational_institution, graduation_year, profession, created_at, updated_at
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
<<<<<<< HEAD
		resume.Profession,
=======
		resume.Profession, // Дополнение - добавлено поле профессии
>>>>>>> e897aad (добавил к резюме поле профессии)
	).Scan(
		&createdResume.ID,
		&createdResume.ApplicantID,
		&createdResume.AboutMe,
		&createdResume.SpecializationID,
		&createdResume.Education,
		&createdResume.EducationalInstitution,
		&createdResume.GraduationYear,
<<<<<<< HEAD
		&createdResume.Profession,
=======
		&createdResume.Profession, // Дополнение - добавлено поле профессии
>>>>>>> e897aad (добавил к резюме поле профессии)
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
	requestID := utils.GetRequestID(ctx)

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
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть statement: %v", err)
		}
	}(stmt)

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
	requestID := utils.GetRequestID(ctx)

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
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на получение резюме по ID GetByID")

	query := `
		SELECT id, applicant_id, about_me, specialization_id, education, 
			   educational_institution, graduation_year, profession, created_at, updated_at
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
<<<<<<< HEAD
		&resume.Profession,
=======
		&resume.Profession, // Дополнение - добавлено поле профессии
>>>>>>> e897aad (добавил к резюме поле профессии)
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
	requestID := utils.GetRequestID(ctx)

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
	requestID := utils.GetRequestID(ctx)

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

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

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
	requestID := utils.GetRequestID(ctx)

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

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть statement: %v", err)
		}
	}(stmt)

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
	requestID := utils.GetRequestID(ctx)

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

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

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
	requestID := utils.GetRequestID(ctx)

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
			profession = $6,
			updated_at = NOW()
		WHERE id = $7 AND applicant_id = $8
		RETURNING id, applicant_id, about_me, specialization_id, education, 
				  educational_institution, graduation_year, profession, created_at, updated_at
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
<<<<<<< HEAD
		resume.Profession,
=======
		resume.Profession, // Дополнение - добавлено поле профессии
>>>>>>> e897aad (добавил к резюме поле профессии)
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
<<<<<<< HEAD
		&updatedResume.Profession,
=======
		&updatedResume.Profession, // Дополнение - добавлено поле профессии
>>>>>>> e897aad (добавил к резюме поле профессии)
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
	requestID := utils.GetRequestID(ctx)

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
	requestID := utils.GetRequestID(ctx)

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
	requestID := utils.GetRequestID(ctx)

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
	requestID := utils.GetRequestID(ctx)

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
	requestID := utils.GetRequestID(ctx)

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
	requestID := utils.GetRequestID(ctx)

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
func (r *ResumeRepository) GetAll(ctx context.Context, limit int, offset int) ([]entity.Resume, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на получение всех резюме GetAll")

	query := `
		SELECT id, applicant_id, about_me, specialization_id, education, 
			   educational_institution, graduation_year, profession, created_at, updated_at
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> 336f233 (добавил пагинацию в методы списка резюме)
		FROM resume
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`
<<<<<<< HEAD
=======
	FROM resume
	ORDER BY updated_at DESC
	LIMIT 100
`
>>>>>>> e897aad (добавил к резюме поле профессии)
=======
>>>>>>> 336f233 (добавил пагинацию в методы списка резюме)

	rows, err := r.DB.QueryContext(ctx, query, limit, offset)
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

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

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
<<<<<<< HEAD
			&resume.Profession,
=======
			&resume.Profession, // Дополнение - добавлено поле профессии
>>>>>>> e897aad (добавил к резюме поле профессии)
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

// GetAllResumesByApplicantID получает список всех резюме одного соискателя
func (r *ResumeRepository) GetAllResumesByApplicantID(ctx context.Context, applicantID int, limit int, offset int) ([]entity.Resume, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":   requestID,
		"applicantID": applicantID,
	}).Info("sql-запрос в БД на получение всех резюме соискателя GetAllResumesByApplicantID")

	query := `
		SELECT id, applicant_id, about_me, specialization_id, education, 
			   educational_institution, graduation_year, profession, created_at, updated_at
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> 336f233 (добавил пагинацию в методы списка резюме)
		FROM resume
		WHERE applicant_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`
<<<<<<< HEAD
=======
	FROM resume
	WHERE applicant_id = $1
	ORDER BY updated_at DESC
	LIMIT 100
`
>>>>>>> e897aad (добавил к резюме поле профессии)
=======
>>>>>>> 336f233 (добавил пагинацию в методы списка резюме)

	rows, err := r.DB.QueryContext(ctx, query, applicantID, limit, offset)
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

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Errorf("не удалось закрыть rows: %v", err)
		}
	}(rows)

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
<<<<<<< HEAD
			&resume.Profession,
=======
			&resume.Profession, // Дополнение - добавлено поле профессии
>>>>>>> e897aad (добавил к резюме поле профессии)
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

// // FindSkillIDsByNames находит ID навыков по их названиям
// func (r *ResumeRepository) FindSkillIDsByNames(ctx context.Context, skillNames []string) ([]int, error) {
// 	requestID := utils.GetRequestID(ctx)

// 	l.Log.WithFields(logrus.Fields{
// 		"requestID": requestID,
// 	}).Info("sql-запрос в БД на поиск ID навыков по названиям FindSkillIDsByNames")

// 	if len(skillNames) == 0 {
// 		return []int{}, nil
// 	}

// 	params := make([]interface{}, len(skillNames))
// 	placeholders := make([]string, len(skillNames))
// 	for i, name := range skillNames {
// 		params[i] = name
// 		placeholders[i] = fmt.Sprintf("$%d", i+1)
// 	}

// 	query := fmt.Sprintf(`
// 		SELECT id
// 		FROM skill
// 		WHERE name IN (%s)
// 	`, strings.Join(placeholders, ", "))

// 	rows, err := r.DB.QueryContext(ctx, query, params...)
// 	if err != nil {
// 		l.Log.WithFields(logrus.Fields{
// 			"requestID": requestID,
// 			"error":     err,
// 		}).Error("ошибка при поиске ID навыков по названиям")

// 		return nil, entity.NewError(
// 			entity.ErrInternal,
// 			fmt.Errorf("ошибка при поиске ID навыков по названиям: %w", err),
// 		)
// 	}
// 	defer rows.Close()

// 	var skillIDs []int
// 	for rows.Next() {
// 		var id int
// 		if err := rows.Scan(&id); err != nil {
// 			l.Log.WithFields(logrus.Fields{
// 				"requestID": requestID,
// 				"error":     err,
// 			}).Error("ошибка при сканировании ID навыка")

// 			return nil, entity.NewError(
// 				entity.ErrInternal,
// 				fmt.Errorf("ошибка при сканировании ID навыка: %w", err),
// 			)
// 		}
// 		skillIDs = append(skillIDs, id)
// 	}

// 	if err := rows.Err(); err != nil {
// 		l.Log.WithFields(logrus.Fields{
// 			"requestID": requestID,
// 			"error":     err,
// 		}).Error("ошибка при итерации по ID навыков")

// 		return nil, entity.NewError(
// 			entity.ErrInternal,
// 			fmt.Errorf("ошибка при итерации по ID навыков: %w", err),
// 		)
// 	}

// 	return skillIDs, nil
// }

// FindSkillIDsByNames находит ID навыков по их названиям, создавая новые при необходимости
func (r *ResumeRepository) FindSkillIDsByNames(ctx context.Context, skillNames []string) ([]int, error) {
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

// // FindSpecializationIDByName находит ID специализации по её названию
// func (r *ResumeRepository) FindSpecializationIDByName(ctx context.Context, specializationName string) (int, error) {
// 	requestID := utils.GetRequestID(ctx)

// 	l.Log.WithFields(logrus.Fields{
// 		"requestID": requestID,
// 	}).Info("sql-запрос в БД на поиск ID специализации по названию FindSpecializationIDByName")

// 	query := `
// 		SELECT id
// 		FROM specialization
// 		WHERE name = $1
// 	`

// 	var id int
// 	err := r.DB.QueryRowContext(ctx, query, specializationName).Scan(&id)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return 0, entity.NewError(
// 				entity.ErrNotFound,
// 				fmt.Errorf("специализация с названием '%s' не найдена", specializationName),
// 			)
// 		}

// 		l.Log.WithFields(logrus.Fields{
// 			"requestID": requestID,
// 			"name":      specializationName,
// 			"error":     err,
// 		}).Error("ошибка при поиске ID специализации по названию")

// 		return 0, entity.NewError(
// 			entity.ErrInternal,
// 			fmt.Errorf("ошибка при поиске ID специализации по названию: %w", err),
// 		)
// 	}

// 	return id, nil
// }

// FindSpecializationIDByName находит ID специализации по её названию, создавая новую при необходимости
func (r *ResumeRepository) FindSpecializationIDByName(ctx context.Context, specializationName string) (int, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на поиск ID специализации по названию FindSpecializationIDByName")

	return r.CreateSpecializationIfNotExists(ctx, specializationName)
}

// // FindSpecializationIDsByNames находит ID специализаций по их названиям
// func (r *ResumeRepository) FindSpecializationIDsByNames(ctx context.Context, specializationNames []string) ([]int, error) {
// 	requestID := utils.GetRequestID(ctx)

// 	l.Log.WithFields(logrus.Fields{
// 		"requestID": requestID,
// 	}).Info("sql-запрос в БД на поиск ID специализаций по названиям FindSpecializationIDsByNames")

// 	if len(specializationNames) == 0 {
// 		return []int{}, nil
// 	}

// 	// Создаем параметры для запроса
// 	params := make([]interface{}, len(specializationNames))
// 	placeholders := make([]string, len(specializationNames))
// 	for i, name := range specializationNames {
// 		params[i] = name
// 		placeholders[i] = fmt.Sprintf("$%d", i+1)
// 	}

// 	query := fmt.Sprintf(`
// 		SELECT id
// 		FROM specialization
// 		WHERE name IN (%s)
// 	`, strings.Join(placeholders, ", "))

// 	rows, err := r.DB.QueryContext(ctx, query, params...)
// 	if err != nil {
// 		l.Log.WithFields(logrus.Fields{
// 			"requestID": requestID,
// 			"error":     err,
// 		}).Error("ошибка при поиске ID специализаций по названиям")

// 		return nil, entity.NewError(
// 			entity.ErrInternal,
// 			fmt.Errorf("ошибка при поиске ID специализаций по названиям: %w", err),
// 		)
// 	}
// 	defer rows.Close()

// 	var specializationIDs []int
// 	for rows.Next() {
// 		var id int
// 		if err := rows.Scan(&id); err != nil {
// 			l.Log.WithFields(logrus.Fields{
// 				"requestID": requestID,
// 				"error":     err,
// 			}).Error("ошибка при сканировании ID специализации")

// 			return nil, entity.NewError(
// 				entity.ErrInternal,
// 				fmt.Errorf("ошибка при сканировании ID специализации: %w", err),
// 			)
// 		}
// 		specializationIDs = append(specializationIDs, id)
// 	}

// 	if err := rows.Err(); err != nil {
// 		l.Log.WithFields(logrus.Fields{
// 			"requestID": requestID,
// 			"error":     err,
// 		}).Error("ошибка при итерации по ID специализаций")

// 		return nil, entity.NewError(
// 			entity.ErrInternal,
// 			fmt.Errorf("ошибка при итерации по ID специализаций: %w", err),
// 		)
// 	}

// 	return specializationIDs, nil
// }

// FindSpecializationIDsByNames находит ID специализаций по их названиям, создавая новые при необходимости
func (r *ResumeRepository) FindSpecializationIDsByNames(ctx context.Context, specializationNames []string) ([]int, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("sql-запрос в БД на поиск ID специализаций по названиям FindSpecializationIDsByNames")

	if len(specializationNames) == 0 {
		return []int{}, nil
	}

	var specializationIDs []int

	// Для каждой специализации проверяем её существование и создаем при необходимости
	for _, name := range specializationNames {
		id, err := r.CreateSpecializationIfNotExists(ctx, name)
		if err != nil {
			return nil, err
		}
		specializationIDs = append(specializationIDs, id)
	}

	return specializationIDs, nil
}

// CreateSkillIfNotExists создает новый навык, если он не существует
func (r *ResumeRepository) CreateSkillIfNotExists(ctx context.Context, skillName string) (int, error) {
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

// CreateSpecializationIfNotExists создает новую специализацию, если она не существует
func (r *ResumeRepository) CreateSpecializationIfNotExists(ctx context.Context, specializationName string) (int, error) {
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

// SearchResumesByProfession ищет резюме по профессии
func (r *ResumeRepository) SearchResumesByProfession(ctx context.Context, profession string, limit int, offset int) ([]entity.Resume, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"profession": profession,
	}).Info("sql-запрос в БД на поиск резюме по профессии SearchResumesByProfession")

	query := `
        SELECT id, applicant_id, about_me, specialization_id, education, 
               educational_institution, graduation_year, profession, created_at, updated_at
        FROM resume
        WHERE profession ILIKE $1
        ORDER BY updated_at DESC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.DB.QueryContext(ctx, query, "%"+profession+"%", limit, offset)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при поиске резюме по профессии")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при поиске резюме по профессии: %w", err),
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
			&resume.Profession,
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

// SearchResumesByProfessionForApplicant ищет резюме по профессии для конкретного соискателя
func (r *ResumeRepository) SearchResumesByProfessionForApplicant(ctx context.Context, applicantID int, profession string, limit int, offset int) ([]entity.Resume, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":   requestID,
		"applicantID": applicantID,
		"profession":  profession,
	}).Info("sql-запрос в БД на поиск резюме по профессии для соискателя SearchResumesByProfessionForApplicant")

	query := `
        SELECT id, applicant_id, about_me, specialization_id, education, 
               educational_institution, graduation_year, profession, created_at, updated_at
        FROM resume
        WHERE applicant_id = $1 AND profession ILIKE $2
        ORDER BY updated_at DESC
        LIMIT $3 OFFSET $4
    `

	rows, err := r.DB.QueryContext(ctx, query, applicantID, "%"+profession+"%", limit, offset)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при поиске резюме по профессии для соискателя")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при поиске резюме по профессии для соискателя: %w", err),
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
			&resume.Profession,
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
