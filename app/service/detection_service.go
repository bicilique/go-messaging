package service

import (
	"context"
	"fmt"
	"go-messaging/model"
	"go-messaging/repository"
)

type DetectionService struct {
	notificationDispatchService NotificationDispatchService
	subscriptionRepo            repository.SubscriptionRepository
	notificationRepo            repository.NotificationTypeRepository
}

// NewDetectionService creates a new instance of DetectionService
func NewDetectionService(notificationDispatchService NotificationDispatchService, subscriptionRepo repository.SubscriptionRepository, notificationRepo repository.NotificationTypeRepository) DetectionInterface {
	return &DetectionService{
		notificationDispatchService: notificationDispatchService,
		subscriptionRepo:            subscriptionRepo,
		notificationRepo:            notificationRepo,
	}

}

// Send detection notification to all relevant subscribers
func (s *DetectionService) SendDetectionNotification(ctx context.Context, request model.DetectionSummary) error {
	notificationType, err := s.notificationRepo.GetByCode(ctx, "security")
	if err != nil {
		return fmt.Errorf("failed to get notification type: %w", err)
	}
	if notificationType == nil || !notificationType.IsActive {
		return fmt.Errorf("notification type 'security' is not active or not found")
	}

	subscribers, err := s.subscriptionRepo.GetActiveByType(ctx, notificationType.ID)
	if err != nil {
		return fmt.Errorf("failed to get active subscriptions: %w", err)
	}
	if len(subscribers) == 0 {
		fmt.Println("No active subscribers for 'security' notifications")
		return nil // No subscribers to notify
	}

	message := s.generateTelegramMessage(request)

	var failed []int64
	var success []int64

	for _, sub := range subscribers {
		if err := s.notificationDispatchService.DispatchToSubscription(ctx, sub, message); err != nil {
			fmt.Printf("❌ Failed to send notification to subscription %d: %v\n", sub.ID, err)
			failed = append(failed, sub.ID)
		} else {
			fmt.Printf("✅ Notification sent to subscription %d\n", sub.ID)
			success = append(success, sub.ID)
		}
	}

	fmt.Printf("Notification dispatch summary: %d succeeded, %d failed\n", len(success), len(failed))
	if len(failed) > 0 {
		return fmt.Errorf("failed to send notification to %d subscriptions: %v", len(failed), failed)
	}

	return nil
}
func (s *DetectionService) generateTelegramMessage(request model.DetectionSummary) string {
	message := "🔍 *Detection Summary*\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━\n"
	message += "📄 *Filename*: `" + request.Filename + "`\n"
	message += "🏷️ *Classification*: *" + request.Classification + "*\n"
	message += "⚠️ *Risk Level*: *" + request.RiskLevel + "*\n"
	message += "📊 *Confidence*: " + request.Confidence + "\n"
	message += "📝 *Action Required*: " + request.ActionRequired + "\n\n"
	message += "🧾 *Summary:*\n" + request.Summary + "\n\n"
	if len(request.KeyFindings) > 0 {
		message += "🔑 *Key Findings:*\n"
		for _, finding := range request.KeyFindings {
			message += "   • " + finding + "\n"
		}
		message += "\n"
	}
	message += "⏱️ *Processing Time*: " + request.ProcessingTime + "\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━"

	return message
}
