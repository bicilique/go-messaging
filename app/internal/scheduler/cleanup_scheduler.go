package scheduler

import (
	"context"
	"go-messaging/service"
	"log/slog"
	"time"
)

type CleanupScheduler struct {
	adminService service.AdminServiceInterface
	ticker       *time.Ticker
	done         chan bool
}

func NewCleanupScheduler(adminService service.AdminServiceInterface) *CleanupScheduler {
	return &CleanupScheduler{
		adminService: adminService,
		done:         make(chan bool),
	}
}

// Start begins the cleanup scheduler - runs every hour
func (s *CleanupScheduler) Start() {
	s.ticker = time.NewTicker(1 * time.Hour)

	go func() {
		slog.Info("Starting cleanup scheduler")

		// Run cleanup immediately on start
		s.runCleanup()

		// Then run every hour
		for {
			select {
			case <-s.done:
				slog.Info("Cleanup scheduler stopped")
				return
			case <-s.ticker.C:
				s.runCleanup()
			}
		}
	}()
}

// Stop stops the cleanup scheduler
func (s *CleanupScheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.done <- true
}

// runCleanup performs the actual cleanup
func (s *CleanupScheduler) runCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	slog.Info("Running scheduled cleanup of pending users")

	count, err := s.adminService.CleanupPendingUsers(ctx)
	if err != nil {
		slog.Error("Failed to run scheduled cleanup", "error", err)
		return
	}

	if count > 0 {
		slog.Info("Scheduled cleanup completed", "deleted_count", count)
	} else {
		slog.Debug("Scheduled cleanup completed, no users to delete")
	}
}
