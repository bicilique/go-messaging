package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpDelivery "go-messaging/delivery/http"
	"go-messaging/delivery/http/dto"
	"go-messaging/entity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateOrUpdateUser(ctx context.Context, telegramUserID int64, username, firstName, lastName, languageCode *string, isBot bool) (*entity.User, error) {
	args := m.Called(ctx, telegramUserID, username, firstName, lastName, languageCode, isBot)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserService) GetUserByTelegramID(ctx context.Context, telegramUserID int64) (*entity.User, error) {
	args := m.Called(ctx, telegramUserID)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, telegramUserID int64) error {
	args := m.Called(ctx, telegramUserID)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, offset, limit int) ([]*entity.User, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*entity.User), args.Error(1)
}

func TestUserHandler_CreateUser(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUserService := new(MockUserService)
	handler := httpDelivery.NewUserHandler(mockUserService)

	// Create test user
	userID := uuid.New()
	telegramUserID := int64(12345)
	username := "testuser"
	firstName := "Test"
	lastName := "User"
	languageCode := "en"

	expectedUser := &entity.User{
		ID:             userID,
		TelegramUserID: telegramUserID,
		Username:       &username,
		FirstName:      &firstName,
		LastName:       &lastName,
		LanguageCode:   &languageCode,
		IsBot:          false,
	}

	// Mock service call
	mockUserService.On("CreateOrUpdateUser", mock.Anything, telegramUserID, &username, &firstName, &lastName, &languageCode, false).Return(expectedUser, nil)

	// Create request
	requestBody := dto.CreateUserRequest{
		TelegramUserID: telegramUserID,
		Username:       &username,
		FirstName:      &firstName,
		LastName:       &lastName,
		LanguageCode:   &languageCode,
		IsBot:          false,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/v1/users", handler.CreateUser)

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response dto.UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, userID, response.ID)
	assert.Equal(t, telegramUserID, response.TelegramUserID)
	assert.Equal(t, username, *response.Username)

	mockUserService.AssertExpectations(t)
}

func TestUserHandler_GetUser(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUserService := new(MockUserService)
	handler := httpDelivery.NewUserHandler(mockUserService)

	// Create test user
	userID := uuid.New()
	telegramUserID := int64(12345)
	username := "testuser"

	expectedUser := &entity.User{
		ID:             userID,
		TelegramUserID: telegramUserID,
		Username:       &username,
		IsBot:          false,
	}

	// Mock service call
	mockUserService.On("GetUserByID", mock.Anything, userID).Return(expectedUser, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/users/"+userID.String(), nil)

	// Create response recorder
	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/api/v1/users/:id", handler.GetUser)

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, userID, response.ID)
	assert.Equal(t, telegramUserID, response.TelegramUserID)
	assert.Equal(t, username, *response.Username)

	mockUserService.AssertExpectations(t)
}
