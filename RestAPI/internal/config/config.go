package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	RabbitMQ RabbitMQConfig `json:"rabbitmq"`
	Redis    RedisConfig    `json:"redis"`
	Database DatabaseConfig `json:"database"`
	Logging  LoggingConfig  `json:"logging"`
	Auth     AuthConfig     `json:"auth"`
	Metrics  MetricsConfig  `json:"metrics"`
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

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	DialTimeout  int    `json:"dial_timeout_seconds"`
	ReadTimeout  int    `json:"read_timeout_seconds"`
	WriteTimeout int    `json:"write_timeout_seconds"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
	Enabled      bool   `json:"enabled"`
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"`
	DBName          string `json:"dbname"`
	SSLMode         string `json:"sslmode"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime_minutes"`
	Enabled         bool   `json:"enabled"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret           string `json:"jwt_secret"`
	JWTExpirationHours  int    `json:"jwt_expiration_hours"`
	RefreshExpirationHours int `json:"refresh_expiration_hours"`
	Enabled             bool   `json:"enabled"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
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

	// Override with environment variables if present
	config.loadFromEnv()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// getEnvString returns the environment variable value or empty string
func getEnvString(key string) string {
	return os.Getenv(key)
}

// getEnvInt returns the environment variable as int, or 0 if not set/invalid
// Logs a warning if the value is set but invalid
func getEnvInt(key string) (int, bool) {
	val := os.Getenv(key)
	if val == "" {
		return 0, false
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("[WARN] Invalid integer value for %s: %q, ignoring", key, val)
		return 0, false
	}

	return intVal, true
}

// loadFromEnv loads sensitive configuration from environment variables
// Environment variables take precedence over config file values
func (c *Config) loadFromEnv() {
	// === Sensitive credentials (highest priority) ===

	// JWT secret (most sensitive)
	if jwtSecret := getEnvString("JWT_SECRET"); jwtSecret != "" {
		c.Auth.JWTSecret = jwtSecret
	}

	// Database password
	if dbPassword := getEnvString("DB_PASSWORD"); dbPassword != "" {
		c.Database.Password = dbPassword
	}

	// RabbitMQ credentials
	if rabbitPassword := getEnvString("RABBITMQ_PASSWORD"); rabbitPassword != "" {
		c.RabbitMQ.Password = rabbitPassword
	}
	if rabbitUser := getEnvString("RABBITMQ_USERNAME"); rabbitUser != "" {
		c.RabbitMQ.Username = rabbitUser
	}

	// Redis password
	if redisPassword := getEnvString("REDIS_PASSWORD"); redisPassword != "" {
		c.Redis.Password = redisPassword
	}

	// === Database connection parameters ===
	if dbHost := getEnvString("DB_HOST"); dbHost != "" {
		c.Database.Host = dbHost
	}
	if port, ok := getEnvInt("DB_PORT"); ok {
		c.Database.Port = port
	}
	if dbName := getEnvString("DB_NAME"); dbName != "" {
		c.Database.DBName = dbName
	}
	if dbUser := getEnvString("DB_USER"); dbUser != "" {
		c.Database.User = dbUser
	}

	// === RabbitMQ connection parameters ===
	if rabbitHost := getEnvString("RABBITMQ_HOST"); rabbitHost != "" {
		c.RabbitMQ.Host = rabbitHost
	}
	if port, ok := getEnvInt("RABBITMQ_PORT"); ok {
		c.RabbitMQ.Port = port
	}

	// === Redis connection parameters ===
	if redisHost := getEnvString("REDIS_HOST"); redisHost != "" {
		c.Redis.Host = redisHost
	}
	if port, ok := getEnvInt("REDIS_PORT"); ok {
		c.Redis.Port = port
	}

	// === Server configuration ===
	if port, ok := getEnvInt("SERVER_PORT"); ok {
		c.Server.Port = port
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout_seconds must be greater than 0")
	}

	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout_seconds must be greater than 0")
	}

	if c.Server.MaxHeaderBytes <= 0 {
		return fmt.Errorf("max_header_bytes must be greater than 0")
	}

	if c.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown_timeout_seconds must be greater than 0")
	}

	if c.RabbitMQ.Host == "" {
		return fmt.Errorf("rabbitmq host is required")
	}

	if c.RabbitMQ.QueueName == "" {
		return fmt.Errorf("rabbitmq queue name is required")
	}

	if c.RabbitMQ.ConnectionRetry < 1 || c.RabbitMQ.ConnectionRetry > 5 {
		return fmt.Errorf("rabbitmq connection_retry must be between 1 and 5")
	}

	if c.RabbitMQ.RetryDelay <= 0 {
		return fmt.Errorf("rabbitmq retry_delay_seconds must be greater than 0")
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

	// Validate JWT secret if authentication is enabled
	if c.Auth.Enabled {
		if c.Auth.JWTSecret == "" {
			return fmt.Errorf("jwt_secret is required when authentication is enabled")
		}
		// Minimum 32 bytes (256 bits) for HMAC-SHA256
		if len(c.Auth.JWTSecret) < 32 {
			return fmt.Errorf("jwt_secret must be at least 32 characters for security (current: %d)", len(c.Auth.JWTSecret))
		}
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

// GetDatabaseDSN returns the PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetRedisAddress returns the Redis address
func (c *Config) GetRedisAddress() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}
