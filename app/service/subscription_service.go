package service

import (
	"context"
	"fmt"
	"time"

	"go-messaging/entity"
	"go-messaging/repository"

	"gorm.io/gorm"
)

// SubscriptionServiceImpl implements SubscriptionService
type SubscriptionServiceImpl struct {
	subscriptionRepo     repository.SubscriptionRepository
	userRepo             repository.UserRepository
	notificationTypeRepo repository.NotificationTypeRepository
	notificationLogRepo  repository.NotificationLogRepository
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	notificationTypeRepo repository.NotificationTypeRepository,
	notificationLogRepo repository.NotificationLogRepository,
) SubscriptionService {
	return &SubscriptionServiceImpl{
		subscriptionRepo:     subscriptionRepo,
		userRepo:             userRepo,
		notificationTypeRepo: notificationTypeRepo,
		notificationLogRepo:  notificationLogRepo,
	}
}

func (s *SubscriptionServiceImpl) Subscribe(ctx context.Context, telegramUserID int64, chatID int64, notificationTypeCode string, preferences *entity.SubscriptionPreferences) (*entity.Subscription, error) {
	// Get or create user
	user, err := s.userRepo.GetByTelegramUserID(ctx, telegramUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found, please start a conversation with the bot first")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get notification type
	notificationType, err := s.notificationTypeRepo.GetByCode(ctx, notificationTypeCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification type '%s' not found", notificationTypeCode)
		}
		return nil, fmt.Errorf("failed to get notification type: %w", err)
	}

	if !notificationType.IsActive {
		return nil, fmt.Errorf("notification type '%s' is not active", notificationTypeCode)
	}

	// Check if subscription already exists
	existing, err := s.subscriptionRepo.GetByUserAndType(ctx, user.ID, notificationType.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing subscription: %w", err)
	}

	if existing != nil {
		// Update existing subscription
		existing.IsActive = true
		existing.ChatID = chatID
		if preferences != nil {
			existing.Preferences = *preferences
		}
		existing.UpdatedAt = time.Now()

		if err := s.subscriptionRepo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update subscription: %w", err)
		}
		return existing, nil
	}

	// Create new subscription
	subscription := &entity.Subscription{
		UserID:             user.ID,
		ChatID:             chatID,
		NotificationTypeID: notificationType.ID,
		IsActive:           true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if preferences != nil {
		subscription.Preferences = *preferences
	} else {
		// Set default preferences
		subscription.Preferences = entity.SubscriptionPreferences{
			Interval: notificationType.DefaultIntervalMinutes,
		}
	}

	if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return subscription, nil
}

func (s *SubscriptionServiceImpl) Unsubscribe(ctx context.Context, telegramUserID int64, notificationTypeCode string) error {
	// Get user
	user, err := s.userRepo.GetByTelegramUserID(ctx, telegramUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Get notification type
	notificationType, err := s.notificationTypeRepo.GetByCode(ctx, notificationTypeCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("notification type '%s' not found", notificationTypeCode)
		}
		return fmt.Errorf("failed to get notification type: %w", err)
	}

	// Delete subscription
	if err := s.subscriptionRepo.DeleteByUserAndType(ctx, user.ID, notificationType.ID); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

func (s *SubscriptionServiceImpl) GetUserSubscriptions(ctx context.Context, telegramUserID int64) ([]*entity.Subscription, error) {
	// Get user
	user, err := s.userRepo.GetByTelegramUserID(ctx, telegramUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []*entity.Subscription{}, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	subscriptions, err := s.subscriptionRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (s *SubscriptionServiceImpl) GetActiveSubscriptions(ctx context.Context, notificationTypeCode string) ([]*entity.Subscription, error) {
	// Get notification type
	notificationType, err := s.notificationTypeRepo.GetByCode(ctx, notificationTypeCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []*entity.Subscription{}, nil
		}
		return nil, fmt.Errorf("failed to get notification type: %w", err)
	}

	subscriptions, err := s.subscriptionRepo.GetActiveByType(ctx, notificationType.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (s *SubscriptionServiceImpl) GetDueSubscriptions(ctx context.Context, notificationTypeCode string) ([]*entity.Subscription, error) {
	// Get notification type
	notificationType, err := s.notificationTypeRepo.GetByCode(ctx, notificationTypeCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []*entity.Subscription{}, nil
		}
		return nil, fmt.Errorf("failed to get notification type: %w", err)
	}

	subscriptions, err := s.subscriptionRepo.GetDueForNotification(ctx, notificationType.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get due subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (s *SubscriptionServiceImpl) UpdatePreferences(ctx context.Context, telegramUserID int64, notificationTypeCode string, preferences *entity.SubscriptionPreferences) error {
	// Get user
	user, err := s.userRepo.GetByTelegramUserID(ctx, telegramUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Get notification type
	notificationType, err := s.notificationTypeRepo.GetByCode(ctx, notificationTypeCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("notification type '%s' not found", notificationTypeCode)
		}
		return fmt.Errorf("failed to get notification type: %w", err)
	}

	// Get subscription
	subscription, err := s.subscriptionRepo.GetByUserAndType(ctx, user.ID, notificationType.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("subscription not found")
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Update preferences
	if preferences != nil {
		subscription.Preferences = *preferences
	}
	subscription.UpdatedAt = time.Now()

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription preferences: %w", err)
	}

	return nil
}

func (s *SubscriptionServiceImpl) MarkNotified(ctx context.Context, subscriptionID int64) error {
	if err := s.subscriptionRepo.UpdateLastNotified(ctx, subscriptionID); err != nil {
		return fmt.Errorf("failed to mark subscription as notified: %w", err)
	}
	return nil
}
