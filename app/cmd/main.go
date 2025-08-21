package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-messaging/config"
	"go-messaging/database"
	"go-messaging/repository"
	"go-messaging/service"
)

func main() {
	// Load configuration
	cfg := config.LoadConfigurations()

	// Setup database connection
	dbConfig := database.Config{
		Host:     cfg.DB_HOST,
		Port:     cfg.DB_PORT,
		User:     cfg.DB_USER,
		Password: cfg.DB_PASSWORD,
		DBName:   cfg.DB_NAME,
		SSLMode:  cfg.DB_SSLMODE,
	}

	// Initialize database connection
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connected successfully")

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("✅ Database migrations completed")

	// Seed default data
	if err := db.Seed(); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}
	log.Println("✅ Database seeded with default notification types")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.Connection)
	notificationTypeRepo := repository.NewNotificationTypeRepository(db.Connection)
	subscriptionRepo := repository.NewSubscriptionRepository(db.Connection)
	notificationLogRepo := repository.NewNotificationLogRepository(db.Connection)

	// Initialize services
	userService := service.NewUserService(userRepo)
	notificationTypeService := service.NewNotificationTypeService(notificationTypeRepo)
	subscriptionService := service.NewSubscriptionService(
		subscriptionRepo,
		userRepo,
		notificationTypeRepo,
		notificationLogRepo,
	)
	notificationLogService := service.NewNotificationLogService(notificationLogRepo)

	// Initialize Telegram bot service
	telegramService := service.NewTelegramBotService(
		cfg.TELEGRAM_BOT_TOKEN,
		userService,
		subscriptionService,
		notificationTypeService,
	)

	// Initialize notification dispatch service
	notificationDispatchService := service.NewNotificationDispatchService(
		subscriptionService,
		notificationLogService,
		telegramService, // Telegram service implements TelegramNotificationSender interface
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("🛑 Received shutdown signal, shutting down gracefully...")
		cancel()
	}()

	// Start the bot
	log.Printf("🚀 Starting Telegram Bot (Token: %s...)", cfg.TELEGRAM_BOT_TOKEN[:10])

	// Start notification scheduler in a separate goroutine
	go startNotificationScheduler(ctx, notificationDispatchService)

	// Start the Telegram bot (this will block until context is cancelled)
	telegramService.StartPolling(ctx)

	log.Println("👋 Application stopped")
}

// startNotificationScheduler runs periodic notification dispatching
func startNotificationScheduler(ctx context.Context, dispatchService service.NotificationDispatchService) {
	log.Println("📡 Starting notification scheduler...")

	// Notification types and their dispatch intervals (in minutes)
	notificationSchedule := map[string]int{
		"coinbase":    1, // Every 1 minute
		"news":        2, // Every 2 minutes
		"weather":     4, // Every 4 minutes
		"price_alert": 5, // Every 5 minute for testing (change back to 5+ for production)
		"custom":      6, // Every 6 minutes
	}

	// Start individual schedulers for each notification type
	for notificationType, intervalMinutes := range notificationSchedule {
		go runNotificationSchedule(ctx, dispatchService, notificationType, intervalMinutes)
	}

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("📡 Notification scheduler stopped")
}

// runNotificationSchedule runs a scheduler for a specific notification type
func runNotificationSchedule(ctx context.Context, dispatchService service.NotificationDispatchService, notificationType string, intervalMinutes int) {
	log.Printf("⏰ Starting %s notification scheduler (every %d minutes)", notificationType, intervalMinutes)

	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	// For development: Don't run immediately on startup, wait for first interval
	log.Printf("⏳ Waiting %d minutes before first %s notification...", intervalMinutes, notificationType)

	for {
		select {
		case <-ctx.Done():
			log.Printf("⏰ %s notification scheduler stopped", notificationType)
			return
		case <-ticker.C:
			log.Printf("🔔 Time to dispatch %s notifications!", notificationType)
			if err := dispatchService.DispatchNotification(ctx, notificationType); err != nil {
				log.Printf("❌ Failed to dispatch %s notifications: %v", notificationType, err)
			} else {
				log.Printf("✅ Dispatched %s notifications", notificationType)
			}
		}
	}
}
