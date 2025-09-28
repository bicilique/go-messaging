package scheduler

import (
	"context"
	"log"
	"time"

	"go-messaging/service"
)

type NotificationScheduler struct {
	dispatchService service.NotificationDispatchService
	schedule        map[string]int // notification type -> interval in minutes
}

func NewNotificationScheduler(dispatchService service.NotificationDispatchService) *NotificationScheduler {
	return &NotificationScheduler{
		dispatchService: dispatchService,
		schedule: map[string]int{
			"coinbase":    1, // Every 1 minute
			"news":        2, // Every 2 minutes
			"weather":     4, // Every 4 minutes
			"price_alert": 5, // Every 5 minutes for testing (change back to 5+ for production)
			"custom":      6, // Every 6 minutes
		},
	}
}

// SetSchedule allows customizing the notification schedule
func (ns *NotificationScheduler) SetSchedule(schedule map[string]int) {
	ns.schedule = schedule
}

// Start begins the notification scheduling process
func (ns *NotificationScheduler) Start(ctx context.Context) {
	log.Println("📡 Starting notification scheduler...")

	// Start individual schedulers for each notification type
	for notificationType, intervalMinutes := range ns.schedule {
		go ns.runNotificationSchedule(ctx, notificationType, intervalMinutes)
	}

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("📡 Notification scheduler stopped")
}

// runNotificationSchedule runs a scheduler for a specific notification type
func (ns *NotificationScheduler) runNotificationSchedule(ctx context.Context, notificationType string, intervalMinutes int) {
	log.Printf("⏰ Starting %s notification scheduler (every %d minutes)", notificationType, intervalMinutes)

	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	// For development: Don't run immediately on startup, wait for first interval
	log.Printf("⏳ Waiting %d minutes before first %s notification...", intervalMinutes, notificationType)

	for {
		select {
		case <-ctx.Done():
			log.Printf("⏰ %s notification scheduler stopped", notificationType)
			return
		case <-ticker.C:
			log.Printf("🔔 Time to dispatch %s notifications!", notificationType)
			if err := ns.dispatchService.DispatchNotification(ctx, notificationType); err != nil {
				log.Printf("❌ Failed to dispatch %s notifications: %v", notificationType, err)
			} else {
				log.Printf("✅ Dispatched %s notifications", notificationType)
			}
		}
	}
}
