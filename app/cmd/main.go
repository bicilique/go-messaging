package main

import (
	"context"
	"go-messaging/config"
	"go-messaging/delivery/http"
	"go-messaging/service"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configurations
	config := config.LoadConfigurations()

	// Initialize Telegram service
	telegramService := service.NewTelegramService(config.TELEGRAM_BOT_TOKEN)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Telegram bot polling
	telegramService.StartPolling(ctx)

	// Initialize HTTP handler
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

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("Received shutdown signal, stopping services...")
		cancel() // This will stop the Telegram polling
	}()

	// Start server
	slog.Info("Starting server on port %s in %s mode...", port, gin.Mode())
	slog.Info("Telegram bot is now listening for messages...")

	err := router.Router.Run(":" + port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
