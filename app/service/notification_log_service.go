package service

import (
	"context"
	"fmt"
	"time"

	"go-messaging/entity"
	"go-messaging/repository"

	"gorm.io/gorm"
)

// NotificationLogServiceImpl implements NotificationLogService
type NotificationLogServiceImpl struct {
	notificationLogRepo repository.NotificationLogRepository
}

// NewNotificationLogService creates a new notification log service
func NewNotificationLogService(notificationLogRepo repository.NotificationLogRepository) NotificationLogService {
	return &NotificationLogServiceImpl{
		notificationLogRepo: notificationLogRepo,
	}
}

func (s *NotificationLogServiceImpl) LogNotification(ctx context.Context, subscriptionID int64, message, status string, errorMessage *string) (*entity.NotificationLog, error) {
	log := &entity.NotificationLog{
		SubscriptionID: subscriptionID,
		Message:        message,
		Status:         status,
		SentAt:         time.Now(),
		ErrorMessage:   errorMessage,
	}

	if err := s.notificationLogRepo.Create(ctx, log); err != nil {
		return nil, fmt.Errorf("failed to create notification log: %w", err)
	}

	return log, nil
}

func (s *NotificationLogServiceImpl) GetSubscriptionLogs(ctx context.Context, subscriptionID int64, offset, limit int) ([]*entity.NotificationLog, error) {
	logs, err := s.notificationLogRepo.GetBySubscriptionID(ctx, subscriptionID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription logs: %w", err)
	}
	return logs, nil
}

func (s *NotificationLogServiceImpl) GetRecentLogs(ctx context.Context, limit int) ([]*entity.NotificationLog, error) {
	logs, err := s.notificationLogRepo.GetRecentLogs(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent logs: %w", err)
	}
	return logs, nil
}

func (s *NotificationLogServiceImpl) UpdateLogStatus(ctx context.Context, logID int64, status string, errorMessage *string) error {
	log, err := s.notificationLogRepo.GetByID(ctx, logID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("notification log not found")
		}
		return fmt.Errorf("failed to get notification log: %w", err)
	}

	log.Status = status
	log.ErrorMessage = errorMessage

	if err := s.notificationLogRepo.Update(ctx, log); err != nil {
		return fmt.Errorf("failed to update notification log: %w", err)
	}

	return nil
}

func (s *NotificationLogServiceImpl) CleanupOldLogs(ctx context.Context, daysOld int) error {
	if err := s.notificationLogRepo.CleanupOldLogs(ctx, daysOld); err != nil {
		return fmt.Errorf("failed to cleanup old logs: %w", err)
	}
	return nil
}
