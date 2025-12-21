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
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/services"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/logger"
)

var (
	configPath = flag.String("config", "config/api_server_config.json", "Path to configuration file")
	version    = "1.0.0"
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

	// Initialize RabbitMQ service
	rabbitMQ, err := services.NewRabbitMQService(&cfg.RabbitMQ)
	if err != nil {
		logger.Fatalf("Failed to initialize RabbitMQ service: %v", err)
	}
	defer rabbitMQ.Close()

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create Gin router
	router := gin.New()

	// Apply middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Create handlers
	messageHandler := handlers.NewMessageHandler(rabbitMQ)

	// Register routes
	registerRoutes(router, messageHandler)

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

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
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

// registerRoutes registers all HTTP routes
func registerRoutes(router *gin.Engine, messageHandler *handlers.MessageHandler) {
	// Health check endpoint
	router.GET("/health", messageHandler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Message routes
		v1.POST("/send_message", messageHandler.SendMessage)
	}

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "RealTimeMessageChat REST API",
			"version": version,
			"status":  "running",
		})
	})
}
