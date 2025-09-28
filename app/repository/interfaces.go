package repository

import (
	"context"
	"go-messaging/entity"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// GetByTelegramUserID retrieves a user by Telegram user ID
	GetByTelegramUserID(ctx context.Context, telegramUserID int64) (*entity.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *entity.User) error

	// Delete deletes a user by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves all users with pagination
	List(ctx context.Context, offset, limit int) ([]*entity.User, error)
}

// NotificationTypeRepository defines the interface for notification type data access
type NotificationTypeRepository interface {
	// GetAll retrieves all notification types
	GetAll(ctx context.Context) ([]*entity.NotificationType, error)

	// GetByID retrieves a notification type by ID
	GetByID(ctx context.Context, id int) (*entity.NotificationType, error)

	// GetByCode retrieves a notification type by code
	GetByCode(ctx context.Context, code string) (*entity.NotificationType, error)

	// GetActive retrieves all active notification types
	GetActive(ctx context.Context) ([]*entity.NotificationType, error)

	// Create creates a new notification type
	Create(ctx context.Context, notificationType *entity.NotificationType) error

	// Update updates an existing notification type
	Update(ctx context.Context, notificationType *entity.NotificationType) error

	// Delete deletes a notification type by ID
	Delete(ctx context.Context, id int) error
}

// SubscriptionRepository defines the interface for subscription data access
type SubscriptionRepository interface {
	// Create creates a new subscription
	Create(ctx context.Context, subscription *entity.Subscription) error

	// GetByID retrieves a subscription by ID
	GetByID(ctx context.Context, id int64) (*entity.Subscription, error)

	// GetByUserID retrieves all subscriptions for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Subscription, error)

	// GetByUserAndType retrieves a subscription by user ID and notification type ID
	GetByUserAndType(ctx context.Context, userID uuid.UUID, notificationTypeID int) (*entity.Subscription, error)

	// GetActiveByChatID retrieves all active subscriptions for a chat
	GetActiveByChatID(ctx context.Context, chatID int64) ([]*entity.Subscription, error)

	// GetActiveByType retrieves all active subscriptions for a notification type
	GetActiveByType(ctx context.Context, notificationTypeID int) ([]*entity.Subscription, error)

	// GetDueForNotification retrieves subscriptions that are due for notification
	GetDueForNotification(ctx context.Context, notificationTypeID int) ([]*entity.Subscription, error)

	// Update updates an existing subscription
	Update(ctx context.Context, subscription *entity.Subscription) error

	// UpdateLastNotified updates the last notified timestamp
	UpdateLastNotified(ctx context.Context, id int64) error

	// Delete deletes a subscription by ID
	Delete(ctx context.Context, id int64) error

	// DeleteByUserAndType deletes a subscription by user ID and notification type ID
	DeleteByUserAndType(ctx context.Context, userID uuid.UUID, notificationTypeID int) error
}

// NotificationLogRepository defines the interface for notification log data access
type NotificationLogRepository interface {
	// Create creates a new notification log
	Create(ctx context.Context, log *entity.NotificationLog) error

	// GetByID retrieves a notification log by ID
	GetByID(ctx context.Context, id int64) (*entity.NotificationLog, error)

	// GetBySubscriptionID retrieves all logs for a subscription
	GetBySubscriptionID(ctx context.Context, subscriptionID int64, offset, limit int) ([]*entity.NotificationLog, error)

	// GetRecentLogs retrieves recent notification logs
	GetRecentLogs(ctx context.Context, limit int) ([]*entity.NotificationLog, error)

	// Update updates an existing notification log
	Update(ctx context.Context, log *entity.NotificationLog) error

	// Delete deletes a notification log by ID
	Delete(ctx context.Context, id int64) error

	// CleanupOldLogs deletes logs older than the specified number of days
	CleanupOldLogs(ctx context.Context, daysOld int) error
}
