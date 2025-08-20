package main

import (
	"go-messaging/config"
	"go-messaging/delivery/http"
	"go-messaging/service"
	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configurations
	config := config.LoadConfigurations()

	// Initialize Telegram service
	telegramService := service.NewTelegramService(config.TELEGRAM_BOT_TOKEN)

	// Example usage: Send a message
	irisHandler := http.NewIrisHandler(*telegramService)

	// Set Gin mode based on configuration
	mode := config.MODE
	if mode == "" {
		mode = gin.ReleaseMode
	} else {
		gin.SetMode(gin.DebugMode)
	}

	port := config.PORT
	if port == "" {
		port = "8080"
	}

	// Set up the Gin router
	router := http.RouteConfig{
		Router:      gin.Default(),
		IrisHandler: irisHandler,
	}
	router.Setup()

	// Start server
	slog.Info("Starting server on port %s in %s mode...", port, gin.Mode())
	err := router.Router.Run(":" + port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
