package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"go-messaging/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	Connection *gorm.DB
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDatabase creates a new database connection
func NewDatabase(config Config) (*Database, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	// Configure logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return &Database{Connection: db}, nil
}

// NewDatabaseFromEnv creates a database connection from environment variables
func NewDatabaseFromEnv() (*Database, error) {
	config := Config{
		Host:     getEnvWithDefault("DB_HOST", "localhost"),
		Port:     getEnvWithDefault("DB_PORT", "5432"),
		User:     getEnvWithDefault("DB_USER", "postgres"),
		Password: getEnvWithDefault("DB_PASSWORD", ""),
		DBName:   getEnvWithDefault("DB_NAME", "go_messaging"),
		SSLMode:  getEnvWithDefault("DB_SSLMODE", "disable"),
	}

	return NewDatabase(config)
}

// AutoMigrate runs database migrations
func (d *Database) AutoMigrate() error {
	// Handle constraint conflicts gracefully
	// if err := d.handleConstraintConflicts(); err != nil {
	// 	log.Printf("Warning: Failed to handle constraint conflicts: %v", err)
	// }

	return d.Connection.AutoMigrate(
		&entity.User{},
		&entity.NotificationType{},
		&entity.Subscription{},
		&entity.NotificationLog{},
	)
}

// handleConstraintConflicts handles known constraint conflicts between SQL schema and GORM
func (d *Database) handleConstraintConflicts() error {
	// List of constraints that GORM might try to drop but may not exist
	constraintsToCheck := []struct {
		table      string
		constraint string
	}{
		{"users", "uni_users_telegram_user_id"},
		{"notification_types", "uni_notification_types_code"},
	}

	for _, item := range constraintsToCheck {
		var count int64
		query := `
			SELECT COUNT(*) 
			FROM information_schema.table_constraints 
			WHERE table_name = $1 
			AND constraint_name = $2
			AND table_schema = CURRENT_SCHEMA()
		`

		if err := d.Connection.Raw(query, item.table, item.constraint).Scan(&count).Error; err != nil {
			log.Printf("Warning: Failed to check constraint %s on table %s: %v", item.constraint, item.table, err)
			continue
		}

		// If constraint exists, drop it safely
		if count > 0 {
			dropSQL := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", item.table, item.constraint)
			if err := d.Connection.Exec(dropSQL).Error; err != nil {
				log.Printf("Warning: Failed to drop constraint %s on table %s: %v", item.constraint, item.table, err)
				continue
			}
			log.Printf("✅ Dropped existing constraint %s on table %s", item.constraint, item.table)
		}
	}

	return nil
}

// Seed inserts default notification types
func (d *Database) Seed() error {
	notificationTypes := []entity.NotificationType{
		{
			Code:                   "coinbase",
			Name:                   "Coinbase Alerts",
			Description:            stringPtr("Cryptocurrency price updates and market alerts"),
			DefaultIntervalMinutes: 1,
			IsActive:               true,
		},
		{
			Code:                   "news",
			Name:                   "News Alerts",
			Description:            stringPtr("Breaking news and important updates"),
			DefaultIntervalMinutes: 2,
			IsActive:               true,
		},
		{
			Code:                   "weather",
			Name:                   "Weather Updates",
			Description:            stringPtr("Weather forecasts and alerts"),
			DefaultIntervalMinutes: 4,
			IsActive:               true,
		},
		{
			Code:                   "price_alert",
			Name:                   "Price Alerts",
			Description:            stringPtr("Custom price threshold notifications"),
			DefaultIntervalMinutes: 5,
			IsActive:               true,
		},
		{
			Code:                   "custom",
			Name:                   "Custom Notifications",
			Description:            stringPtr("Custom notifications for specific needs"),
			DefaultIntervalMinutes: 6,
			IsActive:               true,
		},
	}

	for _, nt := range notificationTypes {
		var existing entity.NotificationType
		result := d.Connection.Where("code = ?", nt.Code).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := d.Connection.Create(&nt).Error; err != nil {
				return fmt.Errorf("failed to seed notification type %s: %w", nt.Code, err)
			}
		}
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.Connection.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping tests the database connection
func (d *Database) Ping() error {
	sqlDB, err := d.Connection.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Helper functions
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func stringPtr(s string) *string {
	return &s
}
