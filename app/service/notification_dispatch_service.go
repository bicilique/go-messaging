package service

import (
	"context"
	"fmt"

	"go-messaging/entity"
	"go-messaging/model"
)

// NotificationDispatchServiceImpl implements NotificationDispatchService
type NotificationDispatchServiceImpl struct {
	subscriptionService SubscriptionService
	logService          NotificationLogService
	telegramService     TelegramNotificationSender
}

// TelegramNotificationSender defines interface for sending Telegram messages
type TelegramNotificationSender interface {
	SendMessage(chatID int64, message string) error
	SendMessageWithKeyboard(chatID int64, message string, keyboard model.InlineKeyboardMarkup) error
	AnswerCallbackQuery(callbackID, text string) error
}

// NewNotificationDispatchService creates a new notification dispatch service
func NewNotificationDispatchService(
	subscriptionService SubscriptionService,
	logService NotificationLogService,
	telegramService TelegramNotificationSender,
) NotificationDispatchService {
	return &NotificationDispatchServiceImpl{
		subscriptionService: subscriptionService,
		logService:          logService,
		telegramService:     telegramService,
	}
}

func (s *NotificationDispatchServiceImpl) DispatchNotification(ctx context.Context, notificationTypeCode string) error {
	// Get subscriptions that are due for notification
	subscriptions, err := s.subscriptionService.GetDueSubscriptions(ctx, notificationTypeCode)
	if err != nil {
		return fmt.Errorf("failed to get due subscriptions: %w", err)
	}

	fmt.Printf("📋 Found %d due subscriptions for %s\n", len(subscriptions), notificationTypeCode)

	if len(subscriptions) == 0 {
		return nil // No subscriptions to notify
	}

	// Send notifications to all due subscriptions
	successCount := 0
	for _, subscription := range subscriptions {
		fmt.Printf("📤 Processing subscription %d for user %d\n", subscription.ID, subscription.UserID)
		if err := s.processSubscriptionNotification(ctx, subscription, notificationTypeCode); err != nil {
			fmt.Printf("Failed to process notification for subscription %d: %v\n", subscription.ID, err)
		} else {
			successCount++
			fmt.Printf("✅ Successfully processed subscription %d\n", subscription.ID)
		}
	}

	fmt.Printf("📊 Processed %d/%d subscriptions successfully for %s\n", successCount, len(subscriptions), notificationTypeCode)
	return nil
}

// DispatchToSubscription sends a notification to a specific subscription
// Message should be pre-generated and validated
func (s *NotificationDispatchServiceImpl) DispatchToSubscription(ctx context.Context, subscription *entity.Subscription, message string) error {
	return s.sendNotificationToSubscription(ctx, subscription, message)
}

// GetNotificationContent generates notification content based on type and preferences
func (s *NotificationDispatchServiceImpl) GetNotificationContent(ctx context.Context, notificationTypeCode string, preferences *entity.SubscriptionPreferences) (string, error) {
	switch notificationTypeCode {
	case "security":
		return s.getSecurityContent(ctx, preferences)
	default:
		return "", fmt.Errorf("unknown notification type: %s", notificationTypeCode)
	}
}

func (s *NotificationDispatchServiceImpl) processSubscriptionNotification(ctx context.Context, subscription *entity.Subscription, notificationTypeCode string) error {
	fmt.Printf("🔄 Generating content for %s notification (subscription %d)\n", notificationTypeCode, subscription.ID)

	// Generate notification content
	content, err := s.GetNotificationContent(ctx, notificationTypeCode, &subscription.Preferences)
	if err != nil {
		return fmt.Errorf("failed to get notification content: %w", err)
	}

	fmt.Printf("📝 Generated content for subscription %d: %.100s...\n", subscription.ID, content)

	// Send the notification
	if err := s.sendNotificationToSubscription(ctx, subscription, content); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	fmt.Printf("📨 Sent notification for subscription %d\n", subscription.ID)

	// Mark subscription as notified
	if err := s.subscriptionService.MarkNotified(ctx, subscription.ID); err != nil {
		return fmt.Errorf("failed to mark subscription as notified: %w", err)
	}

	fmt.Printf("✅ Marked subscription %d as notified\n", subscription.ID)
	return nil
}

func (s *NotificationDispatchServiceImpl) sendNotificationToSubscription(ctx context.Context, subscription *entity.Subscription, message string) error {
	// Validate message length
	if err := model.ValidateMessageString(message); err != nil {
		errorMsg := err.Error()
		_, logErr := s.logService.LogNotification(ctx, subscription.ID, message, "failed", &errorMsg)
		if logErr != nil {
			fmt.Printf("Failed to log notification error: %v\n", logErr)
		}
		return err
	}

	// Send via Telegram
	if err := s.telegramService.SendMessage(subscription.ChatID, message); err != nil {
		errorMsg := err.Error()
		_, logErr := s.logService.LogNotification(ctx, subscription.ID, message, "failed", &errorMsg)
		if logErr != nil {
			fmt.Printf("Failed to log notification error: %v\n", logErr)
		}
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	// Log successful notification
	_, err := s.logService.LogNotification(ctx, subscription.ID, message, "sent", nil)
	if err != nil {
		fmt.Printf("Failed to log notification success: %v\n", err)
		// Don't return error as the notification was sent successfully
	}

	return nil
}

// Content generation methods for different notification types
func (s *NotificationDispatchServiceImpl) getSecurityContent(ctx context.Context, preferences *entity.SubscriptionPreferences) (string, error) {
	return "🔒 Security Alert\n\nThis is a security-related notification.\n\nStay safe!", nil
}
