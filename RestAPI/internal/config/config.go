package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	RabbitMQ RabbitMQConfig `json:"rabbitmq"`
	Logging  LoggingConfig  `json:"logging"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Mode            string `json:"mode"` // debug, release, test
	ReadTimeout     int    `json:"read_timeout_seconds"`
	WriteTimeout    int    `json:"write_timeout_seconds"`
	MaxHeaderBytes  int    `json:"max_header_bytes"`
	ShutdownTimeout int    `json:"shutdown_timeout_seconds"`
}

// RabbitMQConfig holds RabbitMQ connection and queue configuration
type RabbitMQConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	VHost           string `json:"vhost"`
	QueueName       string `json:"queue_name"`
	Durable         bool   `json:"durable"`
	AutoDelete      bool   `json:"auto_delete"`
	Exclusive       bool   `json:"exclusive"`
	NoWait          bool   `json:"no_wait"`
	ConnectionRetry int    `json:"connection_retry"`
	RetryDelay      int    `json:"retry_delay_seconds"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level"` // trace, debug, info, warn, error, fatal, panic
	Format     string `json:"format"` // json, text
	OutputPath string `json:"output_path"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.RabbitMQ.Host == "" {
		return fmt.Errorf("rabbitmq host is required")
	}

	if c.RabbitMQ.QueueName == "" {
		return fmt.Errorf("rabbitmq queue name is required")
	}

	validModes := map[string]bool{"debug": true, "release": true, "test": true}
	if !validModes[c.Server.Mode] {
		return fmt.Errorf("invalid server mode: %s", c.Server.Mode)
	}

	validLevels := map[string]bool{
		"trace": true, "debug": true, "info": true,
		"warn": true, "error": true, "fatal": true, "panic": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid logging level: %s", c.Logging.Level)
	}

	return nil
}

// GetRabbitMQURL returns the RabbitMQ connection URL
func (c *Config) GetRabbitMQURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		c.RabbitMQ.Username,
		c.RabbitMQ.Password,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
		c.RabbitMQ.VHost,
	)
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
