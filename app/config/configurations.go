package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Configurations struct {
	PORT               string
	MODE               string
	TELEGRAM_BOT_TOKEN string
	TELEGRAM_CHAT_ID   string

	// Database configuration
	DB_HOST     string
	DB_PORT     string
	DB_USER     string
	DB_PASSWORD string
	DB_NAME     string
	DB_SSLMODE  string
}

func LoadConfigurations() *Configurations {

	if os.Getenv("DEVELOPER_HOST") == "true" {
		err := godotenv.Load()
		if err != nil {
			panic("Error loading .env file")
		}

	}
	return &Configurations{
		PORT:               os.Getenv("PORT"),
		MODE:               os.Getenv("MODE"),
		TELEGRAM_BOT_TOKEN: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TELEGRAM_CHAT_ID:   os.Getenv("TELEGRAM_CHAT_ID"),

		// Database configuration
		DB_HOST:     getEnvWithDefault("DB_HOST", "localhost"),
		DB_PORT:     getEnvWithDefault("DB_PORT", "5432"),
		DB_USER:     getEnvWithDefault("DB_USER", "postgres"),
		DB_PASSWORD: getEnvWithDefault("DB_PASSWORD", ""),
		DB_NAME:     getEnvWithDefault("DB_NAME", "go_messaging"),
		DB_SSLMODE:  getEnvWithDefault("DB_SSLMODE", "disable"),
	}
}

// Helper function to get environment variable with default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
