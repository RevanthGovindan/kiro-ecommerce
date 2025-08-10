package config

import (
	"os"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	JWTSecret      string
	RazorpayKeyID  string
	RazorpaySecret string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost:5433/ecommerce?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		RazorpayKeyID:  getEnv("RAZORPAY_KEY_ID", ""),
		RazorpaySecret: getEnv("RAZORPAY_SECRET", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
