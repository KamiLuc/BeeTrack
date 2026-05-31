package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	JWTAccessTTLMin   int
	JWTRefreshTTLDays int
	JWTSecret         string
	Port              string
}

func Load() (Config, error) {
	accessTTL, err := strconv.Atoi(getEnv("JWT_ACCESS_TTL_MIN", "15"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_ACCESS_TTL_MIN: %w", err)
	}

	refreshTTL, err := strconv.Atoi(getEnv("JWT_REFRESH_TTL_DAYS", "30"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_REFRESH_TTL_DAYS: %w", err)
	}

	return Config{
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		JWTAccessTTLMin:   accessTTL,
		JWTRefreshTTLDays: refreshTTL,
		JWTSecret:         getEnv("JWT_SECRET", ""),
		Port:              getEnv("PORT", "8080"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
