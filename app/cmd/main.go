package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-messaging/config"
	"go-messaging/database"
	httpDelivery "go-messaging/delivery/http"
	"go-messaging/internal/scheduler"
	"go-messaging/repository"
	"go-messaging/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfigurations()

	// Setup database
	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	repos := initializeRepositories(db)

	// Initialize services
	services := initializeServices(repos, cfg)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup HTTP server
	httpServer := setupHTTPServer(services, db)

	// Setup graceful shutdown
	setupGracefulShutdown(cancel)

	// Start notification scheduler
	go startNotificationScheduler(ctx, services.NotificationDispatch)

	// Start cleanup scheduler
	go startCleanupScheduler(ctx, services.Admin)

	// Start Telegram bot
	go func() {
		log.Printf("🚀 Starting Telegram Bot (Token: %s...)", cfg.TELEGRAM_BOT_TOKEN[:10])
		services.TelegramBot.CheckAdminServices() // Debug check
		services.TelegramBot.StartPolling(ctx)
	}()

	// Start HTTP server
	startHTTPServer(ctx, httpServer)

	log.Println("👋 Application stopped")
}

// setupDatabase initializes and migrates the database
func setupDatabase(cfg *config.Configurations) (*database.Database, error) {
	dbConfig := database.Config{
		Host:     cfg.DB_HOST,
		Port:     cfg.DB_PORT,
		User:     cfg.DB_USER,
		Password: cfg.DB_PASSWORD,
		DBName:   cfg.DB_NAME,
		SSLMode:  cfg.DB_SSLMODE,
	}

	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		return nil, err
	}
	log.Println("✅ Database connected successfully")

	if err := db.AutoMigrate(); err != nil {
		return nil, err
	}
	log.Println("✅ Database migrations completed")

	// Disable seeding for now
	// if err := db.Seed(); err != nil {
	// 	return nil, err
	// }
	// log.Println("✅ Database seeded with default notification types")
	return db, nil
}

// Repositories holds all repository instances
type Repositories struct {
	User             repository.UserRepository
	NotificationType repository.NotificationTypeRepository
	Subscription     repository.SubscriptionRepository
	NotificationLog  repository.NotificationLogRepository
}

// initializeRepositories creates all repository instances
func initializeRepositories(db *database.Database) *Repositories {
	return &Repositories{
		User:             repository.NewUserRepository(db.Connection),
		NotificationType: repository.NewNotificationTypeRepository(db.Connection),
		Subscription:     repository.NewSubscriptionRepository(db.Connection),
		NotificationLog:  repository.NewNotificationLogRepository(db.Connection),
	}
}

// Services holds all service instances
type Services struct {
	User                 service.UserService
	NotificationType     service.NotificationTypeService
	Subscription         service.SubscriptionService
	NotificationLog      service.NotificationLogService
	Admin                service.AdminServiceInterface
	TelegramBot          *service.TelegramBotService
	NotificationDispatch service.NotificationDispatchService
	Detection            service.DetectionInterface
}

// initializeServices creates all service instances
func initializeServices(repos *Repositories, cfg *config.Configurations) *Services {
	userService := service.NewUserService(repos.User)
	notificationTypeService := service.NewNotificationTypeService(repos.NotificationType)
	subscriptionService := service.NewSubscriptionService(
		repos.Subscription,
		repos.User,
		repos.NotificationType,
		repos.NotificationLog,
	)
	notificationLogService := service.NewNotificationLogService(repos.NotificationLog)

	// Create admin service
	adminService := service.NewAdminService(repos.User)

	// Create the main Telegram bot service
	telegramBotService := service.NewTelegramBotService(
		cfg.TELEGRAM_BOT_TOKEN,
		userService,
		subscriptionService,
		notificationTypeService,
		adminService,
	)
	// Create notification dispatch service
	notificationDispatchService := service.NewNotificationDispatchService(
		subscriptionService,
		notificationLogService,
		telegramBotService,
	)

	// Create detection service
	detectionService := service.NewDetectionService(
		notificationDispatchService,
		repos.Subscription,
		repos.NotificationType,
	)

	return &Services{
		User:                 userService,
		NotificationType:     notificationTypeService,
		Subscription:         subscriptionService,
		NotificationLog:      notificationLogService,
		Admin:                adminService,
		TelegramBot:          telegramBotService,
		NotificationDispatch: notificationDispatchService,
		Detection:            detectionService,
	}
}

// setupHTTPServer creates and configures the HTTP server
func setupHTTPServer(services *Services, db *database.Database) *http.Server {
	router := gin.Default()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Initialize handlers
	userHandler := httpDelivery.NewUserHandler(services.User)
	adminHandler := httpDelivery.NewAdminHandler(services.Admin)
	authMiddleware := httpDelivery.NewBasicAuthMiddleware(db.Connection)
	detectionHandler := httpDelivery.NewDetectionHandler(services.Detection)

	// Setup routes
	routeConfig := &httpDelivery.RouteConfig{
		Router:           router,
		UserHandler:      userHandler,
		AdminHandler:     adminHandler,
		AuthMiddleware:   authMiddleware,
		DetectionHandler: detectionHandler,
	}
	routeConfig.Setup()

	return &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
}

// startHTTPServer starts the HTTP server with graceful shutdown
func startHTTPServer(ctx context.Context, server *http.Server) {
	go func() {
		log.Printf("🌐 Starting HTTP server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("🛑 Shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ HTTP server forced to shutdown: %v", err)
	} else {
		log.Println("✅ HTTP server stopped gracefully")
	}
}

// startNotificationScheduler starts the notification scheduling service
func startNotificationScheduler(ctx context.Context, dispatchService service.NotificationDispatchService) {
	notificationScheduler := scheduler.NewNotificationScheduler(dispatchService)
	notificationScheduler.Start(ctx)
}

// startCleanupScheduler starts the cleanup scheduling service
func startCleanupScheduler(ctx context.Context, adminService service.AdminServiceInterface) {
	cleanupScheduler := scheduler.NewCleanupScheduler(adminService)
	cleanupScheduler.Start()

	// Stop scheduler when context is cancelled
	go func() {
		<-ctx.Done()
		cleanupScheduler.Stop()
	}()
}

// setupGracefulShutdown sets up signal handling for graceful shutdown
func setupGracefulShutdown(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("🛑 Received shutdown signal, shutting down gracefully...")
		cancel()
	}()
}
