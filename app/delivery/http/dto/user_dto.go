package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	TelegramUserID int64   `json:"telegram_user_id" binding:"required"`
	Username       *string `json:"username,omitempty"`
	FirstName      *string `json:"first_name,omitempty"`
	LastName       *string `json:"last_name,omitempty"`
	LanguageCode   *string `json:"language_code,omitempty"`
	IsBot          bool    `json:"is_bot"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Username     *string `json:"username,omitempty"`
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	LanguageCode *string `json:"language_code,omitempty"`
	IsBot        *bool   `json:"is_bot,omitempty"`
}

// UserResponse represents the response for user operations
type UserResponse struct {
	ID             uuid.UUID `json:"id"`
	TelegramUserID int64     `json:"telegram_user_id"`
	Username       *string   `json:"username,omitempty"`
	FirstName      *string   `json:"first_name,omitempty"`
	LastName       *string   `json:"last_name,omitempty"`
	LanguageCode   *string   `json:"language_code,omitempty"`
	IsBot          bool      `json:"is_bot"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PaginationQuery represents common pagination query parameters
type PaginationQuery struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=10" binding:"min=1,max=100"`
}

// PaginatedUsersResponse represents paginated user response
type PaginatedUsersResponse struct {
	Users []UserResponse `json:"users"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
	Total int64          `json:"total"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
