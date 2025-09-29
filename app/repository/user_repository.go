package repository

import (
	"context"
	"time"

	"go-messaging/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormUserRepository implements UserRepository using GORM
type GormUserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) GetByTelegramUserID(ctx context.Context, telegramUserID int64) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("telegram_user_id = ?", telegramUserID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *GormUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

func (r *GormUserRepository) List(ctx context.Context, offset, limit int) ([]*entity.User, error) {
	var users []*entity.User
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

// Admin-specific methods
func (r *GormUserRepository) GetUsersByApprovalStatus(ctx context.Context, status string) ([]entity.User, error) {
	var users []entity.User
	err := r.db.WithContext(ctx).Where("approval_status = ?", status).Order("created_at DESC").Find(&users).Error
	return users, err
}

func (r *GormUserRepository) GetUsersByApprovalStatusWithLimit(ctx context.Context, status string, limit int) ([]entity.User, error) {
	var users []entity.User
	err := r.db.WithContext(ctx).Where("approval_status = ?", status).Order("created_at DESC").Limit(limit).Find(&users).Error
	return users, err
}

func (r *GormUserRepository) CountUsersByApprovalStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("approval_status = ?", status).Count(&count).Error
	return count, err
}

func (r *GormUserRepository) CountUsersByRole(ctx context.Context, role string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("role = ?", role).Count(&count).Error
	return count, err
}

func (r *GormUserRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&count).Error
	return count, err
}

func (r *GormUserRepository) DeletePendingUsersOlderThan(ctx context.Context, duration time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-duration)
	result := r.db.WithContext(ctx).Where("approval_status = ? AND created_at < ?", "pending", cutoffTime).Delete(&entity.User{})
	return int(result.RowsAffected), result.Error
}
