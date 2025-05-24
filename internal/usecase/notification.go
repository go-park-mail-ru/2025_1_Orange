package usecase

import (
	"ResuMatch/internal/entity"
	"context"
)

type Notification interface {
	CreateNotification(ctx context.Context, notification *entity.Notification) (*entity.NotificationPreview, error)
	ReadNotification(ctx context.Context, notificationID, userID int) error
	ReadAllNotifications(ctx context.Context, userID int, role string) error
	DeleteAllNotifications(ctx context.Context, userID int, role string) error
	GetNotificationsForUser(ctx context.Context, userID int, role string) ([]*entity.NotificationPreview, error)
}
