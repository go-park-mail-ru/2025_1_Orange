package postgres

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/middleware"
	"ResuMatch/internal/repository"
	l "ResuMatch/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type ResumeRepository struct {
	DB *sql.DB
}

func NewResumeRepository(cfg config.PostgresConfig) (repository.ResumeRepository, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось установить соединение с PostgreSQL из ResumeRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить соединение PostgreSQL из ResumeRepository: %w", err),
		)
	}

	if err := db.Ping(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось выполнить ping PostgreSQL из ResumeRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось выполнить ping PostgreSQL из ResumeRepository: %w", err),
		)
	}
	return &ResumeRepository{DB: db}, nil
}

func (r *ResumeRepository) Create(ctx context.Context, resume *entity.Resume) (*entity.Resume, error) {
	requestID := middleware.GetRequestID(ctx)

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

	query := `
		INSERT INTO work_experience (
			resume_id, employer_name, position, duties, 
			achievements, start_date, end_date, until_now
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, resume_id, employer_name, position, duties, 
				  achievements, start_date, end_date, until_now
	`

	var endDate sql.NullTime
	if !workExperience.UntilNow {
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

	query := `
		SELECT id, resume_id, employer_name, position, duties, 
			   achievements, start_date, end_date, until_now
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

// Добавим новый метод в ResumeRepository
func (r *ResumeRepository) AddSpecializations(ctx context.Context, resumeID int, specializationIDs []int) error {
	requestID := middleware.GetRequestID(ctx)

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
