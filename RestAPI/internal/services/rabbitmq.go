package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/config"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/logger"
)

// RabbitMQService handles RabbitMQ operations
type RabbitMQService struct {
	config  *config.RabbitMQConfig
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.RWMutex
	closed  bool
}

// NewRabbitMQService creates a new RabbitMQ service
func NewRabbitMQService(cfg *config.RabbitMQConfig) (*RabbitMQService, error) {
	service := &RabbitMQService{
		config: cfg,
		closed: false,
	}

	if err := service.connect(); err != nil {
		return nil, err
	}

	return service, nil
}

// connect establishes a connection to RabbitMQ
func (s *RabbitMQService) connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	url := s.getConnectionURL()

	var err error
	var conn *amqp.Connection

	// Retry connection with exponential backoff
	for attempt := 1; attempt <= s.config.ConnectionRetry; attempt++ {
		logger.Infof("Attempting to connect to RabbitMQ (attempt %d/%d)", attempt, s.config.ConnectionRetry)

		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}

		if attempt < s.config.ConnectionRetry {
			delay := time.Duration(s.config.RetryDelay) * time.Second
			logger.Warnf("Failed to connect to RabbitMQ: %v. Retrying in %v...", err, delay)
			time.Sleep(delay)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", s.config.ConnectionRetry, err)
	}

	s.conn = conn
	logger.Info("Successfully connected to RabbitMQ")

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		s.conn.Close()
		return fmt.Errorf("failed to create channel: %w", err)
	}

	s.channel = channel
	logger.Info("Successfully created RabbitMQ channel")

	// Declare queue
	if err := s.declareQueue(); err != nil {
		s.channel.Close()
		s.conn.Close()
		return err
	}

	// Setup connection close handler
	go s.handleReconnect()

	return nil
}

// declareQueue declares the queue
func (s *RabbitMQService) declareQueue() error {
	args := amqp.Table{}

	_, err := s.channel.QueueDeclare(
		s.config.QueueName,
		s.config.Durable,
		s.config.AutoDelete,
		s.config.Exclusive,
		s.config.NoWait,
		args,
	)

	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	logger.Infof("Successfully declared queue: %s", s.config.QueueName)
	return nil
}

// handleReconnect handles reconnection when connection is closed
func (s *RabbitMQService) handleReconnect() {
	for {
		reason, ok := <-s.conn.NotifyClose(make(chan *amqp.Error))
		if !ok {
			logger.Info("RabbitMQ connection closed normally")
			return
		}

		s.mu.RLock()
		closed := s.closed
		s.mu.RUnlock()

		if closed {
			logger.Info("RabbitMQ service is closing, not reconnecting")
			return
		}

		logger.Errorf("RabbitMQ connection closed unexpectedly: %v. Attempting to reconnect...", reason)

		// Attempt to reconnect
		for {
			time.Sleep(time.Duration(s.config.RetryDelay) * time.Second)

			err := s.connect()
			if err == nil {
				logger.Info("Successfully reconnected to RabbitMQ")
				return
			}

			logger.Errorf("Failed to reconnect to RabbitMQ: %v. Retrying...", err)
		}
	}
}

// Publish publishes a message to the queue
func (s *RabbitMQService) Publish(ctx context.Context, message []byte) error {
	return s.PublishWithPriority(ctx, message, 0)
}

// PublishWithPriority publishes a message to the queue with specified priority
func (s *RabbitMQService) PublishWithPriority(ctx context.Context, message []byte, priority uint8) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return fmt.Errorf("rabbitmq service is closed")
	}

	if s.channel == nil {
		return fmt.Errorf("rabbitmq channel is not initialized")
	}

	publishing := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         message,
		Timestamp:    time.Now(),
		Priority:     priority,
	}

	err := s.channel.PublishWithContext(
		ctx,
		"",                 // exchange
		s.config.QueueName, // routing key
		false,              // mandatory
		false,              // immediate
		publishing,
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logger.WithField("queue", s.config.QueueName).Debug("Message published successfully")
	return nil
}

// Close closes the RabbitMQ connection
func (s *RabbitMQService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	var errs []error

	if s.channel != nil {
		if err := s.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	logger.Info("RabbitMQ service closed successfully")
	return nil
}

// IsHealthy checks if the RabbitMQ connection is healthy
func (s *RabbitMQService) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed || s.conn == nil || s.conn.IsClosed() {
		return false
	}

	return true
}

// QueueName returns the configured queue name
func (s *RabbitMQService) QueueName() string {
	return s.config.QueueName
}

// getConnectionURL builds the RabbitMQ connection URL
func (s *RabbitMQService) getConnectionURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		s.config.Username,
		s.config.Password,
		s.config.Host,
		s.config.Port,
		s.config.VHost,
	)
}
