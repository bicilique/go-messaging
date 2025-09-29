package service

import (
	"context"
	"fmt"
	"time"

	"go-messaging/entity"
	"go-messaging/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserServiceImpl implements UserService
type UserServiceImpl struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository) UserService {
	return &UserServiceImpl{
		userRepo: userRepo,
	}
}

func (s *UserServiceImpl) CreateOrUpdateUser(ctx context.Context, telegramUserID int64, username, firstName, lastName, languageCode *string, isBot bool) (*entity.User, error) {
	// Try to get existing user
	existingUser, err := s.userRepo.GetByTelegramUserID(ctx, telegramUserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		// Update existing user
		existingUser.Username = username
		existingUser.FirstName = firstName
		existingUser.LastName = lastName
		existingUser.LanguageCode = languageCode
		existingUser.IsBot = isBot
		existingUser.UpdatedAt = time.Now()

		if err := s.userRepo.Update(ctx, existingUser); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		return existingUser, nil
	}

	// Create new user
	user := &entity.User{
		TelegramUserID: telegramUserID,
		Username:       username,
		FirstName:      firstName,
		LastName:       lastName,
		LanguageCode:   languageCode,
		IsBot:          isBot,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserServiceImpl) GetUserByTelegramID(ctx context.Context, telegramUserID int64) (*entity.User, error) {
	user, err := s.userRepo.GetByTelegramUserID(ctx, telegramUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *UserServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *UserServiceImpl) UpdateUser(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, telegramUserID int64) error {
	user, err := s.userRepo.GetByTelegramUserID(ctx, telegramUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.userRepo.Delete(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *UserServiceImpl) ListUsers(ctx context.Context, offset, limit int) ([]*entity.User, error) {
	users, err := s.userRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *UserServiceImpl) CountUsers(ctx context.Context) (int64, error) {
	count, err := s.userRepo.CountAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
