package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
)

type TelegramService struct {
	BotInstance *bot.Bot
}
type TelegramServiceInterface interface {
	SendMessage(chatID int64, message string) error
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
	}
}

func (s *TelegramService) SendMessage(chatID string, message string) error {
	chatIDInt, err := validateChatID(chatID)
	if err != nil {
		return err
	}

	// Create a context with a timeout for the API call.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send the message using the bot instance
	slog.Info("Attempting to send message to chat ID", "chatID", chatIDInt)
	msg, err := s.BotInstance.SendMessage(ctx,
		&bot.SendMessageParams{
			ChatID: chatIDInt,
			Text:   message,
		})

	if err != nil {
		slog.Error("Failed to send message", "error", err)
	}

	slog.Info("Message sent successfully!", "messageID", msg.ID)
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
