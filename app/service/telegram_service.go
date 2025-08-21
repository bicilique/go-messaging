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
	"time"

	"github.com/go-telegram/bot"
)

type TelegramService struct {
	BotInstance *bot.Bot
	baseURL     string
	client      *http.Client
}
type TelegramServiceInterface interface {
	SendMessage(chatID int64, message string) error
	SendIocMessage(chatID string, payload model.IocPayload) error
	StartPolling(ctx context.Context)
}

func NewTelegramService(botToken string) *TelegramService {
	if botToken == "" {
		panic("TELEGRAM BOT TOKEN environment variable not set.")
	}

	// Create a new bot instance using the provided token.
	botInstance, err := bot.New(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	return &TelegramService{
		BotInstance: botInstance,
		baseURL:     fmt.Sprintf("https://api.telegram.org/bot%s", botToken),
		client:      &http.Client{Timeout: 35 * time.Second},
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
				updates, err := s.getUpdates(offset)
				if err != nil {
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

func (s *TelegramService) SendMessage(chatID int64, message string) error {
	// Create a context with a timeout for the API call.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send the message using the bot instance
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

// getUpdates retrieves updates from the Telegram API
func (t *TelegramService) getUpdates(offset int) ([]model.Update, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=30", t.baseURL, offset)

	resp, err := t.client.Get(url)
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

	// Check if it's a command
	if message.Text != "" && message.Text[0] == '/' {
		s.handleCommand(*message)
	} else if message.Text != "" {
		s.handleMessage(*message)
	}
}

func (t *TelegramService) handleCommand(message model.Message) {
	switch message.Text {
	case "/start":
		t.SendMessage(message.Chat.ID, "🤖 Welcome! Bot started successfully.\n\nAvailable commands:\n/start - Start the bot\n/help - Show help")
	case "/help":
		t.SendMessage(message.Chat.ID, "📋 Available commands:\n/start - Start the bot\n/help - Show this help\n\nJust send me any message and I'll echo it back!")
	default:
		t.SendMessage(message.Chat.ID, "❓ Unknown command. Type /help for available commands.")
	}
}

func (s *TelegramService) handleMessage(message model.Message) {
	response := fmt.Sprintf("💬 You said: %s", message.Text)
	s.SendMessage(message.Chat.ID, response)
}
