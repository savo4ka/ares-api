package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerPort      string
	DatabaseURL     string
	EncryptionKey   string
	AllowedOrigins  string
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	config := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		EncryptionKey:  getEnv("ENCRYPTION_KEY", ""),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
	}

	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if config.EncryptionKey == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY is required")
	}

	// Проверяем, что ключ шифрования имеет правильную длину (16 байт для AES-128)
	if len(config.EncryptionKey) != 16 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be exactly 16 characters for AES-128")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
