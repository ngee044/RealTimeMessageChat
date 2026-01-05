package services

import (
	"fmt"
	"time"

	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/config"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DatabaseService handles PostgreSQL database operations
type DatabaseService struct {
	db     *sqlx.DB
	config *config.DatabaseConfig
}

// NewDatabaseService creates a new database service
func NewDatabaseService(cfg *config.DatabaseConfig) (*DatabaseService, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Infof("Successfully connected to PostgreSQL at %s:%d (DB: %s)", cfg.Host, cfg.Port, cfg.DBName)

	return &DatabaseService{
		db:     db,
		config: cfg,
	}, nil
}

// GetDB returns the underlying database connection
func (d *DatabaseService) GetDB() *sqlx.DB {
	return d.db
}

// Close closes the database connection
func (d *DatabaseService) Close() error {
	if d.db != nil {
		logger.Info("Closing database connection")
		return d.db.Close()
	}
	return nil
}

// IsHealthy checks if the database connection is healthy
func (d *DatabaseService) IsHealthy() bool {
	return d.db.Ping() == nil
}

// InitSchema initializes the database schema
func (d *DatabaseService) InitSchema() error {
	schema := `
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(50) UNIQUE NOT NULL,
		username VARCHAR(100),
		email VARCHAR(255),
		status VARCHAR(20) DEFAULT 'offline',
		last_seen TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Messages table
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		message_id UUID UNIQUE NOT NULL,
		user_id VARCHAR(50) NOT NULL,
		command VARCHAR(50) NOT NULL,
		sub_id VARCHAR(100),
		content TEXT NOT NULL,
		metadata JSONB,
		priority INTEGER DEFAULT 2,
		status VARCHAR(20) DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		processed_at TIMESTAMP
	);

	-- Message history table
	CREATE TABLE IF NOT EXISTS message_history (
		id SERIAL PRIMARY KEY,
		message_id UUID NOT NULL,
		user_id VARCHAR(50) NOT NULL,
		action VARCHAR(50) NOT NULL,
		details JSONB,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_users_user_id ON users(user_id);
	CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
	CREATE INDEX IF NOT EXISTS idx_messages_message_id ON messages(message_id);
	CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id);
	CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
	CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
	CREATE INDEX IF NOT EXISTS idx_message_history_message_id ON message_history(message_id);
	CREATE INDEX IF NOT EXISTS idx_message_history_user_id ON message_history(user_id);

	-- Create updated_at trigger function
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	-- Create trigger for users table
	DROP TRIGGER IF EXISTS update_users_updated_at ON users;
	CREATE TRIGGER update_users_updated_at
		BEFORE UPDATE ON users
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();
	`

	_, err := d.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	logger.Info("Database schema initialized successfully")
	return nil
}
