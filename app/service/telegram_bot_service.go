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
	adminService            AdminServiceInterface
	telegramAdminService    *TelegramAdminService
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
	adminService AdminServiceInterface,
) *TelegramBotService {
	if botToken == "" {
		panic("TELEGRAM BOT TOKEN environment variable not set.")
	}

	botInstance, err := bot.New(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	service := &TelegramBotService{
		botInstance:             botInstance,
		rateLimiter:             model.NewRateLimiter(),
		messageValidator:        model.NewMessageValidator(),
		userService:             userService,
		subscriptionService:     subscriptionService,
		notificationTypeService: notificationTypeService,
		adminService:            adminService,
	}

	// Initialize telegram admin service
	if userService != nil && adminService != nil {
		// Use the TelegramBotService itself as it implements TelegramNotificationSender
		service.telegramAdminService = NewTelegramAdminService(service, adminService, userService)
	}

	return service
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

	// Register handler for callback queries
	ts.botInstance.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return update.CallbackQuery != nil
	}, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		ts.HandleUpdate(ctx, b, update)
	})

	ts.botInstance.Start(ctx)
}

// HandleUpdate processes incoming updates from Telegram
func (ts *TelegramBotService) HandleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("Received update: %+v", update)

	// Handle callback queries first
	if update.CallbackQuery != nil {
		ts.handleCallbackQuery(ctx, update.CallbackQuery)
		return
	}

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
	case "/admin":
		slog.Info("[DEBUG] /admin command detected in TelegramBotService", "userID", userID, "chatID", chatID)
		ts.handleAdminCommand(ctx, chatID, userID, command)
	default:
		ts.SendMessage(chatID, "❓ Unknown command. Type /help to see available commands.")
	}
}

// handleStartCommand handles the /start command
func (ts *TelegramBotService) handleStartCommand(ctx context.Context, chatID, userID int64) {
	log.Printf("Handling /start command for user %d in chat %d", userID, chatID)

	message := `🤖 Welcome to Go Messaging Bot!

I can send you notifications for various services including:
• 🪙 Cryptocurrency prices
• 📰 News updates  
• 🌤️ Weather information
• 🔔 Custom alerts

Available Commands:
• /types - List all notification types
• /subscribe <type> - Subscribe to notifications
• /unsubscribe <type> - Unsubscribe from notifications
• /list - Show your subscriptions
• /help - Show help menu

Examples:
• /subscribe coinbase - Get crypto updates
• /subscribe news - Get news notifications
• /unsubscribe weather - Stop weather updates`

	// Create inline keyboard with quick actions
	keyboard := model.InlineKeyboardMarkup{
		InlineKeyboard: [][]model.InlineKeyboardButton{
			{
				{Text: "📋 View Types", CallbackData: "types:all"},
				{Text: "📱 My Subscriptions", CallbackData: "list:mine"},
			},
			{
				{Text: "❓ Help", CallbackData: "help:main"},
			},
		},
	}

	// Check if user is admin and add admin button
	if ts.userService != nil {
		if user, err := ts.userService.GetUserByTelegramID(ctx, userID); err == nil && user.Role == "admin" {
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []model.InlineKeyboardButton{
				{Text: "🔧 Admin Panel", CallbackData: "admin:main"},
			})
		}
	}

	err := ts.SendMessageWithKeyboard(chatID, message, keyboard)
	if err != nil {
		log.Printf("Failed to send start message: %v", err)
	} else {
		log.Printf("Successfully sent start message to chat %d", chatID)
	}
}

// handleHelpCommand handles the /help command
func (ts *TelegramBotService) handleHelpCommand(ctx context.Context, chatID, userID int64) {
	message := `📚 Help & Support

Getting Started:
1. Use /start to see the main menu
2. Browse available notification types
3. Subscribe to notifications you want!

Main Commands:
• /start - Welcome message and bot introduction
• /help - Show this help message  
• /types - List all notification types

Subscription Management:
• /subscribe <type> - Subscribe to notifications
• /unsubscribe <type> - Unsubscribe from notifications
• /list - Show your current subscriptions

Examples:
• /subscribe coinbase - Get crypto updates
• /subscribe news - Get news notifications
• /subscribe weather - Get weather updates
• /unsubscribe coinbase - Stop crypto notifications`

	// Create help keyboard
	keyboard := model.InlineKeyboardMarkup{
		InlineKeyboard: [][]model.InlineKeyboardButton{
			{
				{Text: "📋 View Types", CallbackData: "types:all"},
				{Text: "📱 My Subscriptions", CallbackData: "list:mine"},
			},
		},
	}

	// Check if user is admin and add admin commands
	if ts.userService != nil {
		if user, err := ts.userService.GetUserByTelegramID(ctx, userID); err == nil && user.Role == "admin" {
			message += `

🔧 Admin Commands:
• /admin - Access admin panel for user management`

			// Add admin button
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []model.InlineKeyboardButton{
				{Text: "🔧 Admin Panel", CallbackData: "admin:main"},
			})
		}
	}

	ts.SendMessageWithKeyboard(chatID, message, keyboard)
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

Use the buttons below to get started!`

		keyboard := model.InlineKeyboardMarkup{
			InlineKeyboard: [][]model.InlineKeyboardButton{
				{
					{Text: "📋 View Available Types", CallbackData: "types:all"},
				},
				{
					{Text: "❓ Help", CallbackData: "help:main"},
				},
			},
		}
		ts.SendMessageWithKeyboard(chatID, message, keyboard)
		return
	}

	var message strings.Builder
	message.WriteString(fmt.Sprintf("📝 Your Active Subscriptions (%d):\n\n", len(subscriptions)))

	// Create keyboard with unsubscribe buttons
	keyboard := model.InlineKeyboardMarkup{
		InlineKeyboard: [][]model.InlineKeyboardButton{},
	}

	for _, sub := range subscriptions {
		if sub.IsActive {
			status := "�"
			interval := "default"
			if sub.Preferences.Interval > 0 {
				interval = fmt.Sprintf("%d min", sub.Preferences.Interval)
			}

			message.WriteString(fmt.Sprintf("%s %s - %s\n", status, sub.NotificationType.Name, interval))

			if sub.LastNotifiedAt != nil {
				message.WriteString(fmt.Sprintf("   📅 Last update: %s\n", sub.LastNotifiedAt.Format("Jan 2, 15:04")))
			}
			message.WriteString("\n")

			// Add unsubscribe button for each active subscription
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []model.InlineKeyboardButton{
				{Text: fmt.Sprintf("❌ Unsubscribe from %s", sub.NotificationType.Name), CallbackData: fmt.Sprintf("unsubscribe:%s", sub.NotificationType.Code)},
			})
		}
	}

	message.WriteString("Click the buttons below to unsubscribe, or use `/unsubscribe <type>`")

	// Add navigation buttons
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []model.InlineKeyboardButton{
		{Text: "📋 Browse Types", CallbackData: "types:all"},
		{Text: "❓ Help", CallbackData: "help:main"},
	})

	ts.SendMessageWithKeyboard(chatID, message.String(), keyboard)
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

	// Create keyboard with subscribe buttons
	keyboard := model.InlineKeyboardMarkup{
		InlineKeyboard: [][]model.InlineKeyboardButton{},
	}

	for _, nt := range types {
		message.WriteString(fmt.Sprintf("🔹 %s (%s)\n", nt.Name, nt.Code))
		if nt.Description != nil {
			message.WriteString(fmt.Sprintf("   %s\n", *nt.Description))
		}
		message.WriteString(fmt.Sprintf("   📊 Default interval: %d minutes\n\n", nt.DefaultIntervalMinutes))

		// Add subscribe button for each type
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []model.InlineKeyboardButton{
			{Text: fmt.Sprintf("✅ Subscribe to %s", nt.Name), CallbackData: fmt.Sprintf("subscribe:%s", nt.Code)},
		})
	}

	message.WriteString("Click the buttons below to subscribe, or use `/subscribe <type>`\n")
	message.WriteString("Example: `/subscribe coinbase`")

	// Add navigation buttons
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []model.InlineKeyboardButton{
		{Text: "📱 My Subscriptions", CallbackData: "list:mine"},
		{Text: "❓ Help", CallbackData: "help:main"},
	})

	ts.SendMessageWithKeyboard(chatID, message.String(), keyboard)
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

// SendMessageWithKeyboard sends a message with an inline keyboard
func (ts *TelegramBotService) SendMessageWithKeyboard(chatID int64, message string, keyboard model.InlineKeyboardMarkup) error {
	// Validate message
	if err := model.ValidateMessageString(message); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert model.InlineKeyboardMarkup to bot package format
	var botKeyboard [][]models.InlineKeyboardButton
	for _, row := range keyboard.InlineKeyboard {
		var botRow []models.InlineKeyboardButton
		for _, button := range row {
			botButton := models.InlineKeyboardButton{
				Text: button.Text,
			}
			if button.CallbackData != "" {
				botButton.CallbackData = button.CallbackData
			}
			if button.URL != "" {
				botButton.URL = button.URL
			}
			botRow = append(botRow, botButton)
		}
		botKeyboard = append(botKeyboard, botRow)
	}

	replyMarkup := &models.InlineKeyboardMarkup{
		InlineKeyboard: botKeyboard,
	}

	_, err := ts.botInstance.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        message,
		ReplyMarkup: replyMarkup,
	})

	return err
}

// AnswerCallbackQuery answers a callback query (public interface method)
func (ts *TelegramBotService) AnswerCallbackQuery(callbackID, text string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ts.answerCallbackQuery(ctx, callbackID, text)
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

// handleAdminCommand handles admin commands with proper role checking
func (ts *TelegramBotService) handleAdminCommand(ctx context.Context, chatID, userID int64, command string) {
	slog.Info("[DEBUG] Admin command received in TelegramBotService", "command", command, "userID", userID)

	if ts.userService == nil {
		ts.SendMessage(chatID, "❌ User service is not available")
		return
	}

	// Check if user exists and has admin role
	user, err := ts.userService.GetUserByTelegramID(ctx, userID)
	if err != nil {
		ts.SendMessage(chatID, "❌ You need to register first. Use /start command.")
		return
	}

	if user.Role != "admin" {
		ts.SendMessage(chatID, "❌ You don't have admin permissions.")
		slog.Info("Non-admin user attempted admin command", "userID", userID, "role", user.Role)
		return
	}

	// Show admin panel with buttons
	ts.showAdminPanel(ctx, chatID, userID)
}

// Debug method to check if services are properly initialized
func (ts *TelegramBotService) CheckAdminServices() {
	slog.Info("TelegramBotService Admin Services Status",
		"userService", ts.userService != nil,
		"adminService", ts.adminService != nil,
		"telegramAdminService", ts.telegramAdminService != nil,
	)
}

// handleCallbackQuery handles callback queries from inline keyboards
func (ts *TelegramBotService) handleCallbackQuery(ctx context.Context, callbackQuery *models.CallbackQuery) {
	log.Printf("Received callback query: %s from user %d", callbackQuery.Data, callbackQuery.From.ID)

	// Answer the callback query first
	ts.answerCallbackQuery(ctx, callbackQuery.ID, "")

	// Parse callback data
	data := callbackQuery.Data
	parts := strings.Split(data, ":")

	if len(parts) < 2 {
		log.Printf("Invalid callback data format: %s", data)
		return
	}

	action := parts[0]
	param := parts[1]

	// Extract chat ID - for callback queries, we need to get it from the original message
	// For now, let's try to extract it from the From ID (assuming private chat)
	chatID := callbackQuery.From.ID
	userID := callbackQuery.From.ID

	switch action {
	case "subscribe":
		ts.handleSubscribeCallback(ctx, chatID, userID, param)
	case "unsubscribe":
		ts.handleUnsubscribeCallback(ctx, chatID, userID, param)
	case "list":
		ts.handleListCommand(ctx, chatID, userID)
	case "types":
		ts.handleTypesCommand(ctx, chatID, userID)
	case "help":
		ts.handleHelpCommand(ctx, chatID, userID)
	case "admin":
		if param == "main" {
			ts.handleAdminCommand(ctx, chatID, userID, "/admin")
		} else if param == "pending" || param == "approved" || param == "stats" || param == "cleanup" {
			ts.handleAdminCallback(ctx, chatID, userID, fmt.Sprintf("/admin_%s", param))
		} else {
			ts.handleAdminCallback(ctx, chatID, userID, "/admin")
		}
	default:
		log.Printf("Unknown callback action: %s", action)
	}
}

// answerCallbackQuery answers a callback query
func (ts *TelegramBotService) answerCallbackQuery(ctx context.Context, callbackQueryID, text string) error {
	_, err := ts.botInstance.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callbackQueryID,
		Text:            text,
	})
	return err
}

// handleSubscribeCallback handles subscription via button callback
func (ts *TelegramBotService) handleSubscribeCallback(ctx context.Context, chatID, userID int64, notificationType string) {
	log.Printf("Subscribe callback: user %d wants to subscribe to %s", userID, notificationType)

	// Use the existing subscribe logic
	parts := []string{"/subscribe", notificationType}
	ts.handleSubscribeCommand(ctx, chatID, userID, parts)
}

// handleUnsubscribeCallback handles unsubscription via button callback
func (ts *TelegramBotService) handleUnsubscribeCallback(ctx context.Context, chatID, userID int64, notificationType string) {
	log.Printf("Unsubscribe callback: user %d wants to unsubscribe from %s", userID, notificationType)

	// Use the existing unsubscribe logic
	parts := []string{"/unsubscribe", notificationType}
	ts.handleUnsubscribeCommand(ctx, chatID, userID, parts)
}

// handleAdminCallback handles admin actions via button callback
func (ts *TelegramBotService) handleAdminCallback(ctx context.Context, chatID, userID int64, command string) {
	log.Printf("Admin callback: user %d executed %s", userID, command)

	// Use the existing admin logic
	ts.handleAdminCommand(ctx, chatID, userID, command)
}

// showAdminPanel displays the admin panel with buttons
func (ts *TelegramBotService) showAdminPanel(ctx context.Context, chatID, userID int64) {
	message := `🔧 Admin Panel

Welcome to the admin panel! Here you can manage users and system settings.

Available Actions:
• View pending user registrations
• Manage approved users  
• View system statistics
• Perform cleanup operations`

	keyboard := model.InlineKeyboardMarkup{
		InlineKeyboard: [][]model.InlineKeyboardButton{
			{
				{Text: "👥 Pending Users", CallbackData: "admin:pending"},
				{Text: "✅ Approved Users", CallbackData: "admin:approved"},
			},
			{
				{Text: "📊 Statistics", CallbackData: "admin:stats"},
				{Text: "🧹 Cleanup", CallbackData: "admin:cleanup"},
			},
			{
				{Text: "🏠 Back to Main Menu", CallbackData: "help:main"},
			},
		},
	}

	if ts.telegramAdminService != nil {
		// Get some quick stats to show
		message += `

Quick Actions:
Use the buttons below to perform admin tasks quickly.`
	} else {
		message += `

⚠️ Note: Some admin features may be limited.`
	}

	ts.SendMessageWithKeyboard(chatID, message, keyboard)
}
