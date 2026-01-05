package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/config"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/handlers"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/middleware"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/repository"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/service"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/services"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/logger"
)

var (
	configPath = flag.String("config", "config/api_server_config.json", "Path to configuration file")
	version    = "2.0.0"
	buildTime  = "unknown"
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.OutputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Infof("Starting REST API Server v%s (built: %s)", version, buildTime)
	logger.Infof("Configuration loaded from: %s", *configPath)

	// Initialize services
	app := initializeApp(cfg)
	defer app.cleanup()

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create Gin router with middleware
	router := setupRouter(cfg, app)

	// Create HTTP server
	srv := &http.Server{
		Addr:           cfg.GetServerAddress(),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("HTTP server listening on %s", cfg.GetServerAddress())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.Server.ShutdownTimeout)*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited successfully")
}

// App holds all application dependencies
type App struct {
	rabbitMQ       *services.RabbitMQService
	redis          *services.RedisService
	db             *services.DatabaseService
	userService    *service.UserService
	messageService *service.MessageService
}

// cleanup closes all services
func (a *App) cleanup() {
	if a.rabbitMQ != nil {
		a.rabbitMQ.Close()
	}
	if a.redis != nil {
		a.redis.Close()
	}
	if a.db != nil {
		a.db.Close()
	}
}

// initializeApp initializes all services and dependencies
func initializeApp(cfg *config.Config) *App {
	app := &App{}

	// Initialize RabbitMQ
	rabbitMQ, err := services.NewRabbitMQService(&cfg.RabbitMQ)
	if err != nil {
		logger.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	app.rabbitMQ = rabbitMQ

	// Initialize Redis (optional)
	var redisService *services.RedisService
	if cfg.Redis.Enabled {
		redis, err := services.NewRedisService(&cfg.Redis)
		if err != nil {
			logger.Warnf("Failed to initialize Redis: %v. Continuing without Redis.", err)
		} else {
			redisService = redis
			app.redis = redis
		}
	}

	// Initialize Database (optional)
	var dbService *services.DatabaseService
	if cfg.Database.Enabled {
		db, err := services.NewDatabaseService(&cfg.Database)
		if err != nil {
			logger.Warnf("Failed to initialize database: %v. Continuing without database.", err)
		} else {
			dbService = db
			app.db = db

			// Initialize schema
			if err := db.InitSchema(); err != nil {
				logger.Warnf("Failed to initialize database schema: %v", err)
			}
		}
	}

	// Initialize repositories
	var userRepo repository.UserRepository
	var messageRepo repository.MessageRepository

	if dbService != nil {
		userRepo = repository.NewUserRepository(dbService.GetDB())
		messageRepo = repository.NewMessageRepository(dbService.GetDB())
	}

	// Initialize services
	if userRepo != nil {
		app.userService = service.NewUserService(userRepo, redisService)
	}

	if messageRepo != nil {
		app.messageService = service.NewMessageService(messageRepo, userRepo, rabbitMQ, redisService)
	}

	return app
}

// setupRouter configures all routes and middleware
func setupRouter(cfg *config.Config, app *App) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	systemHandler := handlers.NewSystemHandler(app.rabbitMQ, app.redis, app.db, version)

	// Metrics middleware (if enabled)
	if cfg.Metrics.Enabled {
		router.Use(middleware.PrometheusMetrics())
		router.GET(cfg.Metrics.Path, handlers.MetricsHandler())
	}

	// Rate limiting middleware
	router.Use(middleware.RateLimitByIP(10, 20)) // 10 requests/sec, burst 20

	// Health check endpoint
	router.GET("/health", systemHandler.Health)

	// Root endpoint
	router.GET("/", systemHandler.Root)

	// API v1 routes
	setupV1Routes(router, cfg, app)

	return router
}

// setupV1Routes sets up API v1 routes
func setupV1Routes(router *gin.Engine, cfg *config.Config, app *App) {
	v1 := router.Group("/api/v1")

	// Create handlers
	messageHandler := handlers.NewMessageHandler(app.rabbitMQ)

	// Message routes (basic)
	messages := v1.Group("/messages")
	{
		messages.POST("/send", messageHandler.SendMessage)
	}

	// Extended message routes (with database)
	if app.messageService != nil {
		extMessageHandler := handlers.NewMessageHandlerExtended(app.messageService)

		messages.GET("/recent", extMessageHandler.GetRecentMessages)
		messages.GET("/stats", extMessageHandler.GetMessageStats)
		messages.GET("/:messageID", extMessageHandler.GetMessage)
		messages.PATCH("/:messageID/status", extMessageHandler.UpdateMessageStatus)
		messages.DELETE("/:messageID", extMessageHandler.DeleteMessage)
		messages.GET("/status/:status", extMessageHandler.GetMessagesByStatus)
	}

	// User routes (with database)
	if app.userService != nil {
		userHandler := handlers.NewUserHandler(app.userService)

		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.ListUsers)
			users.GET("/online", userHandler.GetOnlineUsers)
			users.GET("/:userID", userHandler.GetUser)
			users.PUT("/:userID/status", userHandler.UpdateStatus)
			users.DELETE("/:userID", userHandler.DeleteUser)

			// User messages
			if app.messageService != nil {
				extMessageHandler := handlers.NewMessageHandlerExtended(app.messageService)
				users.GET("/:userID/messages", extMessageHandler.GetUserMessages)
			}
		}
	}

	// Protected routes (with auth)
	if cfg.Auth.Enabled {
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.Auth.JWTSecret))
		{
			// Add protected routes here
		}
	}
}
