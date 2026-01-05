package infrastructure

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// CacheService defines the interface for cache operations
type CacheService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// Hash operations
	HSet(ctx context.Context, key string, values ...interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) error

	// Counter operations
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)

	// Sorted set operations
	ZAdd(ctx context.Context, key string, members ...redis.Z) error
	ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error)
	ZRemRangeByScore(ctx context.Context, key string, min, max string) error

	// Pub/Sub
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub

	// Health
	IsHealthy(ctx context.Context) bool
	Close() error
}

// MessageQueue defines the interface for message queue operations
type MessageQueue interface {
	Publish(ctx context.Context, message []byte) error
	PublishWithPriority(ctx context.Context, message []byte, priority uint8) error
	QueueName() string
	IsHealthy() bool
	Close() error
}

// Database defines the interface for database operations
type Database interface {
	GetDB() *sqlx.DB
	IsHealthy() bool
	InitSchema() error
	Close() error
}

// Closeable defines interface for resources that can be closed
type Closeable interface {
	Close() error
}

// HealthChecker defines interface for health check capable services
type HealthChecker interface {
	IsHealthy() bool
}
