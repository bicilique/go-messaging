package service

import (
	"context"
	"fmt"
	"go-messaging/model"
	"log/slog"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type TelegramAdminService struct {
	telegramService TelegramNotificationSender
	adminService    AdminServiceInterface
	userService     UserService
}

func NewTelegramAdminService(
	telegramService TelegramNotificationSender,
	adminService AdminServiceInterface,
	userService UserService,
) *TelegramAdminService {
	return &TelegramAdminService{
		telegramService: telegramService,
		adminService:    adminService,
		userService:     userService,
	}
}

func (s *TelegramAdminService) HandleAdminCommand(ctx context.Context, message model.Message) {
	isAdmin, err := s.adminService.IsAdmin(ctx, int64(message.From.ID))
	if err != nil {
		slog.Error("Failed to check admin status", "userID", message.From.ID, "error", err)
		s.telegramService.SendMessage(message.Chat.ID, "❌ Error checking admin permissions")
		return
	}

	if !isAdmin {
		s.telegramService.SendMessage(message.Chat.ID, "❌ You don't have admin permissions")
		return
	}

	parts := strings.Fields(message.Text)
	command := parts[0]

	switch command {
	case "/admin":
		s.showAdminMenu(ctx, message.Chat.ID)
	case "/admin_pending":
		s.showPendingUsers(ctx, message.Chat.ID)
	case "/admin_approved":
		s.showApprovedUsers(ctx, message.Chat.ID)
	case "/admin_stats":
		s.showUserStats(ctx, message.Chat.ID)
	case "/admin_cleanup":
		s.cleanupPendingUsers(ctx, message.Chat.ID)
	default:
		s.telegramService.SendMessage(message.Chat.ID, "❓ Unknown admin command. Use /admin to see available options.")
	}
}

func (s *TelegramAdminService) HandleCallbackQuery(ctx context.Context, callback model.CallbackQuery) {
	isAdmin, err := s.adminService.IsAdmin(ctx, int64(callback.From.ID))
	if err != nil {
		slog.Error("Failed to check admin status for callback", "userID", callback.From.ID, "error", err)
		return
	}

	if !isAdmin {
		s.answerCallbackQuery(callback.ID, "❌ You don't have admin permissions")
		return
	}

	data := callback.Data
	parts := strings.Split(data, ":")

	if len(parts) < 2 {
		s.answerCallbackQuery(callback.ID, "❌ Invalid callback data")
		return
	}

	action := parts[0]
	param := parts[1]

	switch action {
	case "admin_menu":
		s.handleAdminMenuCallback(ctx, callback, param)
	case "approve_user":
		s.handleUserApproval(ctx, callback, param, true)
	case "reject_user":
		s.handleUserApproval(ctx, callback, param, false)
	case "disable_user":
		s.handleUserDisable(ctx, callback, param)
	case "enable_user":
		s.handleUserEnable(ctx, callback, param)
	case "view_user":
		s.handleViewUser(ctx, callback, param)
	default:
		s.answerCallbackQuery(callback.ID, "❌ Unknown action")
	}
}

func (s *TelegramAdminService) showAdminMenu(ctx context.Context, chatID int64) {
	keyboard := model.InlineKeyboardMarkup{
		InlineKeyboard: [][]model.InlineKeyboardButton{
			{
				{Text: "📋 Pending Users", CallbackData: "admin_menu:pending"},
				{Text: "✅ Approved Users", CallbackData: "admin_menu:approved"},
			},
			{
				{Text: "📊 User Stats", CallbackData: "admin_menu:stats"},
				{Text: "🧹 Cleanup", CallbackData: "admin_menu:cleanup"},
			},
		},
	}

	message := "🔧 **Admin Panel**\n\n" +
		"Welcome to the admin panel. Choose an option below:"

	s.sendMessageWithKeyboard(chatID, message, keyboard)
}

func (s *TelegramAdminService) showPendingUsers(ctx context.Context, chatID int64) {
	users, err := s.adminService.GetPendingUsers(ctx)
	if err != nil {
		slog.Error("Failed to get pending users", "error", err)
		s.telegramService.SendMessage(chatID, "❌ Failed to get pending users")
		return
	}

	if len(users) == 0 {
		s.telegramService.SendMessage(chatID, "✨ No pending users found!")
		return
	}

	message := "📋 **Pending Users** (" + strconv.Itoa(len(users)) + "):\n\n"

	keyboard := model.InlineKeyboardMarkup{
		InlineKeyboard: [][]model.InlineKeyboardButton{},
	}

	for i, user := range users {
		if i >= 10 { // Limit to 10 users per message
			break
		}

		username := "N/A"
		if user.Username != nil {
			username = *user.Username
		}

		firstName := "N/A"
		if user.FirstName != nil {
			firstName = *user.FirstName
		}

		message += fmt.Sprintf("👤 **%s** (@%s)\n", firstName, username)
		message += fmt.Sprintf("📅 Joined: %s\n", user.CreatedAt.Format("2006-01-02 15:04"))
		message += fmt.Sprintf("🆔 ID: `%s`\n\n", user.ID.String())

		// Add action buttons for each user
		row := []model.InlineKeyboardButton{
			{Text: "✅ Approve", CallbackData: fmt.Sprintf("approve_user:%s", user.ID.String())},
			{Text: "❌ Reject", CallbackData: fmt.Sprintf("reject_user:%s", user.ID.String())},
		}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}

	if len(users) > 10 {
		message += fmt.Sprintf("... and %d more users", len(users)-10)
	}

	// Add back button
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []model.InlineKeyboardButton{
		{Text: "🔙 Back to Menu", CallbackData: "admin_menu:main"},
	})

	s.sendMessageWithKeyboard(chatID, message, keyboard)
}

func (s *TelegramAdminService) showApprovedUsers(ctx context.Context, chatID int64) {
	users, err := s.adminService.GetApprovedUsers(ctx, 10)
	if err != nil {
		slog.Error("Failed to get approved users", "error", err)
		s.telegramService.SendMessage(chatID, "❌ Failed to get approved users")
		return
	}

	if len(users) == 0 {
		s.telegramService.SendMessage(chatID, "📭 No approved users found!")
		return
	}

	message := "✅ **Approved Users** (showing last " + strconv.Itoa(len(users)) + "):\n\n"

	keyboard := model.InlineKeyboardMarkup{
		InlineKeyboard: [][]model.InlineKeyboardButton{},
	}

	for _, user := range users {
		username := "N/A"
		if user.Username != nil {
			username = *user.Username
		}

		firstName := "N/A"
		if user.FirstName != nil {
			firstName = *user.FirstName
		}

		approvedDate := "N/A"
		if user.ApprovedAt != nil {
			approvedDate = user.ApprovedAt.Format("2006-01-02 15:04")
		}

		message += fmt.Sprintf("👤 **%s** (@%s)\n", firstName, username)
		message += fmt.Sprintf("✅ Approved: %s\n", approvedDate)
		message += fmt.Sprintf("🆔 ID: `%s`\n\n", user.ID.String())

		// Add action button for each user
		row := []model.InlineKeyboardButton{
			{Text: "🚫 Disable", CallbackData: fmt.Sprintf("disable_user:%s", user.ID.String())},
			{Text: "👁️ View", CallbackData: fmt.Sprintf("view_user:%s", user.ID.String())},
		}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}

	// Add back button
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []model.InlineKeyboardButton{
		{Text: "🔙 Back to Menu", CallbackData: "admin_menu:main"},
	})

	s.sendMessageWithKeyboard(chatID, message, keyboard)
}

func (s *TelegramAdminService) showUserStats(ctx context.Context, chatID int64) {
	stats, err := s.adminService.GetUserStats(ctx)
	if err != nil {
		slog.Error("Failed to get user stats", "error", err)
		s.telegramService.SendMessage(chatID, "❌ Failed to get user statistics")
		return
	}

	message := "📊 **User Statistics**\n\n"
	message += fmt.Sprintf("⏳ Pending: %d\n", stats["pending"])
	message += fmt.Sprintf("✅ Approved: %d\n", stats["approved"])
	message += fmt.Sprintf("❌ Rejected: %d\n", stats["rejected"])
	message += fmt.Sprintf("🚫 Disabled: %d\n", stats["disabled"])
	message += fmt.Sprintf("👑 Admins: %d\n", stats["admins"])
	message += fmt.Sprintf("\n📈 Total Users: %d", stats["pending"]+stats["approved"]+stats["rejected"]+stats["disabled"])

	keyboard := model.InlineKeyboardMarkup{
		InlineKeyboard: [][]model.InlineKeyboardButton{
			{
				{Text: "🔙 Back to Menu", CallbackData: "admin_menu:main"},
			},
		},
	}

	s.sendMessageWithKeyboard(chatID, message, keyboard)
}

func (s *TelegramAdminService) cleanupPendingUsers(ctx context.Context, chatID int64) {
	count, err := s.adminService.CleanupPendingUsers(ctx)
	if err != nil {
		slog.Error("Failed to cleanup pending users", "error", err)
		s.telegramService.SendMessage(chatID, "❌ Failed to cleanup pending users")
		return
	}

	message := fmt.Sprintf("🧹 **Cleanup Complete**\n\nRemoved %d pending users older than 6 hours.", count)

	keyboard := model.InlineKeyboardMarkup{
		InlineKeyboard: [][]model.InlineKeyboardButton{
			{
				{Text: "🔙 Back to Menu", CallbackData: "admin_menu:main"},
			},
		},
	}

	s.sendMessageWithKeyboard(chatID, message, keyboard)
}

func (s *TelegramAdminService) handleAdminMenuCallback(ctx context.Context, callback model.CallbackQuery, param string) {
	switch param {
	case "main":
		s.showAdminMenu(ctx, callback.Message.Chat.ID)
	case "pending":
		s.showPendingUsers(ctx, callback.Message.Chat.ID)
	case "approved":
		s.showApprovedUsers(ctx, callback.Message.Chat.ID)
	case "stats":
		s.showUserStats(ctx, callback.Message.Chat.ID)
	case "cleanup":
		s.cleanupPendingUsers(ctx, callback.Message.Chat.ID)
	}
	s.answerCallbackQuery(callback.ID, "")
}

func (s *TelegramAdminService) handleUserApproval(ctx context.Context, callback model.CallbackQuery, userIDStr string, approve bool) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.answerCallbackQuery(callback.ID, "❌ Invalid user ID")
		return
	}

	// Get admin user
	admin, err := s.userService.GetUserByTelegramID(ctx, int64(callback.From.ID))
	if err != nil {
		s.answerCallbackQuery(callback.ID, "❌ Failed to get admin info")
		return
	}

	var action string
	var actionText string
	if approve {
		err = s.adminService.ApproveUser(ctx, userID, admin.ID)
		action = "approved"
		actionText = "✅ Approved"
	} else {
		err = s.adminService.RejectUser(ctx, userID, admin.ID)
		action = "rejected"
		actionText = "❌ Rejected"
	}

	if err != nil {
		s.answerCallbackQuery(callback.ID, fmt.Sprintf("❌ Failed to %s user", action))
		return
	}

	s.answerCallbackQuery(callback.ID, fmt.Sprintf("%s user successfully", actionText))

	// Refresh the pending users list
	s.showPendingUsers(ctx, callback.Message.Chat.ID)
}

func (s *TelegramAdminService) handleUserDisable(ctx context.Context, callback model.CallbackQuery, userIDStr string) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.answerCallbackQuery(callback.ID, "❌ Invalid user ID")
		return
	}

	// Get admin user
	admin, err := s.userService.GetUserByTelegramID(ctx, int64(callback.From.ID))
	if err != nil {
		s.answerCallbackQuery(callback.ID, "❌ Failed to get admin info")
		return
	}

	err = s.adminService.DisableUser(ctx, userID, admin.ID)
	if err != nil {
		s.answerCallbackQuery(callback.ID, "❌ Failed to disable user")
		return
	}

	s.answerCallbackQuery(callback.ID, "🚫 User disabled successfully")

	// Refresh the approved users list
	s.showApprovedUsers(ctx, callback.Message.Chat.ID)
}

func (s *TelegramAdminService) handleUserEnable(ctx context.Context, callback model.CallbackQuery, userIDStr string) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.answerCallbackQuery(callback.ID, "❌ Invalid user ID")
		return
	}

	// Get admin user
	admin, err := s.userService.GetUserByTelegramID(ctx, int64(callback.From.ID))
	if err != nil {
		s.answerCallbackQuery(callback.ID, "❌ Failed to get admin info")
		return
	}

	err = s.adminService.EnableUser(ctx, userID, admin.ID)
	if err != nil {
		s.answerCallbackQuery(callback.ID, "❌ Failed to enable user")
		return
	}

	s.answerCallbackQuery(callback.ID, "✅ User enabled successfully")
}

func (s *TelegramAdminService) handleViewUser(ctx context.Context, callback model.CallbackQuery, userIDStr string) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.answerCallbackQuery(callback.ID, "❌ Invalid user ID")
		return
	}

	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		s.answerCallbackQuery(callback.ID, "❌ User not found")
		return
	}

	username := "N/A"
	if user.Username != nil {
		username = *user.Username
	}

	firstName := "N/A"
	if user.FirstName != nil {
		firstName = *user.FirstName
	}

	lastName := "N/A"
	if user.LastName != nil {
		lastName = *user.LastName
	}

	message := "👤 **User Details**\n\n"
	message += fmt.Sprintf("🆔 ID: `%s`\n", user.ID.String())
	message += fmt.Sprintf("📱 Telegram ID: %d\n", user.TelegramUserID)
	message += fmt.Sprintf("👤 Name: %s %s\n", firstName, lastName)
	message += fmt.Sprintf("🔖 Username: @%s\n", username)
	message += fmt.Sprintf("🏷️ Role: %s\n", user.Role)
	message += fmt.Sprintf("📊 Status: %s\n", user.ApprovalStatus)
	message += fmt.Sprintf("📅 Joined: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))

	if user.ApprovedAt != nil {
		message += fmt.Sprintf("✅ Approved: %s\n", user.ApprovedAt.Format("2006-01-02 15:04:05"))
	}

	keyboard := model.InlineKeyboardMarkup{
		InlineKeyboard: [][]model.InlineKeyboardButton{},
	}

	// Add action buttons based on user status
	if user.ApprovalStatus == "approved" {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []model.InlineKeyboardButton{
			{Text: "🚫 Disable", CallbackData: fmt.Sprintf("disable_user:%s", user.ID.String())},
		})
	} else if user.ApprovalStatus == "disabled" {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []model.InlineKeyboardButton{
			{Text: "✅ Enable", CallbackData: fmt.Sprintf("enable_user:%s", user.ID.String())},
		})
	}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []model.InlineKeyboardButton{
		{Text: "🔙 Back", CallbackData: "admin_menu:approved"},
	})

	s.sendMessageWithKeyboard(callback.Message.Chat.ID, message, keyboard)
	s.answerCallbackQuery(callback.ID, "")
}

func (s *TelegramAdminService) sendMessageWithKeyboard(chatID int64, message string, keyboard model.InlineKeyboardMarkup) {
	err := s.telegramService.SendMessageWithKeyboard(chatID, message, keyboard)
	if err != nil {
		// Fallback to regular message if keyboard fails
		s.telegramService.SendMessage(chatID, message)
	}
}

func (s *TelegramAdminService) answerCallbackQuery(callbackID, text string) {
	err := s.telegramService.AnswerCallbackQuery(callbackID, text)
	if err != nil {
		slog.Error("Failed to answer callback query", "callbackID", callbackID, "error", err)
	}
}
