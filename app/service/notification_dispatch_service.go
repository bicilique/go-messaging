package service

import (
	"context"
	"fmt"
	"strings"
	"time"

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
			// Log error but continue with other subscriptions
			fmt.Printf("Failed to process notification for subscription %d: %v\n", subscription.ID, err)
		} else {
			successCount++
			fmt.Printf("✅ Successfully processed subscription %d\n", subscription.ID)
		}
	}

	fmt.Printf("📊 Processed %d/%d subscriptions successfully for %s\n", successCount, len(subscriptions), notificationTypeCode)
	return nil
}

func (s *NotificationDispatchServiceImpl) DispatchToSubscription(ctx context.Context, subscription *entity.Subscription, message string) error {
	return s.sendNotificationToSubscription(ctx, subscription, message)
}

func (s *NotificationDispatchServiceImpl) GetNotificationContent(ctx context.Context, notificationTypeCode string, preferences *entity.SubscriptionPreferences) (string, error) {
	switch notificationTypeCode {
	case "coinbase":
		return s.getCoinbaseContent(ctx, preferences)
	case "news":
		return s.getNewsContent(ctx, preferences)
	case "weather":
		return s.getWeatherContent(ctx, preferences)
	case "price_alert":
		return s.getPriceAlertContent(ctx, preferences)
	case "custom":
		return s.getCustomContent(ctx, preferences)
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

func (s *NotificationDispatchServiceImpl) getCoinbaseContent(ctx context.Context, preferences *entity.SubscriptionPreferences) (string, error) {
	currency := "BTC"
	if preferences != nil && preferences.Currency != "" {
		currency = strings.ToUpper(preferences.Currency)
	}

	// Mock API call - replace with actual Coinbase API integration
	price, err := s.fetchCoinbasePrice(currency)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s price: %w", currency, err)
	}

	return fmt.Sprintf("🪙 %s Price Update\n\nCurrent price: $%.2f\n\nUpdated: %s",
		currency, price, time.Now().Format("15:04 MST")), nil
}

func (s *NotificationDispatchServiceImpl) getNewsContent(ctx context.Context, preferences *entity.SubscriptionPreferences) (string, error) {
	keywords := []string{"technology", "crypto"}
	if preferences != nil && len(preferences.Keywords) > 0 {
		keywords = preferences.Keywords
	}

	// Mock news content - replace with actual news API integration
	news := s.fetchNews(keywords)

	var content strings.Builder
	content.WriteString("📰 Latest News\n\n")

	for i, article := range news {
		if i >= 3 { // Limit to 3 articles
			break
		}
		content.WriteString(fmt.Sprintf("• %s\n", article))
	}

	content.WriteString(fmt.Sprintf("\nUpdated: %s", time.Now().Format("15:04 MST")))

	return content.String(), nil
}

func (s *NotificationDispatchServiceImpl) getWeatherContent(ctx context.Context, preferences *entity.SubscriptionPreferences) (string, error) {
	location := "San Francisco, CA"
	if preferences != nil && preferences.Settings != nil {
		if loc, ok := preferences.Settings["location"]; ok {
			location = loc
		}
	}

	// Mock weather data - replace with actual weather API integration
	weather := s.fetchWeather(location)

	return fmt.Sprintf("🌤 Weather Update for %s\n\n%s\n\nUpdated: %s",
		location, weather, time.Now().Format("15:04 MST")), nil
}

func (s *NotificationDispatchServiceImpl) getPriceAlertContent(ctx context.Context, preferences *entity.SubscriptionPreferences) (string, error) {
	// Provide default values if preferences are missing or incomplete
	currency := "BTC"
	threshold := 50000.0

	if preferences != nil {
		if preferences.Currency != "" {
			currency = strings.ToUpper(preferences.Currency)
		}
		if preferences.Threshold > 0 {
			threshold = preferences.Threshold
		}
	}

	// Mock price check - replace with actual API integration
	currentPrice, err := s.fetchCoinbasePrice(currency)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s price: %w", currency, err)
	}

	// For prototype/dev: Always send notification regardless of threshold
	// In production, you'd uncomment the condition below:
	/*
		if currentPrice >= threshold {
			return fmt.Sprintf("🚨 Price Alert: %s\n\nCurrent price: $%.2f\nThreshold: $%.2f\n\nAlert triggered at %s",
				currency, currentPrice, threshold, time.Now().Format("15:04 MST")), nil
		}
		return "", fmt.Errorf("price threshold not met")
	*/

	// Development version: Always send notification with current price info
	status := "📊" // Default status
	if currentPrice >= threshold {
		status = "🚨" // Alert status if threshold would be met
	}

	return fmt.Sprintf("%s Price Alert: %s\n\nCurrent price: $%.2f\nThreshold: $%.2f\nStatus: %s\n\nUpdate time: %s",
		status, currency, currentPrice, threshold,
		func() string {
			if currentPrice >= threshold {
				return "THRESHOLD MET"
			}
			return "Monitoring"
		}(),
		time.Now().Format("15:04 MST")), nil
}

func (s *NotificationDispatchServiceImpl) getCustomContent(ctx context.Context, preferences *entity.SubscriptionPreferences) (string, error) {
	customMessage := "Custom notification"
	if preferences != nil && preferences.Settings != nil {
		if msg, ok := preferences.Settings["message"]; ok {
			customMessage = msg
		}
	}

	return fmt.Sprintf("🔔 Custom Notification\n\n%s\n\nSent: %s",
		customMessage, time.Now().Format("15:04 MST")), nil
}

// Mock external API calls - replace with actual implementations

func (s *NotificationDispatchServiceImpl) fetchCoinbasePrice(currency string) (float64, error) {
	// Mock implementation - replace with actual Coinbase API call
	prices := map[string]float64{
		"BTC": 45000.50,
		"ETH": 3200.75,
		"ADA": 1.25,
		"DOT": 35.80,
	}

	if price, ok := prices[currency]; ok {
		// Add some randomness to simulate price changes
		return price + (float64(time.Now().Unix()%100) - 50), nil
	}

	return 0, fmt.Errorf("currency %s not supported", currency)
}

func (s *NotificationDispatchServiceImpl) fetchNews(keywords []string) []string {
	// Mock implementation - replace with actual news API call
	articles := []string{
		"Bitcoin reaches new all-time high amid institutional adoption",
		"Major tech companies announce blockchain partnerships",
		"Cryptocurrency regulation updates from global markets",
		"New DeFi protocol launches with innovative features",
		"Market analysis: Crypto winter may be ending",
	}

	// Filter by keywords (simplified)
	var filtered []string
	for _, article := range articles {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(article), strings.ToLower(keyword)) {
				filtered = append(filtered, article)
				break
			}
		}
	}

	if len(filtered) == 0 {
		return articles[:3] // Return first 3 if no matches
	}

	return filtered
}

func (s *NotificationDispatchServiceImpl) fetchWeather(location string) string {
	// Mock implementation - replace with actual weather API call
	weathers := []string{
		"Sunny, 72°F (22°C)\nWind: 5 mph\nHumidity: 45%",
		"Partly cloudy, 68°F (20°C)\nWind: 8 mph\nHumidity: 55%",
		"Light rain, 65°F (18°C)\nWind: 12 mph\nHumidity: 78%",
	}

	// Return based on location hash (simplified)
	index := len(location) % len(weathers)
	return weathers[index]
}
