package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ecommerce-website/internal/database"
	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	Requests int                       // Number of requests allowed
	Window   time.Duration             // Time window for the requests
	KeyFunc  func(*gin.Context) string // Function to generate rate limit key
}

// DefaultKeyFunc generates a rate limit key based on client IP
func DefaultKeyFunc(c *gin.Context) string {
	return fmt.Sprintf("rate_limit:%s", c.ClientIP())
}

// AuthenticatedUserKeyFunc generates a rate limit key based on user ID if authenticated, otherwise IP
func AuthenticatedUserKeyFunc(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("rate_limit:user:%s", userID)
	}
	return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := config.KeyFunc(c)

		// Get Redis client
		rdb := database.GetRedisClient()
		if rdb == nil {
			// If Redis is not available, allow the request to proceed
			c.Next()
			return
		}

		ctx := c.Request.Context()

		// Get current count
		current, err := rdb.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			// If Redis error, allow the request to proceed
			c.Next()
			return
		}

		var count int
		if current != "" {
			count, _ = strconv.Atoi(current)
		}

		// Check if rate limit exceeded
		if count >= config.Requests {
			// Get TTL to inform client when they can retry
			ttl, _ := rdb.TTL(ctx, key).Result()

			c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

			utils.ErrorResponse(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
				fmt.Sprintf("Rate limit exceeded. Try again in %v", ttl), nil)
			c.Abort()
			return
		}

		// Increment counter
		pipe := rdb.Pipeline()
		pipe.Incr(ctx, key)
		if count == 0 {
			// Set expiration only for the first request
			pipe.Expire(ctx, key, config.Window)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			// If Redis error, allow the request to proceed
			c.Next()
			return
		}

		// Set rate limit headers
		remaining := config.Requests - count - 1
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		c.Next()
	}
}

// Common rate limit configurations
var (
	// General API rate limit: 100 requests per minute
	GeneralRateLimit = RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
		KeyFunc:  DefaultKeyFunc,
	}

	// Authentication rate limit: 5 attempts per minute
	AuthRateLimit = RateLimitConfig{
		Requests: 5,
		Window:   time.Minute,
		KeyFunc:  DefaultKeyFunc,
	}

	// Admin operations rate limit: 50 requests per minute
	AdminRateLimit = RateLimitConfig{
		Requests: 50,
		Window:   time.Minute,
		KeyFunc:  AuthenticatedUserKeyFunc,
	}

	// Payment operations rate limit: 10 requests per minute
	PaymentRateLimit = RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		KeyFunc:  AuthenticatedUserKeyFunc,
	}
)
