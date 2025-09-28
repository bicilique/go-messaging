package repository

import (
	"context"
	"time"

	"go-messaging/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormSubscriptionRepository implements SubscriptionRepository using GORM
type GormSubscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &GormSubscriptionRepository{db: db}
}

func (r *GormSubscriptionRepository) Create(ctx context.Context, subscription *entity.Subscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

func (r *GormSubscriptionRepository) GetByID(ctx context.Context, id int64) (*entity.Subscription, error) {
	var subscription entity.Subscription
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("NotificationType").
		First(&subscription, id).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *GormSubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Subscription, error) {
	var subscriptions []*entity.Subscription
	err := r.db.WithContext(ctx).
		Preload("NotificationType").
		Where("user_id = ?", userID).
		Find(&subscriptions).Error
	return subscriptions, err
}

func (r *GormSubscriptionRepository) GetByUserAndType(ctx context.Context, userID uuid.UUID, notificationTypeID int) (*entity.Subscription, error) {
	var subscription entity.Subscription
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("NotificationType").
		Where("user_id = ? AND notification_type_id = ?", userID, notificationTypeID).
		First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *GormSubscriptionRepository) GetActiveByChatID(ctx context.Context, chatID int64) ([]*entity.Subscription, error) {
	var subscriptions []*entity.Subscription
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("NotificationType").
		Where("chat_id = ? AND is_active = ?", chatID, true).
		Find(&subscriptions).Error
	return subscriptions, err
}

func (r *GormSubscriptionRepository) GetActiveByType(ctx context.Context, notificationTypeID int) ([]*entity.Subscription, error) {
	var subscriptions []*entity.Subscription
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("NotificationType").
		Where("notification_type_id = ? AND is_active = ?", notificationTypeID, true).
		Find(&subscriptions).Error
	return subscriptions, err
}

func (r *GormSubscriptionRepository) GetDueForNotification(ctx context.Context, notificationTypeID int) ([]*entity.Subscription, error) {
	var subscriptions []*entity.Subscription

	// Subquery to get the interval from preferences or default from notification type
	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("NotificationType").
		Where("notification_type_id = ? AND is_active = ?", notificationTypeID, true)

	// Get subscriptions that haven't been notified yet or are due based on interval
	query = query.Where(`
		last_notified_at IS NULL OR 
		last_notified_at <= NOW() - INTERVAL '1 minute' * COALESCE(
			CAST(preferences->>'interval' AS INTEGER), 
			(SELECT default_interval_minutes FROM notification_types WHERE id = subscriptions.notification_type_id)
		)
	`)

	err := query.Find(&subscriptions).Error
	return subscriptions, err
}

func (r *GormSubscriptionRepository) Update(ctx context.Context, subscription *entity.Subscription) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

func (r *GormSubscriptionRepository) UpdateLastNotified(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.Subscription{}).
		Where("id = ?", id).
		Update("last_notified_at", now).Error
}

func (r *GormSubscriptionRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Subscription{}, id).Error
}

func (r *GormSubscriptionRepository) DeleteByUserAndType(ctx context.Context, userID uuid.UUID, notificationTypeID int) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND notification_type_id = ?", userID, notificationTypeID).
		Delete(&entity.Subscription{}).Error
}
