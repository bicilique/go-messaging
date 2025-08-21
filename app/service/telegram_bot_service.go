package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"go-messaging/entity"
	"go-messaging/model"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// TelegramBotService provides methods to interact with the Telegram Bot API
type TelegramBotService struct {
	botInstance             *bot.Bot
	rateLimiter             *model.RateLimiter
	messageValidator        *model.MessageValidator
	userService             UserService
	subscriptionService     SubscriptionService
	notificationTypeService NotificationTypeService
}

// TelegramBotServiceInterface defines the interface for telegram bot operations
type TelegramBotServiceInterface interface {
	SendMessage(chatID int64, message string) error
	StartPolling(ctx context.Context)
	HandleUpdate(ctx context.Context, b *bot.Bot, update *models.Update)
}

// NewTelegramBotService creates a new telegram bot service with all dependencies
func NewTelegramBotService(
	botToken string,
	userService UserService,
	subscriptionService SubscriptionService,
	notificationTypeService NotificationTypeService,
) *TelegramBotService {
	if botToken == "" {
		panic("TELEGRAM BOT TOKEN environment variable not set.")
	}

	botInstance, err := bot.New(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	return &TelegramBotService{
		botInstance:             botInstance,
		rateLimiter:             model.NewRateLimiter(),
		messageValidator:        model.NewMessageValidator(),
		userService:             userService,
		subscriptionService:     subscriptionService,
		notificationTypeService: notificationTypeService,
	}
}

// SendMessage sends a message to a specific chat
func (ts *TelegramBotService) SendMessage(chatID int64, message string) error {
	// Note: Rate limiting is applied to incoming messages, not outgoing bot responses

	// Validate message
	if err := model.ValidateMessageString(message); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := ts.botInstance.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   message,
		// Remove ParseMode to send as plain text to avoid markdown parsing issues
	})

	return err
}

// StartPolling starts the bot polling loop
func (ts *TelegramBotService) StartPolling(ctx context.Context) {
	log.Println("Starting Telegram bot polling...")

	// Register handler for all text messages and commands
	ts.botInstance.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return update.Message != nil && update.Message.Text != ""
	}, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		ts.HandleUpdate(ctx, b, update)
	})

	ts.botInstance.Start(ctx)
}

// HandleUpdate processes incoming updates from Telegram
func (ts *TelegramBotService) HandleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("Received update: %+v", update)

	if update.Message == nil {
		log.Println("Update message is nil, skipping")
		return
	}

	message := update.Message
	chatID := message.Chat.ID
	userID := message.From.ID
	text := message.Text

	log.Printf("Processing message from user %d in chat %d: %s", userID, chatID, text)

	// Apply rate limiting per user
	allowed, reason := ts.rateLimiter.IsAllowed(userID)
	if !allowed {
		ts.SendMessage(chatID, fmt.Sprintf("⏰ Please slow down! %s", reason))
		return
	}

	// Create or update user
	var lastName *string
	if message.From.LastName != "" {
		lastName = &message.From.LastName
	}

	var languageCode *string
	if message.From.LanguageCode != "" {
		languageCode = &message.From.LanguageCode
	}

	user, err := ts.userService.CreateOrUpdateUser(
		ctx,
		userID,
		&message.From.Username,
		&message.From.FirstName,
		lastName,
		languageCode,
		message.From.IsBot,
	)
	if err != nil {
		log.Printf("Failed to create/update user: %v", err)
		ts.SendMessage(chatID, "❌ Sorry, there was an error processing your request.")
		return
	}

	log.Printf("User %s (%d) sent: %s", getDisplayName(user), userID, text)

	// Handle commands
	if strings.HasPrefix(text, "/") {
		log.Printf("Detected command: %s", text)
		ts.handleCommand(ctx, chatID, userID, text)
		return
	}

	log.Printf("Handling regular message: %s", text)
	// Handle regular messages
	ts.handleMessage(ctx, chatID, userID, text)
}

// handleCommand processes bot commands
func (ts *TelegramBotService) handleCommand(ctx context.Context, chatID, userID int64, command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])
	slog.Debug("Received command", "command", cmd, "chatID", chatID, "userID", userID)

	switch cmd {
	case "/start":
		ts.handleStartCommand(ctx, chatID, userID)
	case "/help":
		ts.handleHelpCommand(ctx, chatID, userID)
	case "/subscribe":
		ts.handleSubscribeCommand(ctx, chatID, userID, parts)
	case "/unsubscribe":
		ts.handleUnsubscribeCommand(ctx, chatID, userID, parts)
	case "/list":
		ts.handleListCommand(ctx, chatID, userID)
	case "/types":
		ts.handleTypesCommand(ctx, chatID, userID)
	default:
		ts.SendMessage(chatID, "❓ Unknown command. Type /help to see available commands.")
	}
}

// handleStartCommand handles the /start command
func (ts *TelegramBotService) handleStartCommand(ctx context.Context, chatID, userID int64) {
	log.Printf("Handling /start command for user %d in chat %d", userID, chatID)

	message := `🤖 Welcome to the Notification Bot!

I can send you notifications for various services including:
• Cryptocurrency prices
• News updates  
• Weather information
• Custom alerts

Available Commands:
/help - Show this help message
/types - List all available notification types
/subscribe <type> - Subscribe to notifications
/unsubscribe <type> - Unsubscribe from notifications
/list - Show your current subscriptions

Type /types to see what notifications you can subscribe to!`

	err := ts.SendMessage(chatID, message)
	if err != nil {
		log.Printf("Failed to send start message: %v", err)
	} else {
		log.Printf("Successfully sent start message to chat %d", chatID)
	}
}

// handleHelpCommand handles the /help command
func (ts *TelegramBotService) handleHelpCommand(ctx context.Context, chatID, userID int64) {
	message := `📚 Bot Commands Help

Basic Commands:
/start - Welcome message and bot introduction
/help - Show this help message
/types - List all available notification types

Subscription Management:
/subscribe <type> - Subscribe to a notification type
/unsubscribe <type> - Unsubscribe from a notification type
/list - Show your current subscriptions

Examples:
/subscribe coinbase - Get crypto price updates
/subscribe news - Get news notifications
/subscribe weather - Get weather updates
/unsubscribe coinbase - Stop crypto notifications

Type /types to see all available notification types!`

	ts.SendMessage(chatID, message)
}

// handleSubscribeCommand handles the /subscribe command
func (ts *TelegramBotService) handleSubscribeCommand(ctx context.Context, chatID, userID int64, parts []string) {
	if len(parts) < 2 {
		message := `❓ How to Subscribe

Usage: /subscribe <notification_type>

Available types:
• coinbase - Cryptocurrency price updates
• news - Breaking news alerts  
• weather - Weather forecasts
• price_alert - Custom price alerts (requires currency and threshold)
• custom - Custom notifications

Example: /subscribe coinbase

For price alerts, you'll need to provide additional preferences after subscribing.
Type /types for more details about each type.`
		ts.SendMessage(chatID, message)
		return
	}

	notificationType := strings.ToLower(parts[1])

	// Check if notification type exists
	notificationTypeEntity, err := ts.notificationTypeService.GetTypeByCode(ctx, notificationType)
	if err != nil {
		ts.SendMessage(chatID, fmt.Sprintf("❌ Unknown notification type '%s'. Type /types to see available options.", notificationType))
		return
	}

	// Handle special cases for notification types that require preferences
	var preferences *entity.SubscriptionPreferences
	if notificationType == "price_alert" {
		// For now, set default preferences that will work
		preferences = &entity.SubscriptionPreferences{
			Currency:  "BTC",
			Threshold: 50000.0,
			Interval:  5, // 5 minutes
		}

		ts.SendMessage(chatID, "⚠️ Price alerts require specific settings. I've set default values for you:\n• Currency: BTC\n• Threshold: $50,000\n• Check interval: 5 minutes\n\nYou can modify these later if needed.")
	}

	// Subscribe user
	subscription, err := ts.subscriptionService.Subscribe(ctx, userID, chatID, notificationType, preferences)
	if err != nil {
		log.Printf("Failed to subscribe user %d to %s: %v", userID, notificationType, err)
		ts.SendMessage(chatID, "❌ Failed to subscribe. Please try again later.")
		return
	}

	var successMessage string
	if notificationType == "price_alert" {
		successMessage = fmt.Sprintf("✅ Successfully subscribed to %s notifications!\n\nDefault settings:\n• Currency: BTC\n• Threshold: $50,000\n• Interval: 5 minutes\n\nType /list to see all your subscriptions.", notificationTypeEntity.Name)
	} else {
		successMessage = fmt.Sprintf("✅ Successfully subscribed to %s notifications!\n\nYou'll receive updates based on the default interval. Type /list to see all your subscriptions.", notificationTypeEntity.Name)
	}

	ts.SendMessage(chatID, successMessage)
	log.Printf("User %d subscribed to %s (subscription ID: %d)", userID, notificationType, subscription.ID)
}

// handleUnsubscribeCommand handles the /unsubscribe command
func (ts *TelegramBotService) handleUnsubscribeCommand(ctx context.Context, chatID, userID int64, parts []string) {
	if len(parts) < 2 {
		message := `❓ How to Unsubscribe

Usage: /unsubscribe <notification_type>

Example: /unsubscribe coinbase

Type /list to see your current subscriptions.`
		ts.SendMessage(chatID, message)
		return
	}

	notificationType := strings.ToLower(parts[1])

	// Unsubscribe user
	err := ts.subscriptionService.Unsubscribe(ctx, userID, notificationType)
	if err != nil {
		log.Printf("Failed to unsubscribe user %d from %s: %v", userID, notificationType, err)
		ts.SendMessage(chatID, "❌ Failed to unsubscribe. You might not be subscribed to this type.")
		return
	}

	ts.SendMessage(chatID, fmt.Sprintf("✅ Successfully unsubscribed from %s notifications.", notificationType))
	log.Printf("User %d unsubscribed from %s", userID, notificationType)
}

// handleListCommand handles the /list command
func (ts *TelegramBotService) handleListCommand(ctx context.Context, chatID, userID int64) {
	subscriptions, err := ts.subscriptionService.GetUserSubscriptions(ctx, userID)
	if err != nil {
		log.Printf("Failed to get subscriptions for user %d: %v", userID, err)
		ts.SendMessage(chatID, "❌ Failed to retrieve your subscriptions.")
		return
	}

	if len(subscriptions) == 0 {
		message := `📝 Your Subscriptions

You're not subscribed to any notifications yet.

Type /types to see available notification types, then use /subscribe <type> to get started!`
		ts.SendMessage(chatID, message)
		return
	}

	var message strings.Builder
	message.WriteString("📝 Your Active Subscriptions:\n\n")

	for _, sub := range subscriptions {
		if sub.IsActive {
			status := "🟢"
			if !sub.IsActive {
				status = "🔴"
			}

			interval := "default"
			if sub.Preferences.Interval > 0 {
				interval = fmt.Sprintf("%d min", sub.Preferences.Interval)
			}

			message.WriteString(fmt.Sprintf("%s %s - %s\n", status, sub.NotificationType.Name, interval))

			if sub.LastNotifiedAt != nil {
				message.WriteString(fmt.Sprintf("   Last update: %s\n", sub.LastNotifiedAt.Format("Jan 2, 15:04")))
			}
			message.WriteString("\n")
		}
	}

	message.WriteString("Use /unsubscribe <type> to stop notifications.")
	ts.SendMessage(chatID, message.String())
}

// handleTypesCommand handles the /types command
func (ts *TelegramBotService) handleTypesCommand(ctx context.Context, chatID, userID int64) {
	types, err := ts.notificationTypeService.GetActiveTypes(ctx)
	if err != nil {
		log.Printf("Failed to get notification types: %v", err)
		ts.SendMessage(chatID, "❌ Failed to retrieve notification types.")
		return
	}

	var message strings.Builder
	message.WriteString("📋 Available Notification Types:\n\n")

	for _, nt := range types {
		message.WriteString(fmt.Sprintf("🔹 %s (%s)\n", nt.Name, nt.Code))
		if nt.Description != nil {
			message.WriteString(fmt.Sprintf("   %s\n", *nt.Description))
		}
		message.WriteString(fmt.Sprintf("   Default interval: %d minutes\n\n", nt.DefaultIntervalMinutes))
	}

	message.WriteString("Use /subscribe <type> to subscribe to any of these notifications.\n")
	message.WriteString("Example: /subscribe coinbase")

	ts.SendMessage(chatID, message.String())
}

// handleMessage processes regular (non-command) messages
func (ts *TelegramBotService) handleMessage(ctx context.Context, chatID, userID int64, text string) {
	// For now, just acknowledge the message
	responses := []string{
		"Thanks for your message! Use /help to see what I can do.",
		"I received your message. Type /help for available commands.",
		"Hello! I'm here to send you notifications. Use /help to get started.",
	}

	response := responses[int(userID)%len(responses)]
	ts.SendMessage(chatID, response)
}

// Helper functions

func getDisplayName(user *entity.User) string {
	if user.Username != nil && *user.Username != "" {
		return *user.Username
	}
	if user.FirstName != nil && *user.FirstName != "" {
		name := *user.FirstName
		if user.LastName != nil && *user.LastName != "" {
			name += " " + *user.LastName
		}
		return name
	}
	return fmt.Sprintf("User_%d", user.TelegramUserID)
}
