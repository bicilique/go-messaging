package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-messaging/model"
	"go-messaging/util"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
)

// TelegramService provides methods to interact with the Telegram Bot API
type TelegramService struct {
	BotInstance      *bot.Bot
	baseURL          string
	client           *http.Client
	rateLimiter      *model.RateLimiter
	messageValidator *model.MessageValidator
}
type TelegramServiceInterface interface {
	SendMessage(chatID int64, message string) error
	SendMessageByStringID(chatID string, message string) error
	SendIocMessage(chatID string, payload model.IocPayload) error
	StartPolling(ctx context.Context)
}

func NewTelegramService(botToken string) *TelegramService {
	if botToken == "" {
		panic("TELEGRAM BOT TOKEN environment variable not set.")
	}
	botInstance, err := bot.New(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	return &TelegramService{
		BotInstance:      botInstance,
		baseURL:          fmt.Sprintf("https://api.telegram.org/bot%s", botToken),
		client:           &http.Client{Timeout: 35 * time.Second},
		rateLimiter:      model.NewRateLimiter(),
		messageValidator: model.NewMessageValidator(),
	}
}

func (s *TelegramService) StartPolling(ctx context.Context) {
	go func() {
		slog.Info("Starting Telegram bot polling...")
		offset := 0

		for {
			select {
			case <-ctx.Done():
				slog.Info("Stopping Telegram bot polling...")
				return
			default:
				updates, err := s.getUpdates(ctx, offset)
				if err != nil {
					// Check if error is due to context cancellation
					if ctx.Err() != nil {
						slog.Info("Context cancelled, stopping polling...")
						return
					}
					slog.Error("Error getting updates", "error", err)
					time.Sleep(3 * time.Second)
					continue
				}

				for _, update := range updates {
					s.handleUpdate(update)
					offset = update.UpdateID + 1
				}

				time.Sleep(1 * time.Second)
			}
		}
	}()
}

// SendMessage sends a message to a specified chat ID
func (s *TelegramService) SendMessage(chatID int64, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.Info("Attempting to send message to chat ID", "chatID", chatID)
	msg, err := s.BotInstance.SendMessage(ctx,
		&bot.SendMessageParams{
			ChatID: chatID,
			Text:   message,
		})

	if err != nil {
		slog.Error("Failed to send message", "error", err)
		return err
	}

	slog.Info("Message sent successfully!", "messageID", msg.ID)
	return nil
}

// SendMessageByStringID is a helper method for string chat IDs
func (s *TelegramService) SendMessageByStringID(chatID string, message string) error {
	chatIDInt, err := validateChatID(chatID)
	if err != nil {
		return err
	}
	return s.SendMessage(chatIDInt, message)
}

// SendIocMessage sends an IOC message to a specified chat ID
func (s *TelegramService) SendIocMessage(chatID string, payload model.IocPayload) error {
	chatIDInt, err := validateChatID(chatID)
	if err != nil {
		return err
	}

	// Create a context with a timeout for the API call.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send the message using the bot instance
	slog.Info("Attempting to send IOC message to chat ID", "chatID", chatIDInt)
	msg, err := s.BotInstance.SendMessage(ctx,
		&bot.SendMessageParams{
			ChatID:    chatIDInt,
			Text:      util.FormatIocMessage(payload),
			ParseMode: "MarkdownV2", // Use Markdown V2 for formatting
		})

	if err != nil {
		slog.Error("Failed to send IOC message", "error", err)
		return err
	}

	slog.Info("IOC message sent successfully!", "messageID", msg.ID)
	return nil
}

func validateChatID(chatID string) (int64, error) {
	// Handle group chat IDs that start with "-"
	if chatID == "" {
		return 0, fmt.Errorf("chat ID cannot be empty")
	}

	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid chat ID format: %v", err)
	}

	// Log the type of chat for debugging
	if chatIDInt < 0 {
		slog.Info("Sending to group chat", "chatID", chatIDInt)
	} else {
		slog.Info("Sending to individual chat", "chatID", chatIDInt)
	}

	return chatIDInt, nil
}

// getUpdates retrieves updates from the Telegram API with context support
func (t *TelegramService) getUpdates(ctx context.Context, offset int) ([]model.Update, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=10", t.baseURL, offset)

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		OK     bool           `json:"ok"`
		Result []model.Update `json:"result"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

func (s *TelegramService) handleUpdate(update model.Update) {
	if update.Message == nil {
		return
	}

	message := update.Message

	// Check rate limiting first
	allowed, limitMsg := s.rateLimiter.IsAllowed(int64(message.From.ID))
	if !allowed {
		s.SendMessage(message.Chat.ID, limitMsg)
		slog.Warn("Rate limit exceeded",
			"userID", message.From.ID,
			"username", message.From.Username)
		return
	}

	// Validate message length using the validator
	valid, validationMsg := s.messageValidator.ValidateLength(*message)
	if !valid {
		s.SendMessage(message.Chat.ID, validationMsg)
		slog.Warn("Message validation failed",
			"userID", message.From.ID,
			"username", message.From.Username,
			"length", len(message.Text))
		return
	}

	// Check if it's a command
	if message.Text != "" && message.Text[0] == '/' {
		s.handleCommand(*message)
	} else if message.Text != "" {
		s.handleMessage(*message)
	}
}

func (s *TelegramService) handleCommand(message model.Message) {
	// Extract command and arguments
	parts := strings.Fields(message.Text)
	command := parts[0]

	switch command {
	case "/start":
		welcomeMsg := "🤖 Welcome! Bot started successfully.\n\n" +
			"📋 Available commands:\n" +
			"/start - Start the bot\n" +
			"/help - Show help\n" +
			"/limits - Show message limits\n" +
			"/status - Bot status\n" +
			"/info - Bot information\n\n" +
			"💬 Just send me any message and I'll echo it back!"
		s.SendMessage(message.Chat.ID, welcomeMsg)

	case "/help":
		helpMsg := "📋 Available commands:\n" +
			"/start - Start the bot\n" +
			"/help - Show this help\n" +
			"/limits - Show message limits\n" +
			"/status - Bot status\n" +
			"/info - Bot information\n\n" +
			"💬 Send me any text message and I'll echo it back!"
		s.SendMessage(message.Chat.ID, helpMsg)

	case "/limits":
		// Get limits info from helpers
		msgLimits := s.messageValidator.GetLimitsInfo()
		rateLimits := s.rateLimiter.GetStats()

		limitsMsg := fmt.Sprintf("📏 Message Limits:\n\n"+
			"• Regular messages: %d characters max\n"+
			"• Commands: %d characters max\n"+
			"• Telegram limit: %d characters max\n"+
			"• Rate limit: %d messages per minute\n"+
			"• Minimum interval: %v between messages\n\n"+
			"📊 Your last message: %d characters\n"+
			"👥 Active users being tracked: %d",
			msgLimits["regular_messages"], msgLimits["commands"], msgLimits["telegram_max"],
			rateLimits["messages_limit"], rateLimits["min_interval"], len(message.Text),
			rateLimits["active_users"])
		s.SendMessage(message.Chat.ID, limitsMsg)

	case "/status":
		s.SendMessage(message.Chat.ID, "🟢 Bot is running normally!")

	case "/info":
		s.SendMessage(message.Chat.ID, "ℹ️ Go Messaging Bot v1.0\n🔧 With rate limiting and message validation")

	default:
		s.SendMessage(message.Chat.ID, "❓ Unknown command. Type /help for available commands.")
	}
}

func (s *TelegramService) handleMessage(message model.Message) {
	// Show character count for longer messages
	charCount := len(message.Text)
	var response string

	if charCount > 100 {
		response = fmt.Sprintf("💬 You said (%d chars): %s", charCount, message.Text)
	} else {
		response = fmt.Sprintf("💬 You said: %s", message.Text)
	}

	s.SendMessage(message.Chat.ID, response)
}
