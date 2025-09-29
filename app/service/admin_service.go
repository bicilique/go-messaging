package service

import (
	"context"
	"fmt"
	"go-messaging/entity"
	"go-messaging/repository"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type AdminService struct {
	userRepo repository.UserRepository
}

type AdminServiceInterface interface {
	GetPendingUsers(ctx context.Context) ([]entity.User, error)
	GetApprovedUsers(ctx context.Context, limit int) ([]entity.User, error)
	ApproveUser(ctx context.Context, userID uuid.UUID, adminID uuid.UUID) error
	RejectUser(ctx context.Context, userID uuid.UUID, adminID uuid.UUID) error
	DisableUser(ctx context.Context, userID uuid.UUID, adminID uuid.UUID) error
	EnableUser(ctx context.Context, userID uuid.UUID, adminID uuid.UUID) error
	CreateAdmin(ctx context.Context, telegramUserID int64, username, firstName, lastName string) error
	IsAdmin(ctx context.Context, telegramUserID int64) (bool, error)
	GetUserStats(ctx context.Context) (map[string]int64, error)
	CleanupPendingUsers(ctx context.Context) (int, error)
}

func NewAdminService(userRepo repository.UserRepository) AdminServiceInterface {
	return &AdminService{
		userRepo: userRepo,
	}
}

func (s *AdminService) GetPendingUsers(ctx context.Context) ([]entity.User, error) {
	users, err := s.userRepo.GetUsersByApprovalStatus(ctx, "pending")
	if err != nil {
		slog.Error("Failed to get pending users", "error", err)
		return nil, err
	}
	return users, nil
}

func (s *AdminService) GetApprovedUsers(ctx context.Context, limit int) ([]entity.User, error) {
	users, err := s.userRepo.GetUsersByApprovalStatusWithLimit(ctx, "approved", limit)
	if err != nil {
		slog.Error("Failed to get approved users", "error", err)
		return nil, err
	}
	return users, nil
}

func (s *AdminService) ApproveUser(ctx context.Context, userID uuid.UUID, adminID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user for approval", "userID", userID, "error", err)
		return err
	}

	if user.ApprovalStatus == "approved" {
		return fmt.Errorf("user is already approved")
	}

	now := time.Now()
	user.ApprovalStatus = "approved"
	user.ApprovedBy = &adminID
	user.ApprovedAt = &now

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		slog.Error("Failed to approve user", "userID", userID, "adminID", adminID, "error", err)
		return err
	}

	slog.Info("User approved successfully", "userID", userID, "adminID", adminID)
	return nil
}

func (s *AdminService) RejectUser(ctx context.Context, userID uuid.UUID, adminID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user for rejection", "userID", userID, "error", err)
		return err
	}

	now := time.Now()
	user.ApprovalStatus = "rejected"
	user.ApprovedBy = &adminID
	user.ApprovedAt = &now

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		slog.Error("Failed to reject user", "userID", userID, "adminID", adminID, "error", err)
		return err
	}

	slog.Info("User rejected successfully", "userID", userID, "adminID", adminID)
	return nil
}

func (s *AdminService) DisableUser(ctx context.Context, userID uuid.UUID, adminID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user for disabling", "userID", userID, "error", err)
		return err
	}

	now := time.Now()
	user.ApprovalStatus = "disabled"
	user.ApprovedBy = &adminID
	user.ApprovedAt = &now

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		slog.Error("Failed to disable user", "userID", userID, "adminID", adminID, "error", err)
		return err
	}

	slog.Info("User disabled successfully", "userID", userID, "adminID", adminID)
	return nil
}

func (s *AdminService) EnableUser(ctx context.Context, userID uuid.UUID, adminID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user for enabling", "userID", userID, "error", err)
		return err
	}

	now := time.Now()
	user.ApprovalStatus = "approved"
	user.ApprovedBy = &adminID
	user.ApprovedAt = &now

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		slog.Error("Failed to enable user", "userID", userID, "adminID", adminID, "error", err)
		return err
	}

	slog.Info("User enabled successfully", "userID", userID, "adminID", adminID)
	return nil
}

func (s *AdminService) CreateAdmin(ctx context.Context, telegramUserID int64, username, firstName, lastName string) error {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByTelegramUserID(ctx, telegramUserID)
	if err == nil {
		// User exists, just update their role
		existingUser.Role = "admin"
		existingUser.ApprovalStatus = "approved"
		now := time.Now()
		existingUser.ApprovedAt = &now

		err = s.userRepo.Update(ctx, existingUser)
		if err != nil {
			slog.Error("Failed to update user to admin", "telegramUserID", telegramUserID, "error", err)
			return err
		}

		slog.Info("User updated to admin successfully", "telegramUserID", telegramUserID)
		return nil
	}

	// Create new admin user
	user := &entity.User{
		TelegramUserID: telegramUserID,
		Username:       &username,
		FirstName:      &firstName,
		LastName:       &lastName,
		Role:           "admin",
		ApprovalStatus: "approved",
	}

	now := time.Now()
	user.ApprovedAt = &now

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		slog.Error("Failed to create admin user", "telegramUserID", telegramUserID, "error", err)
		return err
	}

	slog.Info("Admin user created successfully", "telegramUserID", telegramUserID)
	return nil
}

func (s *AdminService) IsAdmin(ctx context.Context, telegramUserID int64) (bool, error) {
	user, err := s.userRepo.GetByTelegramUserID(ctx, telegramUserID)
	if err != nil {
		return false, err
	}

	return user.Role == "admin" && user.ApprovalStatus == "approved", nil
}

func (s *AdminService) GetUserStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count users by approval status
	pendingCount, err := s.userRepo.CountUsersByApprovalStatus(ctx, "pending")
	if err != nil {
		slog.Error("Failed to count pending users", "error", err)
		return nil, err
	}
	stats["pending"] = pendingCount

	approvedCount, err := s.userRepo.CountUsersByApprovalStatus(ctx, "approved")
	if err != nil {
		slog.Error("Failed to count approved users", "error", err)
		return nil, err
	}
	stats["approved"] = approvedCount

	rejectedCount, err := s.userRepo.CountUsersByApprovalStatus(ctx, "rejected")
	if err != nil {
		slog.Error("Failed to count rejected users", "error", err)
		return nil, err
	}
	stats["rejected"] = rejectedCount

	disabledCount, err := s.userRepo.CountUsersByApprovalStatus(ctx, "disabled")
	if err != nil {
		slog.Error("Failed to count disabled users", "error", err)
		return nil, err
	}
	stats["disabled"] = disabledCount

	// Count admins
	adminCount, err := s.userRepo.CountUsersByRole(ctx, "admin")
	if err != nil {
		slog.Error("Failed to count admin users", "error", err)
		return nil, err
	}
	stats["admins"] = adminCount

	return stats, nil
}

func (s *AdminService) CleanupPendingUsers(ctx context.Context) (int, error) {
	count, err := s.userRepo.DeletePendingUsersOlderThan(ctx, 6*time.Hour)
	if err != nil {
		slog.Error("Failed to cleanup pending users", "error", err)
		return 0, err
	}

	if count > 0 {
		slog.Info("Cleaned up pending users", "count", count)
	}

	return count, nil
}
