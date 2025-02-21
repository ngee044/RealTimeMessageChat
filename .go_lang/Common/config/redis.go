package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var redisClient *redis.Client

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

func SetKey(key string, value string, ttl int) error {
	expiration := time.Duration(ttl) * time.Second // TTL 적용
	err := redisClient.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Printf("Redis SET failed: %v", err)
		return err
	}
	log.Printf("Key '%s' set successfully with TTL %v seconds", key, expiration)
	return nil
}

func GetKey(key string) (string, error) {
	val, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("Key '%s' not found", key)
		return "", nil
	} else if err != nil {
		log.Printf("Redis GET failed: %v", err)
		return "", err
	}
	return val, nil
}
