package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AllowedOrigins    string
	APIURL            string
	AppURL            string
	DatabaseURL       string
	ImageStoragePath  string
	JWTAccessTTLMin   int
	JWTRefreshTTLDays int
	JWTSecret         string
	Port              string
	SMTPFrom          string
	SMTPHost          string
	SMTPPass          string
	SMTPPort          string
	SMTPUser          string
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
		AllowedOrigins:    getEnv("CORS_ALLOWED_ORIGINS", "*"),
		APIURL:            getEnv("API_URL", "http://localhost:8080"),
		AppURL:            getEnv("APP_URL", "http://localhost:5000"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		ImageStoragePath:  getEnv("IMAGE_STORAGE_PATH", "/data/images"),
		JWTAccessTTLMin:   accessTTL,
		JWTRefreshTTLDays: refreshTTL,
		JWTSecret:         getEnv("JWT_SECRET", ""),
		Port:              getEnv("PORT", "8080"),
		SMTPFrom:          getEnv("SMTP_FROM", "noreply@beetrack.app"),
		SMTPHost:          getEnv("SMTP_HOST", "localhost"),
		SMTPPass:          getEnv("SMTP_PASS", ""),
		SMTPPort:          getEnv("SMTP_PORT", "1025"),
		SMTPUser:          getEnv("SMTP_USER", ""),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
