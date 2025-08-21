package service

import (
	"context"
	"go-messaging/entity"
)

// SubscriptionService defines the interface for subscription business logic
type SubscriptionService interface {
	// Subscribe creates or updates a subscription for a user
	Subscribe(ctx context.Context, telegramUserID int64, chatID int64, notificationTypeCode string, preferences *entity.SubscriptionPreferences) (*entity.Subscription, error)

	// Unsubscribe removes a subscription for a user
	Unsubscribe(ctx context.Context, telegramUserID int64, notificationTypeCode string) error

	// GetUserSubscriptions retrieves all subscriptions for a user
	GetUserSubscriptions(ctx context.Context, telegramUserID int64) ([]*entity.Subscription, error)

	// GetActiveSubscriptions retrieves all active subscriptions for a notification type
	GetActiveSubscriptions(ctx context.Context, notificationTypeCode string) ([]*entity.Subscription, error)

	// GetDueSubscriptions retrieves subscriptions that are due for notification
	GetDueSubscriptions(ctx context.Context, notificationTypeCode string) ([]*entity.Subscription, error)

	// UpdatePreferences updates subscription preferences
	UpdatePreferences(ctx context.Context, telegramUserID int64, notificationTypeCode string, preferences *entity.SubscriptionPreferences) error

	// MarkNotified updates the last notified timestamp for a subscription
	MarkNotified(ctx context.Context, subscriptionID int64) error
}

// UserService defines the interface for user business logic
type UserService interface {
	// CreateOrUpdateUser creates a new user or updates existing user info
	CreateOrUpdateUser(ctx context.Context, telegramUserID int64, username, firstName, lastName, languageCode *string, isBot bool) (*entity.User, error)

	// GetUserByTelegramID retrieves a user by Telegram user ID
	GetUserByTelegramID(ctx context.Context, telegramUserID int64) (*entity.User, error)

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)

	// UpdateUser updates user information
	UpdateUser(ctx context.Context, user *entity.User) error

	// DeleteUser deletes a user and all related data
	DeleteUser(ctx context.Context, telegramUserID int64) error

	// ListUsers retrieves users with pagination
	ListUsers(ctx context.Context, offset, limit int) ([]*entity.User, error)
}

// NotificationTypeService defines the interface for notification type business logic
type NotificationTypeService interface {
	// GetAllTypes retrieves all notification types
	GetAllTypes(ctx context.Context) ([]*entity.NotificationType, error)

	// GetActiveTypes retrieves all active notification types
	GetActiveTypes(ctx context.Context) ([]*entity.NotificationType, error)

	// GetTypeByCode retrieves a notification type by code
	GetTypeByCode(ctx context.Context, code string) (*entity.NotificationType, error)

	// GetTypeByID retrieves a notification type by ID
	GetTypeByID(ctx context.Context, id int) (*entity.NotificationType, error)

	// CreateType creates a new notification type
	CreateType(ctx context.Context, code, name string, description *string, defaultInterval int) (*entity.NotificationType, error)

	// UpdateType updates a notification type
	UpdateType(ctx context.Context, notificationType *entity.NotificationType) error

	// ActivateType activates a notification type
	ActivateType(ctx context.Context, code string) error

	// DeactivateType deactivates a notification type
	DeactivateType(ctx context.Context, code string) error
}

// NotificationLogService defines the interface for notification log business logic
type NotificationLogService interface {
	// LogNotification creates a notification log entry
	LogNotification(ctx context.Context, subscriptionID int64, message, status string, errorMessage *string) (*entity.NotificationLog, error)

	// GetSubscriptionLogs retrieves logs for a subscription with pagination
	GetSubscriptionLogs(ctx context.Context, subscriptionID int64, offset, limit int) ([]*entity.NotificationLog, error)

	// GetRecentLogs retrieves recent notification logs
	GetRecentLogs(ctx context.Context, limit int) ([]*entity.NotificationLog, error)

	// UpdateLogStatus updates the status of a notification log
	UpdateLogStatus(ctx context.Context, logID int64, status string, errorMessage *string) error

	// CleanupOldLogs removes logs older than specified days
	CleanupOldLogs(ctx context.Context, daysOld int) error
}

// NotificationDispatchService defines the interface for sending notifications
type NotificationDispatchService interface {
	// DispatchNotification sends a notification for a specific type
	DispatchNotification(ctx context.Context, notificationTypeCode string) error

	// DispatchToSubscription sends a notification to a specific subscription
	DispatchToSubscription(ctx context.Context, subscription *entity.Subscription, message string) error

	// GetNotificationContent generates content for a notification type
	GetNotificationContent(ctx context.Context, notificationTypeCode string, preferences *entity.SubscriptionPreferences) (string, error)
}
