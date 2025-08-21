package model

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Rate limiting configuration constants
const (
	RATE_LIMIT_MESSAGES  = 10              // 10 messages
	RATE_LIMIT_WINDOW    = time.Minute     // per minute
	MIN_MESSAGE_INTERVAL = 1 * time.Second // 1 second between messages
)

// RateLimiter manages rate limiting for users
type RateLimiter struct {
	users map[int64]*UserLimiter
	mutex sync.RWMutex
}

// UserLimiter keeps track of message counts and timestamps for each user
type UserLimiter struct {
	LastMessage  time.Time
	MessageCount int
	WindowStart  time.Time
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		users: make(map[int64]*UserLimiter),
	}

	// Start cleanup routine
	go rl.cleanup()

	return rl
}

// IsAllowed checks if a user is allowed to send a message
func (rl *RateLimiter) IsAllowed(userID int64) (bool, string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Get or create user limiter
	user, exists := rl.users[userID]
	if !exists {
		user = &UserLimiter{
			LastMessage:  now,
			MessageCount: 1,
			WindowStart:  now,
		}
		rl.users[userID] = user
		return true, ""
	}

	// Check minimum interval between messages
	if now.Sub(user.LastMessage) < MIN_MESSAGE_INTERVAL {
		return false, "⏱️ Please wait 1 second between messages"
	}

	// Reset window if needed
	if now.Sub(user.WindowStart) > RATE_LIMIT_WINDOW {
		user.MessageCount = 0
		user.WindowStart = now
	}

	// Check rate limit
	if user.MessageCount >= RATE_LIMIT_MESSAGES {
		remaining := RATE_LIMIT_WINDOW - now.Sub(user.WindowStart)
		return false, fmt.Sprintf("🚫 Rate limit exceeded! Try again in %v", remaining.Round(time.Second))
	}

	// Update counters
	user.LastMessage = now
	user.MessageCount++

	return true, ""
}

// Enhanced IsAllowed with timeout protection
func (rl *RateLimiter) IsAllowedWithContext(ctx context.Context, userID int64) (bool, string) {
	// Try to acquire lock with context timeout
	acquired := make(chan bool, 1)
	go func() {
		rl.mutex.Lock()
		acquired <- true
	}()

	select {
	case <-ctx.Done():
		return false, "⏱️ Rate limit check timeout"
	case <-acquired:
		defer rl.mutex.Unlock()
		return rl.isAllowedInternal(userID)
	}
}

// Internal method (existing logic)
func (rl *RateLimiter) isAllowedInternal(userID int64) (bool, string) {
	now := time.Now()

	// Get or create user limiter
	user, exists := rl.users[userID]
	if !exists {
		user = &UserLimiter{
			LastMessage:  now,
			MessageCount: 1,
			WindowStart:  now,
		}
		rl.users[userID] = user
		return true, ""
	}

	// Check minimum interval between messages
	if now.Sub(user.LastMessage) < MIN_MESSAGE_INTERVAL {
		return false, "⏱️ Please wait 1 second between messages"
	}

	// Reset window if needed
	if now.Sub(user.WindowStart) > RATE_LIMIT_WINDOW {
		user.MessageCount = 0
		user.WindowStart = now
	}

	// Check rate limit
	if user.MessageCount >= RATE_LIMIT_MESSAGES {
		remaining := RATE_LIMIT_WINDOW - now.Sub(user.WindowStart)
		return false, fmt.Sprintf("🚫 Rate limit exceeded! Try again in %v", remaining.Round(time.Second))
	}

	// Update counters
	user.LastMessage = now
	user.MessageCount++

	return true, ""
}

// cleanup removes inactive users from memory
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()

		for userID, user := range rl.users {
			// Remove users inactive for 1 hour
			if now.Sub(user.LastMessage) > time.Hour {
				delete(rl.users, userID)
			}
		}

		rl.mutex.Unlock()
	}
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	return map[string]interface{}{
		"active_users":    len(rl.users),
		"messages_limit":  RATE_LIMIT_MESSAGES,
		"window_duration": RATE_LIMIT_WINDOW,
		"min_interval":    MIN_MESSAGE_INTERVAL,
	}
}
