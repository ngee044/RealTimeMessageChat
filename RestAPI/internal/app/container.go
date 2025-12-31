package app

import (
	"context"

	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/config"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/infrastructure"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/repository"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/service"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/services"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/logger"
)

// Container holds all application dependencies
type Container struct {
	Config *config.Config

	// Infrastructure services
	RabbitMQ infrastructure.MessageQueue
	Redis    infrastructure.CacheService
	Database infrastructure.Database

	// Repositories
	UserRepo    repository.UserRepository
	MessageRepo repository.MessageRepository

	// Business services
	UserService    service.UserServiceInterface
	MessageService service.MessageServiceInterface
}

// NewContainer creates and initializes all dependencies
func NewContainer(cfg *config.Config) (*Container, error) {
	c := &Container{Config: cfg}

	// Initialize infrastructure
	if err := c.initInfrastructure(); err != nil {
		return nil, err
	}

	// Initialize repositories
	c.initRepositories()

	// Initialize services
	c.initServices()

	return c, nil
}

// initInfrastructure initializes all infrastructure services
func (c *Container) initInfrastructure() error {
	// RabbitMQ (required)
	rabbitMQ, err := services.NewRabbitMQService(&c.Config.RabbitMQ)
	if err != nil {
		return err
	}
	c.RabbitMQ = rabbitMQ

	// Redis (optional)
	if c.Config.Redis.Enabled {
		redis, err := services.NewRedisService(&c.Config.Redis)
		if err != nil {
			logger.Warnf("Failed to initialize Redis: %v. Continuing without Redis.", err)
		} else {
			c.Redis = redis
		}
	}

	// Database (optional)
	if c.Config.Database.Enabled {
		db, err := services.NewDatabaseService(&c.Config.Database)
		if err != nil {
			logger.Warnf("Failed to initialize database: %v. Continuing without database.", err)
		} else {
			c.Database = db
			// Initialize schema
			if err := db.InitSchema(); err != nil {
				logger.Warnf("Failed to initialize database schema: %v", err)
			}
		}
	}

	return nil
}

// initRepositories initializes all repositories
func (c *Container) initRepositories() {
	if c.Database == nil {
		return
	}

	c.UserRepo = repository.NewUserRepository(c.Database.GetDB())
	c.MessageRepo = repository.NewMessageRepository(c.Database.GetDB())
}

// initServices initializes all business services
func (c *Container) initServices() {
	// Get concrete Redis service for now (can be refactored later)
	var redisService *services.RedisService
	if rs, ok := c.Redis.(*services.RedisService); ok {
		redisService = rs
	}

	// Get concrete RabbitMQ service
	var rabbitMQService *services.RabbitMQService
	if rmq, ok := c.RabbitMQ.(*services.RabbitMQService); ok {
		rabbitMQService = rmq
	}

	if c.UserRepo != nil {
		c.UserService = service.NewUserService(c.UserRepo, redisService)
	}

	if c.MessageRepo != nil && rabbitMQService != nil {
		c.MessageService = service.NewMessageService(
			c.MessageRepo,
			c.UserRepo,
			rabbitMQService,
			redisService,
		)
	}
}

// Close closes all services
func (c *Container) Close() {
	if c.RabbitMQ != nil {
		c.RabbitMQ.Close()
	}
	if c.Redis != nil {
		c.Redis.Close()
	}
	if c.Database != nil {
		c.Database.Close()
	}
}

// HealthStatus returns the health status of all services
type HealthStatus struct {
	RabbitMQ bool `json:"rabbitmq"`
	Redis    bool `json:"redis"`
	Database bool `json:"database"`
	Overall  bool `json:"overall"`
}

// GetHealthStatus checks health of all services
func (c *Container) GetHealthStatus() HealthStatus {
	status := HealthStatus{
		RabbitMQ: true,
		Redis:    true,
		Database: true,
	}

	if c.RabbitMQ != nil {
		status.RabbitMQ = c.RabbitMQ.IsHealthy()
	}

	// Redis는 optional이므로 nil이면 healthy로 취급
	if c.Redis != nil {
		ctx := context.Background()
		status.Redis = c.Redis.IsHealthy(ctx)
	}

	if c.Database != nil {
		status.Database = c.Database.IsHealthy()
	}

	// Overall status - critical services only
	status.Overall = status.RabbitMQ && status.Database

	return status
}
