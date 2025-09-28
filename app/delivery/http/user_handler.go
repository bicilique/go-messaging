package http

import (
	"net/http"
	"strconv"
	"time"

	"go-messaging/delivery/http/dto"
	"go-messaging/entity"
	"go-messaging/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser creates a new user or updates existing one
// @Summary Create or update user
// @Description Create a new user or update existing user based on telegram_user_id
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User data"
// @Success 201 {object} dto.UserResponse
// @Success 200 {object} dto.UserResponse "User updated"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}

	user, err := h.userService.CreateOrUpdateUser(
		c.Request.Context(),
		req.TelegramUserID,
		req.Username,
		req.FirstName,
		req.LastName,
		req.LanguageCode,
		req.IsBot,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create user",
			Message: err.Error(),
		})
		return
	}

	response := h.entityToResponse(user)

	// Check if user was created or updated by checking if CreatedAt and UpdatedAt are close
	statusCode := http.StatusCreated
	if user.UpdatedAt.Sub(user.CreatedAt) > time.Second {
		statusCode = http.StatusOK
	}

	c.JSON(statusCode, response)
}

// GetUser retrieves a user by ID
// @Summary Get user by ID
// @Description Get a single user by their UUID
// @Tags users
// @Produce json
// @Param id path string true "User UUID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user ID format",
			Message: "User ID must be a valid UUID",
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get user",
			Message: err.Error(),
		})
		return
	}

	response := h.entityToResponse(user)
	c.JSON(http.StatusOK, response)
}

// GetUserByTelegramID retrieves a user by Telegram user ID
// @Summary Get user by Telegram ID
// @Description Get a user by their Telegram user ID
// @Tags users
// @Produce json
// @Param telegram_user_id path int true "Telegram User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/telegram/{telegram_user_id} [get]
func (h *UserHandler) GetUserByTelegramID(c *gin.Context) {
	telegramUserIDParam := c.Param("telegram_user_id")
	telegramUserID, err := strconv.ParseInt(telegramUserIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Telegram user ID",
			Message: "Telegram user ID must be a valid integer",
		})
		return
	}

	user, err := h.userService.GetUserByTelegramID(c.Request.Context(), telegramUserID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the given Telegram user ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get user",
			Message: err.Error(),
		})
		return
	}

	response := h.entityToResponse(user)
	c.JSON(http.StatusOK, response)
}

// UpdateUser updates an existing user
// @Summary Update user
// @Description Update an existing user's information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param user body dto.UpdateUserRequest true "Updated user data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user ID format",
			Message: "User ID must be a valid UUID",
		})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}

	// Get existing user
	user, err := h.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get user",
			Message: err.Error(),
		})
		return
	}

	// Update fields if provided
	if req.Username != nil {
		user.Username = req.Username
	}
	if req.FirstName != nil {
		user.FirstName = req.FirstName
	}
	if req.LastName != nil {
		user.LastName = req.LastName
	}
	if req.LanguageCode != nil {
		user.LanguageCode = req.LanguageCode
	}
	if req.IsBot != nil {
		user.IsBot = *req.IsBot
	}

	if err := h.userService.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to update user",
			Message: err.Error(),
		})
		return
	}

	response := h.entityToResponse(user)
	c.JSON(http.StatusOK, response)
}

// DeleteUser deletes a user by Telegram user ID
// @Summary Delete user
// @Description Delete a user and all related data by Telegram user ID
// @Tags users
// @Param telegram_user_id path int true "Telegram User ID"
// @Success 204 "User deleted successfully"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/telegram/{telegram_user_id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	telegramUserIDParam := c.Param("telegram_user_id")
	telegramUserID, err := strconv.ParseInt(telegramUserIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Telegram user ID",
			Message: "Telegram user ID must be a valid integer",
		})
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), telegramUserID); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the given Telegram user ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to delete user",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListUsers retrieves users with pagination
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} dto.PaginatedUsersResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid query parameters",
			Message: err.Error(),
		})
		return
	}

	offset := (query.Page - 1) * query.Limit
	users, err := h.userService.ListUsers(c.Request.Context(), offset, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list users",
			Message: err.Error(),
		})
		return
	}

	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = h.entityToResponse(user)
	}

	response := dto.PaginatedUsersResponse{
		Users: userResponses,
		Page:  query.Page,
		Limit: query.Limit,
		Total: int64(len(users)), // Note: This should be updated to get actual total count
	}

	c.JSON(http.StatusOK, response)
}

// entityToResponse converts entity.User to dto.UserResponse
func (h *UserHandler) entityToResponse(user *entity.User) dto.UserResponse {
	return dto.UserResponse{
		ID:             user.ID,
		TelegramUserID: user.TelegramUserID,
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		LanguageCode:   user.LanguageCode,
		IsBot:          user.IsBot,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}
