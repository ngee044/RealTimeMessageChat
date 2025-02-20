package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads the environment variables from the service's .env file
func LoadEnv(service string) {
	envFile := "../" + service + "/.env" // 각 서비스별 .env 파일 사용
	err := godotenv.Load(envFile)
	if err != nil {
		log.Printf("Warning: No .env file found for %s, using system environment variables", service)
	}
}

// GetEnv retrieves an environment variable with a default fallback
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
