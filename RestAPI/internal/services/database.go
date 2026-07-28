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

// requiredSchema lists the columns this service reads or writes.
// DDL 은 database/schema.sql 이 단일 출처다 — 여기서 테이블을 만들지 않는다.
// 예전에는 이 파일이 같은 이름의 messages 를 다른 컬럼으로 CREATE TABLE IF NOT EXISTS 해서,
// initdb 가 먼저 만든 정의와 어긋난 채 조용히 no-op 이 되고 이후 모든 쿼리가 실패했다.
var requiredSchema = map[string][]string{
	"users": {
		"id", "user_id", "username", "email", "status", "last_seen", "created_at", "updated_at",
	},
	"messages": {
		"id", "message_id", "user_id", "sub_id", "command", "publisher_info",
		"server_name", "content", "is_encrypted", "status", "created_at", "processed_at",
	},
}

// VerifySchema fails fast when the database does not match database/schema.sql.
// 여기서 넘어가면 확장 라우트가 등록된 채 모든 쿼리가 500 을 내므로 호출측은 치명 오류로 다뤄야 한다.
func (d *DatabaseService) VerifySchema() error {
	for table, columns := range requiredSchema {
		var found []string
		query := `SELECT column_name FROM information_schema.columns
		          WHERE table_schema = 'public' AND table_name = $1`

		if err := d.db.Select(&found, query, table); err != nil {
			return fmt.Errorf("failed to inspect table %q: %w", table, err)
		}

		if len(found) == 0 {
			return fmt.Errorf("table %q is missing - apply database/schema.sql", table)
		}

		present := make(map[string]struct{}, len(found))
		for _, column := range found {
			present[column] = struct{}{}
		}

		var missing []string
		for _, column := range columns {
			if _, ok := present[column]; !ok {
				missing = append(missing, column)
			}
		}

		if len(missing) > 0 {
			return fmt.Errorf("table %q is missing column(s) %v - apply database/migrations/001_unify_messages.sql", table, missing)
		}
	}

	logger.Info("Database schema verified")
	return nil
}
