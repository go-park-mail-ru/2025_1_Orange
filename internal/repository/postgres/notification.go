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
	"github.com/sirupsen/logrus"
)

type NotificationRepository struct {
	DB *sql.DB
}

func NewNotificationRepository(db *sql.DB) (repository.NotificationRepository, error) {
	return &NotificationRepository{DB: db}, nil
}

func (r *NotificationRepository) GetApplyNotificationPreview(ctx context.Context, notificationID int) (*entity.NotificationPreview, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("Выполнение sql-запроса получения уведомлений откликов GetApplyNotificationPreview")

	query := `
		SELECT 
			n.id,
			n.type,
			n.sender_id,
			n.receiver_id,
			n.object_id,
			n.is_viewed,
			n.created_at,
			a.first_name AS applicant_name,
			e.company_name AS employer_name,
			v.title
		FROM notification n
		LEFT JOIN applicant a ON n.sender_id = a.id
		LEFT JOIN employer e ON n.receiver_id = e.id
		LEFT JOIN vacancy v ON n.object_id = v.id
		WHERE n.id = $1 AND n.type = 'apply'
	`

	var preview entity.NotificationPreview

	err := r.DB.QueryRowContext(ctx, query, notificationID).Scan(
		&preview.ID,
		&preview.Type,
		&preview.SenderID,
		&preview.ReceiverID,
		&preview.ObjectID,
		&preview.IsViewed,
		&preview.CreatedAt,
		&preview.ApplicantName,
		&preview.EmployerName,
		&preview.Title,
	)

	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID":      requestID,
			"notificationID": notificationID,
			"error":          err,
		}).Error("Ошибка при получении уведомления")
		return nil, entity.NewError(
			entity.ErrNotFound,
			fmt.Errorf("ошибка при получении уведомления: %v", err),
		)
	}

	return &preview, nil
}

func (r *NotificationRepository) GetDownloadResumeNotificationPreview(ctx context.Context, notificationID int) (*entity.NotificationPreview, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":      requestID,
		"notificationID": notificationID,
	}).Info("Выполнение sql-запроса получения уведомлений о скачивании резюме GetDownloadResumeNotificationPreview")

	query := `
		SELECT 
			n.id,
			n.type,
			n.sender_id,
			n.receiver_id,
			n.object_id,
			n.is_viewed,
			n.created_at,
			a.first_name AS applicant_name,
			e.company_name AS employer_name,
			r.profession AS title
		FROM notification n
		LEFT JOIN resume r ON r.id = n.object_id
		LEFT JOIN applicant a ON a.id = r.applicant_id
		LEFT JOIN employer e ON e.id = n.sender_id
		WHERE n.id = $1 AND n.type = 'download_resume'
	`

	var preview entity.NotificationPreview
	err := r.DB.QueryRowContext(ctx, query, notificationID).Scan(
		&preview.ID,
		&preview.Type,
		&preview.SenderID,
		&preview.ReceiverID,
		&preview.ObjectID,
		&preview.IsViewed,
		&preview.CreatedAt,
		&preview.ApplicantName,
		&preview.EmployerName,
		&preview.Title,
	)

	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("Ошибка при выполнении запроса GetDownloadResumeNotificationPreview")
		return nil, entity.NewError(
			entity.ErrNotFound,
			fmt.Errorf("ошибка при выполнении запроса GetDownloadResumeNotificationPreview: %v", err),
		)
	}

	return &preview, nil
}

func (r *NotificationRepository) GetNotificationByID(ctx context.Context, notificationID int) (*entity.Notification, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":      requestID,
		"notificationID": notificationID,
	}).Info("Выполнение SQL-запроса получения уведомления по ID")

	query := `
		SELECT 
			id,
			type,
			sender_id,
			receiver_id,
			object_id,
			is_viewed,
			created_at
		FROM notification
		WHERE id = $1
	`

	row := r.DB.QueryRowContext(ctx, query, notificationID)

	var n entity.Notification
	err := row.Scan(
		&n.ID,
		&n.Type,
		&n.SenderID,
		&n.ReceiverID,
		&n.ObjectID,
		&n.IsViewed,
		&n.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
			}).Warn("Уведомление не найдено")
			return nil, entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("уведомление не найдено"),
			)
		}
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("Ошибка при выполнении SQL-запроса")
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при выполнении SQL-запроса: %v", err),
		)
	}

	return &n, nil
}

func (r *NotificationRepository) CreateNotification(ctx context.Context, notification *entity.Notification) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("Выполнение sql-запроса создания уведомления CreateNotification")

	query := `
		INSERT INTO notification (
			type,
			sender_id,
			receiver_id,
			object_id,
			is_viewed
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.DB.QueryRow(
		query,
		notification.Type,
		notification.SenderID,
		notification.ReceiverID,
		notification.ObjectID,
		notification.IsViewed,
	).Scan(&notification.ID)

	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"sender_id":    notification.SenderID,
			"receiver_id":  notification.ReceiverID,
			"object_id":    notification.ObjectID,
			"notification": notification.Type,
			"error":        err,
		}).Error("Ошибка при создании уведомления")
		return entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("ошибка при создании уведомления: %v", err),
		)
	}

	l.Log.WithFields(logrus.Fields{
		"notification_id": notification.ID,
		"type":            notification.Type,
		"receiver_id":     notification.ReceiverID,
	}).Info("Уведомление успешно создано")

	return nil
}

func (r *NotificationRepository) ReadNotification(ctx context.Context, notificationID int) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":      requestID,
		"notificationID": notificationID,
	}).Info("Выполнение sql-запроса ReadNotification")

	query := `
		UPDATE notification
		SET is_viewed = true
		WHERE id = $1
	`

	result, err := r.DB.ExecContext(ctx, query, notificationID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err.Error(),
		}).Error("Ошибка при выполнении ReadNotification")
		return entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("ошибка при выполнении ReadNotification: %v", err),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err.Error(),
		}).Error("Не удалось получить число затронутых строк в ReadNotification")
		return entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("не удалось получить число затронутых строк в ReadNotification: %v", err),
		)
	}

	if rowsAffected == 0 {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
		}).Warn("Уведомление не найдено в ReadNotification")
		return entity.NewError(
			entity.ErrNotFound,
			fmt.Errorf("уведомление не найдено в ReadNotification"),
		)
	}

	return nil
}

func (r *NotificationRepository) ReadAllNotifications(ctx context.Context, userID int, role string) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"userID":    userID,
		"role":      role,
	}).Info("Выполнение sql-запроса ReadAllNotifications")

	var query string
	switch role {
	case "applicant":
		query = `
			UPDATE notification
			SET is_viewed = true
			WHERE receiver_id = $1 AND type = 'download_resume'
		`
	case "employer":
		query = `
			UPDATE notification
			SET is_viewed = true
			WHERE receiver_id = $1 AND type = 'apply'
		`
	default:
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"role":      role,
		}).Warn("Неподдерживаемая роль в ReadAllNotifications")
		return entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("несуществующая роль: %s", role),
		)
	}

	_, err := r.DB.ExecContext(ctx, query, userID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err.Error(),
		}).Error("Ошибка при выполнении ReadAllNotifications")
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при выполнении ReadAllNotifications: %v", err),
		)
	}

	return nil
}

func (r *NotificationRepository) DeleteAllNotifications(ctx context.Context, userID int, role string) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"userID":    userID,
		"role":      role,
	}).Info("Выполнение sql-запроса DeleteAllNotifications")

	var query string
	switch role {
	case "applicant":
		query = `
			DELETE FROM notification
			WHERE receiver_id = $1 AND type = 'download_resume'
		`
	case "employer":
		query = `
			DELETE FROM notification
			WHERE receiver_id = $1 AND type = 'apply'
		`
	default:
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"role":      role,
		}).Warn("Неподдерживаемая роль в DeleteAllNotifications")
		return entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("несуществующая роль: %s", role),
		)
	}

	_, err := r.DB.ExecContext(ctx, query, userID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err.Error(),
		}).Error("Ошибка при выполнении DeleteAllNotifications")
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при выполнении DeleteAllNotifications: %v", err),
		)
	}

	return nil
}

func (r *NotificationRepository) GetApplyNotificationsForUser(ctx context.Context, userID int) ([]*entity.NotificationPreview, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"userID":    userID,
	}).Info("Выполнение sql-запроса получения всех уведомлений откликов для пользователя GetApplyNotificationsForUser")

	query := `
		SELECT 
			n.id,
			n.type,
			n.sender_id,
			n.receiver_id,
			n.object_id,
			n.is_viewed,
			n.created_at,
			a.first_name AS applicant_name,
			e.company_name AS employer_name,
			v.title
		FROM notification n
		LEFT JOIN applicant a ON n.sender_id = a.id
		LEFT JOIN employer e ON n.receiver_id = e.id
		LEFT JOIN vacancy v ON n.object_id = v.id
		WHERE n.type = 'apply' AND n.receiver_id = $1
		ORDER BY n.created_at DESC
	`

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"userID":    userID,
			"error":     err,
		}).Error("Ошибка при выполнении запроса GetApplyNotificationsForUser")
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при выполнении запроса GetApplyNotificationsForUser: %v", err),
		)
	}
	defer rows.Close()

	var notifications []*entity.NotificationPreview

	for rows.Next() {
		var preview entity.NotificationPreview
		err := rows.Scan(
			&preview.ID,
			&preview.Type,
			&preview.SenderID,
			&preview.ReceiverID,
			&preview.ObjectID,
			&preview.IsViewed,
			&preview.CreatedAt,
			&preview.ApplicantName,
			&preview.EmployerName,
			&preview.Title,
		)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("Ошибка при сканировании строки результата")
			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании строки результата запроса GetApplyNotificationsForUser: %v", err),
			)
		}
		notifications = append(notifications, &preview)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("Ошибка при обходе результатов запроса")
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при обходе результатов запроса GetApplyNotificationsForUser: %v", err),
		)
	}

	return notifications, nil
}

func (r *NotificationRepository) GetDownloadResumeNotificationsForUser(ctx context.Context, userID int) ([]*entity.NotificationPreview, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"userID":    userID,
	}).Info("Выполнение sql-запроса получения всех уведомлений о скачивании резюме для пользователя")

	query := `
		SELECT 
			n.id,
			n.type,
			n.sender_id,
			n.receiver_id,
			n.object_id,
			n.is_viewed,
			n.created_at,
			a.first_name AS applicant_name,
			e.company_name AS employer_name,
			r.profession AS title
		FROM notification n
		LEFT JOIN resume r ON r.id = n.object_id
		LEFT JOIN applicant a ON a.id = r.applicant_id
		LEFT JOIN employer e ON e.id = n.sender_id
		WHERE n.receiver_id = $1 AND n.type = 'download_resume'
		ORDER BY n.created_at DESC
	`

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("Ошибка при выполнении запроса GetDownloadResumeNotificationsForUser")
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при выполнении запроса GetDownloadResumeNotificationsForUser: %v", err),
		)
	}
	defer rows.Close()

	var notifications []*entity.NotificationPreview

	for rows.Next() {
		var preview entity.NotificationPreview
		err := rows.Scan(
			&preview.ID,
			&preview.Type,
			&preview.SenderID,
			&preview.ReceiverID,
			&preview.ObjectID,
			&preview.IsViewed,
			&preview.CreatedAt,
			&preview.ApplicantName,
			&preview.EmployerName,
			&preview.Title,
		)
		if err != nil {
			l.Log.WithFields(logrus.Fields{
				"requestID": requestID,
				"error":     err,
			}).Error("Ошибка при сканировании результата GetDownloadResumeNotificationsForUser")
			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании результата запроса GetDownloadResumeNotificationsForUser: %v", err),
			)
		}
		notifications = append(notifications, &preview)
	}

	if err := rows.Err(); err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка после итерации по строкам GetDownloadResumeNotificationsForUser")
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка после итерации по строкам запроса GetDownloadResumeNotificationsForUser: %v", err),
		)
	}

	return notifications, nil
}
