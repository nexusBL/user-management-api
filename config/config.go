package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv           string
	ServerPort       string
	DatabaseURL      string
	DefaultPageLimit int32
	MaxPageLimit     int32
	ShutdownTimeout  time.Duration
}

func Load() (Config, error) {
	defaultPageLimit, err := getEnvInt32("DEFAULT_PAGE_LIMIT", 10)
	if err != nil {
		return Config{}, err
	}

	maxPageLimit, err := getEnvInt32("MAX_PAGE_LIMIT", 50)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := getEnvDuration("SHUTDOWN_TIMEOUT", "10s")
	if err != nil {
		return Config{}, err
	}

	if defaultPageLimit <= 0 {
		return Config{}, fmt.Errorf("DEFAULT_PAGE_LIMIT must be greater than zero")
	}

	if maxPageLimit < defaultPageLimit {
		return Config{}, fmt.Errorf("MAX_PAGE_LIMIT must be greater than or equal to DEFAULT_PAGE_LIMIT")
	}

	return Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		ServerPort:       getEnv("SERVER_PORT", "3000"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/user_management?sslmode=disable"),
		DefaultPageLimit: defaultPageLimit,
		MaxPageLimit:     maxPageLimit,
		ShutdownTimeout:  shutdownTimeout,
	}, nil
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvInt32(key string, fallback int32) (int32, error) {
	value := getEnv(key, strconv.FormatInt(int64(fallback), 10))
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}

	return int32(parsed), nil
}

func getEnvDuration(key string, fallback string) (time.Duration, error) {
	value := getEnv(key, fallback)
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}

	return duration, nil
}
