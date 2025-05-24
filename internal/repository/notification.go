package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *entity.Notification) error
	GetApplyNotificationPreview(ctx context.Context, notificationID int) (*entity.NotificationPreview, error)
	GetDownloadResumeNotificationPreview(ctx context.Context, notificationID int) (*entity.NotificationPreview, error)
	GetNotificationByID(ctx context.Context, notificationID int) (*entity.Notification, error)
	GetApplyNotificationsForUser(ctx context.Context, notificationID int) ([]*entity.NotificationPreview, error)
	GetDownloadResumeNotificationsForUser(ctx context.Context, notificationID int) ([]*entity.NotificationPreview, error)
	ReadNotification(ctx context.Context, notificationID int) error
	ReadAllNotifications(ctx context.Context, userID int, role string) error
	DeleteAllNotifications(ctx context.Context, userID int, role string) error
}
