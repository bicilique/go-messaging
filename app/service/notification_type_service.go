package service

import (
	"context"
	"fmt"
	"time"

	"go-messaging/entity"
	"go-messaging/repository"

	"gorm.io/gorm"
)

// NotificationTypeServiceImpl implements NotificationTypeService
type NotificationTypeServiceImpl struct {
	notificationTypeRepo repository.NotificationTypeRepository
}

// NewNotificationTypeService creates a new notification type service
func NewNotificationTypeService(notificationTypeRepo repository.NotificationTypeRepository) NotificationTypeService {
	return &NotificationTypeServiceImpl{
		notificationTypeRepo: notificationTypeRepo,
	}
}

func (s *NotificationTypeServiceImpl) GetAllTypes(ctx context.Context) ([]*entity.NotificationType, error) {
	types, err := s.notificationTypeRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification types: %w", err)
	}
	return types, nil
}

func (s *NotificationTypeServiceImpl) GetActiveTypes(ctx context.Context) ([]*entity.NotificationType, error) {
	types, err := s.notificationTypeRepo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active notification types: %w", err)
	}
	return types, nil
}

func (s *NotificationTypeServiceImpl) GetTypeByCode(ctx context.Context, code string) (*entity.NotificationType, error) {
	notificationType, err := s.notificationTypeRepo.GetByCode(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification type '%s' not found", code)
		}
		return nil, fmt.Errorf("failed to get notification type: %w", err)
	}
	return notificationType, nil
}

func (s *NotificationTypeServiceImpl) GetTypeByID(ctx context.Context, id int) (*entity.NotificationType, error) {
	notificationType, err := s.notificationTypeRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification type with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get notification type: %w", err)
	}
	return notificationType, nil
}

func (s *NotificationTypeServiceImpl) CreateType(ctx context.Context, code, name string, description *string, defaultInterval int) (*entity.NotificationType, error) {
	// Check if type with code already exists
	_, err := s.notificationTypeRepo.GetByCode(ctx, code)
	if err == nil {
		return nil, fmt.Errorf("notification type with code '%s' already exists", code)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing notification type: %w", err)
	}

	notificationType := &entity.NotificationType{
		Code:                   code,
		Name:                   name,
		Description:            description,
		DefaultIntervalMinutes: defaultInterval,
		IsActive:               true,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := s.notificationTypeRepo.Create(ctx, notificationType); err != nil {
		return nil, fmt.Errorf("failed to create notification type: %w", err)
	}

	return notificationType, nil
}

func (s *NotificationTypeServiceImpl) UpdateType(ctx context.Context, notificationType *entity.NotificationType) error {
	notificationType.UpdatedAt = time.Now()
	if err := s.notificationTypeRepo.Update(ctx, notificationType); err != nil {
		return fmt.Errorf("failed to update notification type: %w", err)
	}
	return nil
}

func (s *NotificationTypeServiceImpl) ActivateType(ctx context.Context, code string) error {
	notificationType, err := s.notificationTypeRepo.GetByCode(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("notification type '%s' not found", code)
		}
		return fmt.Errorf("failed to get notification type: %w", err)
	}

	notificationType.IsActive = true
	notificationType.UpdatedAt = time.Now()

	if err := s.notificationTypeRepo.Update(ctx, notificationType); err != nil {
		return fmt.Errorf("failed to activate notification type: %w", err)
	}

	return nil
}

func (s *NotificationTypeServiceImpl) DeactivateType(ctx context.Context, code string) error {
	notificationType, err := s.notificationTypeRepo.GetByCode(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("notification type '%s' not found", code)
		}
		return fmt.Errorf("failed to get notification type: %w", err)
	}

	notificationType.IsActive = false
	notificationType.UpdatedAt = time.Now()

	if err := s.notificationTypeRepo.Update(ctx, notificationType); err != nil {
		return fmt.Errorf("failed to deactivate notification type: %w", err)
	}

	return nil
}
