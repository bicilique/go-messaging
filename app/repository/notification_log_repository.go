package repository

import (
	"context"
	"time"

	"go-messaging/entity"

	"gorm.io/gorm"
)

// GormNotificationLogRepository implements NotificationLogRepository using GORM
type GormNotificationLogRepository struct {
	db *gorm.DB
}

// NewNotificationLogRepository creates a new notification log repository
func NewNotificationLogRepository(db *gorm.DB) NotificationLogRepository {
	return &GormNotificationLogRepository{db: db}
}

func (r *GormNotificationLogRepository) Create(ctx context.Context, log *entity.NotificationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *GormNotificationLogRepository) GetByID(ctx context.Context, id int64) (*entity.NotificationLog, error) {
	var log entity.NotificationLog
	err := r.db.WithContext(ctx).
		Preload("Subscription").
		Preload("Subscription.User").
		Preload("Subscription.NotificationType").
		First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *GormNotificationLogRepository) GetBySubscriptionID(ctx context.Context, subscriptionID int64, offset, limit int) ([]*entity.NotificationLog, error) {
	var logs []*entity.NotificationLog
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("sent_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *GormNotificationLogRepository) GetRecentLogs(ctx context.Context, limit int) ([]*entity.NotificationLog, error) {
	var logs []*entity.NotificationLog
	err := r.db.WithContext(ctx).
		Preload("Subscription").
		Preload("Subscription.User").
		Preload("Subscription.NotificationType").
		Order("sent_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *GormNotificationLogRepository) Update(ctx context.Context, log *entity.NotificationLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

func (r *GormNotificationLogRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.NotificationLog{}, id).Error
}

func (r *GormNotificationLogRepository) CleanupOldLogs(ctx context.Context, daysOld int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)
	return r.db.WithContext(ctx).
		Where("sent_at < ?", cutoffDate).
		Delete(&entity.NotificationLog{}).Error
}
