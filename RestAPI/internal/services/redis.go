package services

import (
	"context"
	"fmt"
	"time"

	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/config"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// RedisService handles Redis operations
type RedisService struct {
	client *redis.Client
	config *config.RedisConfig
}

// NewRedisService creates a new Redis service
func NewRedisService(cfg *config.RedisConfig) (*RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Infof("Successfully connected to Redis at %s:%d (DB: %d)", cfg.Host, cfg.Port, cfg.DB)

	return &RedisService{
		client: client,
		config: cfg,
	}, nil
}

// Set sets a key-value pair with optional expiration
func (r *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (r *RedisService) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return val, err
}

// Delete deletes one or more keys
func (r *RedisService) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *RedisService) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// SetNX sets a key only if it doesn't exist (with expiration)
func (r *RedisService) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

// Expire sets an expiration time on a key
func (r *RedisService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// HSet sets a field in a hash
func (r *RedisService) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGet gets a field from a hash
func (r *RedisService) HGet(ctx context.Context, key, field string) (string, error) {
	val, err := r.client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("field not found: %s", field)
	}
	return val, err
}

// HGetAll gets all fields from a hash
func (r *RedisService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel deletes fields from a hash
func (r *RedisService) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// Incr increments a counter
func (r *RedisService) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Decr decrements a counter
func (r *RedisService) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// ZAdd adds members to a sorted set
func (r *RedisService) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return r.client.ZAdd(ctx, key, members...).Err()
}

// ZRangeByScore retrieves members from sorted set by score range
func (r *RedisService) ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
}

// ZRemRangeByScore removes members from sorted set by score range
func (r *RedisService) ZRemRangeByScore(ctx context.Context, key string, min, max string) error {
	return r.client.ZRemRangeByScore(ctx, key, min, max).Err()
}

// Publish publishes a message to a channel
func (r *RedisService) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels
func (r *RedisService) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// IsHealthy checks if Redis connection is healthy
func (r *RedisService) IsHealthy(ctx context.Context) bool {
	return r.client.Ping(ctx).Err() == nil
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	if r.client != nil {
		logger.Info("Closing Redis connection")
		return r.client.Close()
	}
	return nil
}

// GetClient returns the underlying Redis client for advanced operations
func (r *RedisService) GetClient() *redis.Client {
	return r.client
}
