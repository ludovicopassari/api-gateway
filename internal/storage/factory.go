package storage

import (
	"fmt"
	"os"
)

type StorageType string

const (
	Memory   StorageType = "memory"
	Redis    StorageType = "redis"
	Postgres StorageType = "postgres"
)

type Config struct {
	Type StorageType

	// Redis config
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Postgres config
	PostgresConnStr string
}

func NewStorageFromEnv() (Storage, error) {
	storageType := StorageType(os.Getenv("STORAGE_TYPE"))
	if storageType == "" {
		storageType = Memory // Default
	}

	config := Config{
		Type:            storageType,
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   os.Getenv("REDIS_PASSWORD"),
		RedisDB:         getEnvInt("REDIS_DB", 0),
		PostgresConnStr: getEnv("POSTGRES_CONN", ""),
	}

	return NewStorage(config)
}

func NewStorage(config Config) (Storage, error) {
	switch config.Type {
	case Memory:
		return NewMemoryStorage(), nil

	case Redis:
		if config.RedisAddr == "" {
			return nil, fmt.Errorf("Redis address is required")
		}
		return NewRedisStorage(config.RedisAddr, config.RedisPassword, config.RedisDB)

	case Postgres:
		if config.PostgresConnStr == "" {
			return nil, fmt.Errorf("Postgres connection string is required")
		}
		return NewPostgresStorage(config.PostgresConnStr)

	default:
		return nil, fmt.Errorf("unknown storage type: %s", config.Type)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		fmt.Sscanf(value, "%d", &result)
		return result
	}
	return defaultValue
}
