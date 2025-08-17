package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	JWTSecret      string
	RazorpayKeyID  string
	RazorpaySecret string
	SMTPHost       string
	SMTPPort       string
	SMTPUsername   string
	SMTPPassword   string
	FromEmail      string
	CDNBaseURL     string
	MaxRequestSize int64
	Environment    string
	AdminEmail     string
	AdminPassword  string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost:5433/ecommerce?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		RazorpayKeyID:  getEnv("RAZORPAY_KEY_ID", ""),
		RazorpaySecret: getEnv("RAZORPAY_SECRET", ""),
		SMTPHost:       getEnv("SMTP_HOST", ""),
		SMTPPort:       getEnv("SMTP_PORT", "587"),
		SMTPUsername:   getEnv("SMTP_USERNAME", ""),
		SMTPPassword:   getEnv("SMTP_PASSWORD", ""),
		FromEmail:      getEnv("FROM_EMAIL", ""),
		CDNBaseURL:     getEnv("CDN_BASE_URL", ""),
		MaxRequestSize: getEnvInt64("MAX_REQUEST_SIZE", 10*1024*1024), // 10MB default
		Environment:    getEnv("ENVIRONMENT", "development"),
		AdminEmail:     getEnv("ADMIN_EMAIL", "admin@ecommerce.com"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", "admin123456"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}
