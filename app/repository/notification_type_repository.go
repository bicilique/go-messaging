package repository

import (
	"context"

	"go-messaging/entity"

	"gorm.io/gorm"
)

// GormNotificationTypeRepository implements NotificationTypeRepository using GORM
type GormNotificationTypeRepository struct {
	db *gorm.DB
}

// NewNotificationTypeRepository creates a new notification type repository
func NewNotificationTypeRepository(db *gorm.DB) NotificationTypeRepository {
	return &GormNotificationTypeRepository{db: db}
}

func (r *GormNotificationTypeRepository) GetAll(ctx context.Context) ([]*entity.NotificationType, error) {
	var types []*entity.NotificationType
	err := r.db.WithContext(ctx).Find(&types).Error
	return types, err
}

func (r *GormNotificationTypeRepository) GetByID(ctx context.Context, id int) (*entity.NotificationType, error) {
	var notificationType entity.NotificationType
	err := r.db.WithContext(ctx).First(&notificationType, id).Error
	if err != nil {
		return nil, err
	}
	return &notificationType, nil
}

func (r *GormNotificationTypeRepository) GetByCode(ctx context.Context, code string) (*entity.NotificationType, error) {
	var notificationType entity.NotificationType
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&notificationType).Error
	if err != nil {
		return nil, err
	}
	return &notificationType, nil
}

func (r *GormNotificationTypeRepository) GetActive(ctx context.Context) ([]*entity.NotificationType, error) {
	var types []*entity.NotificationType
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&types).Error
	return types, err
}

func (r *GormNotificationTypeRepository) Create(ctx context.Context, notificationType *entity.NotificationType) error {
	return r.db.WithContext(ctx).Create(notificationType).Error
}

func (r *GormNotificationTypeRepository) Update(ctx context.Context, notificationType *entity.NotificationType) error {
	return r.db.WithContext(ctx).Save(notificationType).Error
}

func (r *GormNotificationTypeRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&entity.NotificationType{}, id).Error
}
