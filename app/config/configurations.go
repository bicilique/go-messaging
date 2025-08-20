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
}

func LoadConfigurations() *Configurations {

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	return &Configurations{
		PORT:               os.Getenv("PORT"),
		MODE:               os.Getenv("MODE"),
		TELEGRAM_BOT_TOKEN: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TELEGRAM_CHAT_ID:   os.Getenv("TELEGRAM_CHAT_ID"),
	}
}
