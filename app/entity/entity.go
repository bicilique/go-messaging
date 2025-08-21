package entity

import (
	"encoding/json"
	"time"
)

type User struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	TelegramUserID int64     `json:"telegram_user_id" gorm:"uniqueIndex;not null"`
	Username       *string   `json:"username"`
	FirstName      *string   `json:"first_name"`
	LastName       *string   `json:"last_name"`
	LanguageCode   *string   `json:"language_code"`
	IsBot          bool      `json:"is_bot" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	Subscriptions []Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:UserID"`
}

type NotificationType struct {
	ID                     int       `json:"id" gorm:"primaryKey"`
	Code                   string    `json:"code" gorm:"uniqueIndex;not null"`
	Name                   string    `json:"name" gorm:"not null"`
	Description            *string   `json:"description"`
	DefaultIntervalMinutes int       `json:"default_interval_minutes" gorm:"default:60"`
	IsActive               bool      `json:"is_active" gorm:"default:true"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`

	// Relationships
	Subscriptions []Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:NotificationTypeID"`
}

type SubscriptionPreferences struct {
	Currency  string            `json:"currency,omitempty"`
	Interval  int               `json:"interval,omitempty"` // minutes
	Keywords  []string          `json:"keywords,omitempty"`
	Threshold float64           `json:"threshold,omitempty"`
	Settings  map[string]string `json:"settings,omitempty"`
}

type Subscription struct {
	ID                 int64                   `json:"id" gorm:"primaryKey"`
	UserID             int64                   `json:"user_id" gorm:"not null;index"`
	ChatID             int64                   `json:"chat_id" gorm:"not null;index"`
	NotificationTypeID int                     `json:"notification_type_id" gorm:"not null;index"`
	IsActive           bool                    `json:"is_active" gorm:"default:true;index"`
	Preferences        SubscriptionPreferences `json:"preferences" gorm:"type:jsonb"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
	LastNotifiedAt     *time.Time              `json:"last_notified_at"`

	// Relationships
	User             User              `json:"user,omitempty" gorm:"foreignKey:UserID"`
	NotificationType NotificationType  `json:"notification_type,omitempty" gorm:"foreignKey:NotificationTypeID"`
	NotificationLogs []NotificationLog `json:"notification_logs,omitempty" gorm:"foreignKey:SubscriptionID"`
}

type NotificationLog struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	SubscriptionID int64     `json:"subscription_id" gorm:"not null;index"`
	Message        string    `json:"message" gorm:"not null"`
	Status         string    `json:"status" gorm:"default:'sent'"` // sent, failed, delivered
	SentAt         time.Time `json:"sent_at"`
	ErrorMessage   *string   `json:"error_message"`

	// Relationships
	Subscription Subscription `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
}

// Scan implements the sql.Scanner interface for JSONB
func (sp *SubscriptionPreferences) Scan(value interface{}) error {
	if value == nil {
		*sp = SubscriptionPreferences{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, sp)
}

// Value implements the driver.Valuer interface for JSONB
func (sp SubscriptionPreferences) Value() (interface{}, error) {
	// Check if struct is empty by comparing individual fields
	if sp.Currency == "" && sp.Interval == 0 && len(sp.Keywords) == 0 &&
		sp.Threshold == 0 && len(sp.Settings) == 0 {
		return "{}", nil
	}
	return json.Marshal(sp)
}

// TableName methods for GORM
func (User) TableName() string             { return "users" }
func (NotificationType) TableName() string { return "notification_types" }
func (Subscription) TableName() string     { return "subscriptions" }
func (NotificationLog) TableName() string  { return "notification_logs" }
