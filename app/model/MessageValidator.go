package model

import "fmt"

// Message validation configuration constants
const (
	MAX_MESSAGE_LENGTH      = 4096 // Telegram's limit
	MAX_COMMAND_LENGTH      = 256  // Command length limit
	MAX_USER_MESSAGE_LENGTH = 1000 // Custom limit for user messages
)

type MessageValidator struct{}

func NewMessageValidator() *MessageValidator {
	return &MessageValidator{}
}

// ValidateLength checks if a message meets length requirements
func (mv *MessageValidator) ValidateLength(message Message) (bool, string) {
	if len(message.Text) == 0 {
		return true, "" // Allow empty messages (could be media)
	}

	// Check command length
	if message.Text[0] == '/' {
		if len(message.Text) > MAX_COMMAND_LENGTH {
			return false, fmt.Sprintf("❌ Command too long! Maximum %d characters allowed.", MAX_COMMAND_LENGTH)
		}
		return true, ""
	}

	// Check regular message length
	if len(message.Text) > MAX_USER_MESSAGE_LENGTH {
		return false, fmt.Sprintf("❌ Message too long! Maximum %d characters allowed.\nYour message: %d characters",
			MAX_USER_MESSAGE_LENGTH, len(message.Text))
	}

	return true, ""
}

// GetLimitsInfo returns information about current message limits
func (mv *MessageValidator) GetLimitsInfo() map[string]interface{} {
	return map[string]interface{}{
		"regular_messages": MAX_USER_MESSAGE_LENGTH,
		"commands":         MAX_COMMAND_LENGTH,
		"telegram_max":     MAX_MESSAGE_LENGTH,
	}
}

// ValidateMessageString validates a simple message string
func ValidateMessageString(message string) error {
	if len(message) == 0 {
		return fmt.Errorf("message cannot be empty")
	}

	if len(message) > MAX_MESSAGE_LENGTH {
		return fmt.Errorf("message too long: %d characters (max %d)", len(message), MAX_MESSAGE_LENGTH)
	}

	return nil
}
