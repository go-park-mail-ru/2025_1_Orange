package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"context"
	"fmt"
)

type NotificationService struct {
	notificationRepo repository.NotificationRepository
}

func NewNotificationService(notificationRepo repository.NotificationRepository) usecase.Notification {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

func (s NotificationService) CreateNotification(ctx context.Context, notification *entity.Notification) (*entity.NotificationPreview, error) {
	if notification == nil || notification.SenderID == 0 {
		return nil, nil
	}

	if _, ok := entity.AllowedNotificationTypes[string(notification.Type)]; !ok {
		return nil, entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("неверный тип уведомления"),
		)
	}

	if notification.SenderID == notification.ReceiverID && notification.SenderRole == notification.ReceiverRole {
		return nil, nil
	}

	if err := s.notificationRepo.CreateNotification(ctx, notification); err != nil {
		return nil, err
	}

	if notification.Type == entity.ApplyNotificationType {
		return s.notificationRepo.GetApplyNotificationPreview(ctx, notification.ID)
	}
	return s.notificationRepo.GetDownloadResumeNotificationPreview(ctx, notification.ID)
}

func (s NotificationService) ReadNotification(ctx context.Context, notificationID, userID int) error {
	notification, err := s.notificationRepo.GetNotificationByID(ctx, notificationID)
	if err != nil {
		return err
	}
	if notification.IsViewed {
		return nil
	}

	if notification.ReceiverID != userID {
		return entity.NewError(
			entity.ErrForbidden,
			fmt.Errorf("нет доступа к уведомлению"),
		)
	}
	return s.notificationRepo.ReadNotification(ctx, notificationID)
}

func (s NotificationService) ReadAllNotifications(ctx context.Context, userID int, role string) error {
	return s.notificationRepo.ReadAllNotifications(ctx, userID, role)
}

func (s NotificationService) DeleteAllNotifications(ctx context.Context, userID int, role string) error {
	return s.notificationRepo.DeleteAllNotifications(ctx, userID, role)
}

func (s NotificationService) GetNotificationsForUser(ctx context.Context, userID int, role string) ([]*entity.NotificationPreview, error) {
	switch role {
	case "applicant":
		return s.notificationRepo.GetDownloadResumeNotificationsForUser(ctx, userID)
	case "employer":
		return s.notificationRepo.GetApplyNotificationsForUser(ctx, userID)
	default:
		return nil, entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("неверная роль"),
		)
	}
}
