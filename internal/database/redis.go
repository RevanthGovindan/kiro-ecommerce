package database

import (
	"context"
	"fmt"
	"log"

	"ecommerce-website/internal/config"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitializeRedis sets up the Redis connection
func InitializeRedis(cfg *config.Config) error {
	// Parse Redis URL
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Create Redis client
	RedisClient = redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connection established successfully")
	return nil
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// GetRedisClient returns the Redis client instance
func GetRedisClient() *redis.Client {
	return RedisClient
}