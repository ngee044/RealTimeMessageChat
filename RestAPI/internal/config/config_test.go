package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a temporary config file
func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")

	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	return configPath
}

// Helper to set and cleanup environment variables
func setEnvWithCleanup(t *testing.T, key, value string) {
	t.Helper()

	oldValue, existed := os.LookupEnv(key)
	os.Setenv(key, value)

	t.Cleanup(func() {
		if existed {
			os.Setenv(key, oldValue)
		} else {
			os.Unsetenv(key)
		}
	})
}

// Valid minimal config for testing
const validConfigJSON = `{
	"server": {
		"host": "0.0.0.0",
		"port": 8080,
		"mode": "release",
		"read_timeout_seconds": 30,
		"write_timeout_seconds": 30,
		"max_header_bytes": 1048576,
		"shutdown_timeout_seconds": 15
	},
	"rabbitmq": {
		"host": "localhost",
		"port": 5672,
		"username": "guest",
		"password": "guest",
		"vhost": "/",
		"queue_name": "test_queue",
		"durable": true,
		"auto_delete": false,
		"exclusive": false,
		"no_wait": false,
		"connection_retry": 3,
		"retry_delay_seconds": 5
	},
	"redis": {
		"host": "localhost",
		"port": 6379,
		"password": "",
		"db": 0,
		"dial_timeout_seconds": 5,
		"read_timeout_seconds": 3,
		"write_timeout_seconds": 3,
		"pool_size": 10,
		"min_idle_conns": 5,
		"enabled": false
	},
	"database": {
		"host": "localhost",
		"port": 5432,
		"user": "postgres",
		"password": "postgres",
		"dbname": "testdb",
		"sslmode": "disable",
		"max_open_conns": 25,
		"max_idle_conns": 5,
		"conn_max_lifetime_minutes": 5,
		"enabled": true
	},
	"logging": {
		"level": "info",
		"format": "json",
		"output_path": "logs/test.log"
	},
	"auth": {
		"jwt_secret": "test-secret-key-for-testing-only-32ch",
		"jwt_expiration_hours": 24,
		"refresh_expiration_hours": 168,
		"enabled": false
	},
	"metrics": {
		"enabled": false,
		"path": "/metrics"
	}
}`

func TestLoadConfig_Success(t *testing.T) {
	configPath := createTempConfigFile(t, validConfigJSON)

	cfg, err := LoadConfig(configPath)

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.RabbitMQ.Host)
	assert.Equal(t, "test_queue", cfg.RabbitMQ.QueueName)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.json")

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to open config file")
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	configPath := createTempConfigFile(t, `{invalid json}`)

	cfg, err := LoadConfig(configPath)

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to decode config file")
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	configPath := createTempConfigFile(t, validConfigJSON)

	// Set environment variables
	setEnvWithCleanup(t, "DB_PASSWORD", "env_db_password")
	setEnvWithCleanup(t, "DB_HOST", "env-db-host")
	setEnvWithCleanup(t, "DB_PORT", "5433")
	setEnvWithCleanup(t, "JWT_SECRET", "env-jwt-secret-that-is-long-enough-for-testing")
	setEnvWithCleanup(t, "SERVER_PORT", "9090")
	setEnvWithCleanup(t, "RABBITMQ_PASSWORD", "env_rabbit_pass")
	setEnvWithCleanup(t, "REDIS_PASSWORD", "env_redis_pass")

	cfg, err := LoadConfig(configPath)

	require.NoError(t, err)
	assert.Equal(t, "env_db_password", cfg.Database.Password)
	assert.Equal(t, "env-db-host", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "env-jwt-secret-that-is-long-enough-for-testing", cfg.Auth.JWTSecret)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "env_rabbit_pass", cfg.RabbitMQ.Password)
	assert.Equal(t, "env_redis_pass", cfg.Redis.Password)
}

func TestLoadConfig_EnvInvalidPort(t *testing.T) {
	configPath := createTempConfigFile(t, validConfigJSON)

	// Set invalid port (should be ignored)
	setEnvWithCleanup(t, "DB_PORT", "not_a_number")

	cfg, err := LoadConfig(configPath)

	require.NoError(t, err)
	// Should keep the original value from config file
	assert.Equal(t, 5432, cfg.Database.Port)
}

func TestGetEnvInt(t *testing.T) {
	t.Run("valid integer", func(t *testing.T) {
		setEnvWithCleanup(t, "TEST_INT", "12345")

		val, ok := getEnvInt("TEST_INT")

		assert.True(t, ok)
		assert.Equal(t, 12345, val)
	})

	t.Run("empty value", func(t *testing.T) {
		setEnvWithCleanup(t, "TEST_INT_EMPTY", "")

		val, ok := getEnvInt("TEST_INT_EMPTY")

		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})

	t.Run("not set", func(t *testing.T) {
		os.Unsetenv("TEST_INT_NOTSET")

		val, ok := getEnvInt("TEST_INT_NOTSET")

		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})

	t.Run("invalid integer", func(t *testing.T) {
		setEnvWithCleanup(t, "TEST_INT_INVALID", "abc")

		val, ok := getEnvInt("TEST_INT_INVALID")

		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})
}

func TestValidate_ServerPort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"valid port", 8080, false},
		{"min valid port", 1, false},
		{"max valid port", 65535, false},
		{"zero port", 0, true},
		{"negative port", -1, true},
		{"port too high", 65536, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			cfg.Server.Port = tt.port

			err := cfg.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "port")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_ServerMode(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{"debug mode", "debug", false},
		{"release mode", "release", false},
		{"test mode", "test", false},
		{"invalid mode", "production", true},
		{"empty mode", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			cfg.Server.Mode = tt.mode

			err := cfg.Validate()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_LoggingLevel(t *testing.T) {
	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}

	for _, level := range validLevels {
		t.Run("valid level "+level, func(t *testing.T) {
			cfg := createValidConfig()
			cfg.Logging.Level = level

			err := cfg.Validate()

			assert.NoError(t, err)
		})
	}

	t.Run("invalid level", func(t *testing.T) {
		cfg := createValidConfig()
		cfg.Logging.Level = "invalid"

		err := cfg.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid logging level")
	})
}

func TestValidate_JWTSecret(t *testing.T) {
	t.Run("auth disabled - no secret required", func(t *testing.T) {
		cfg := createValidConfig()
		cfg.Auth.Enabled = false
		cfg.Auth.JWTSecret = ""

		err := cfg.Validate()

		assert.NoError(t, err)
	})

	t.Run("auth enabled - secret required", func(t *testing.T) {
		cfg := createValidConfig()
		cfg.Auth.Enabled = true
		cfg.Auth.JWTSecret = ""

		err := cfg.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "jwt_secret is required")
	})

	t.Run("auth enabled - secret too short", func(t *testing.T) {
		cfg := createValidConfig()
		cfg.Auth.Enabled = true
		cfg.Auth.JWTSecret = "short"

		err := cfg.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 32 characters")
	})

	t.Run("auth enabled - valid secret", func(t *testing.T) {
		cfg := createValidConfig()
		cfg.Auth.Enabled = true
		cfg.Auth.JWTSecret = "this-is-a-very-long-secret-key-that-is-at-least-32-characters"

		err := cfg.Validate()

		assert.NoError(t, err)
	})
}

func TestValidate_RabbitMQ(t *testing.T) {
	t.Run("missing host", func(t *testing.T) {
		cfg := createValidConfig()
		cfg.RabbitMQ.Host = ""

		err := cfg.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rabbitmq host is required")
	})

	t.Run("missing queue name", func(t *testing.T) {
		cfg := createValidConfig()
		cfg.RabbitMQ.QueueName = ""

		err := cfg.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rabbitmq queue name is required")
	})

	t.Run("invalid connection retry", func(t *testing.T) {
		cfg := createValidConfig()
		cfg.RabbitMQ.ConnectionRetry = 10

		err := cfg.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection_retry must be between 1 and 5")
	})
}

func TestGetRabbitMQURL(t *testing.T) {
	cfg := createValidConfig()
	cfg.RabbitMQ.Username = "user"
	cfg.RabbitMQ.Password = "pass"
	cfg.RabbitMQ.Host = "rabbit.example.com"
	cfg.RabbitMQ.Port = 5672
	cfg.RabbitMQ.VHost = "/vhost"

	url := cfg.GetRabbitMQURL()

	assert.Equal(t, "amqp://user:pass@rabbit.example.com:5672/vhost", url)
}

func TestGetDatabaseDSN(t *testing.T) {
	cfg := createValidConfig()
	cfg.Database.Host = "db.example.com"
	cfg.Database.Port = 5432
	cfg.Database.User = "dbuser"
	cfg.Database.Password = "dbpass"
	cfg.Database.DBName = "mydb"
	cfg.Database.SSLMode = "require"

	dsn := cfg.GetDatabaseDSN()

	assert.Contains(t, dsn, "host=db.example.com")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=dbuser")
	assert.Contains(t, dsn, "password=dbpass")
	assert.Contains(t, dsn, "dbname=mydb")
	assert.Contains(t, dsn, "sslmode=require")
}

func TestGetServerAddress(t *testing.T) {
	cfg := createValidConfig()
	cfg.Server.Host = "192.168.1.1"
	cfg.Server.Port = 9090

	addr := cfg.GetServerAddress()

	assert.Equal(t, "192.168.1.1:9090", addr)
}

func TestGetRedisAddress(t *testing.T) {
	cfg := createValidConfig()
	cfg.Redis.Host = "redis.example.com"
	cfg.Redis.Port = 6380

	addr := cfg.GetRedisAddress()

	assert.Equal(t, "redis.example.com:6380", addr)
}

// Helper to create a valid config for testing
func createValidConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			Mode:            "release",
			ReadTimeout:     30,
			WriteTimeout:    30,
			MaxHeaderBytes:  1048576,
			ShutdownTimeout: 15,
		},
		RabbitMQ: RabbitMQConfig{
			Host:            "localhost",
			Port:            5672,
			Username:        "guest",
			Password:        "guest",
			VHost:           "/",
			QueueName:       "test_queue",
			ConnectionRetry: 3,
			RetryDelay:      5,
		},
		Redis: RedisConfig{
			Host:    "localhost",
			Port:    6379,
			Enabled: false,
		},
		Database: DatabaseConfig{
			Host:    "localhost",
			Port:    5432,
			User:    "postgres",
			DBName:  "testdb",
			SSLMode: "disable",
			Enabled: true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "logs/test.log",
		},
		Auth: AuthConfig{
			JWTSecret: "test-secret-key-for-testing-purposes",
			Enabled:   false,
		},
		Metrics: MetricsConfig{
			Enabled: false,
			Path:    "/metrics",
		},
	}
}
