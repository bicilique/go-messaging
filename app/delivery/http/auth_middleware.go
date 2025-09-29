package http

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type APICredential struct {
	ID           uuid.UUID `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Role         string    `json:"role" gorm:"default:'admin'"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
}

func (APICredential) TableName() string {
	return "api_credentials"
}

type BasicAuthMiddleware struct {
	db *gorm.DB
}

func NewBasicAuthMiddleware(db *gorm.DB) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{db: db}
}

// BasicAuth provides HTTP Basic Authentication middleware
func (m *BasicAuthMiddleware) BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			m.requireAuth(c)
			return
		}

		if !strings.HasPrefix(auth, "Basic ") {
			m.requireAuth(c)
			return
		}

		// Extract credentials
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			m.requireAuth(c)
			return
		}

		// Validate credentials against database
		if !m.validateCredentials(username, password) {
			m.requireAuth(c)
			return
		}

		// Set username in context for logging
		c.Set("auth_username", username)
		c.Next()
	}
}

// AdminAuth provides admin-specific authentication
func (m *BasicAuthMiddleware) AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			m.requireAuth(c)
			return
		}

		if !strings.HasPrefix(auth, "Basic ") {
			m.requireAuth(c)
			return
		}

		username, password, ok := c.Request.BasicAuth()
		if !ok {
			m.requireAuth(c)
			return
		}

		// Validate credentials and check admin role
		if !m.validateAdminCredentials(username, password) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Set("auth_username", username)
		c.Set("auth_role", "admin")
		c.Next()
	}
}

func (m *BasicAuthMiddleware) requireAuth(c *gin.Context) {
	c.Header("WWW-Authenticate", `Basic realm="Go Messaging API"`)
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "Authentication required",
	})
	c.Abort()
}

// validateCredentials checks the provided username and password against the database
func (m *BasicAuthMiddleware) validateCredentials(username, password string) bool {
	var cred APICredential
	err := m.db.Where("username = ? AND is_active = ?", username, true).First(&cred).Error
	if err != nil {
		return false
	}

	// Compare password hash
	err = bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(password))
	if err != nil {
		fmt.Printf("[DEBUG] bcrypt error: %v\n", err)
	}
	return err == nil
}

// validateAdminCredentials checks the provided username and password and ensures the user has admin role
func (m *BasicAuthMiddleware) validateAdminCredentials(username, password string) bool {
	var cred APICredential

	err := m.db.Where("username = ? AND is_active = ? AND role = ?", username, true, "admin").First(&cred).Error
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(password))
	if err != nil {
		fmt.Printf("[DEBUG] bcrypt error: %v\n", err)
	}
	return err == nil
}

// Simple auth for development/testing (not recommended for production)
func SimpleBasicAuth(username, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		username: password,
	})
}

// SecureCompare performs a constant-time comparison of two strings
func SecureCompare(given, actual string) bool {
	return subtle.ConstantTimeCompare([]byte(given), []byte(actual)) == 1
}
