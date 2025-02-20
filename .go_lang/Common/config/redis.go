package config

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var redisClient *redis.Client

// ConnectRedis initializes the Redis client
func ConnectRedis(service string) {
	LoadEnv(service) // 서비스별 환경 변수 로드

	redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", GetEnv("REDIS_HOST", "localhost"), GetEnv("REDIS_PORT", "6379")),
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis")
}

// SetKey sets a value in Redis
func SetKey(key string, value string) error {
	return redisClient.Set(ctx, key, value, 0).Err()
}

// GetKey retrieves a value from Redis
func GetKey(key string) (string, error) {
	return redisClient.Get(ctx, key).Result()
}
