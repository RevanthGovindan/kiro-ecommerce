package middleware

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ecommerce-website/internal/database"

	"github.com/gin-gonic/gin"
)

// CacheConfig defines caching configuration
type CacheConfig struct {
	TTL     time.Duration
	KeyFunc func(*gin.Context) string
}

// DefaultCacheKeyFunc generates a cache key based on request path and query parameters
func DefaultCacheKeyFunc(c *gin.Context) string {
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery

	// Include user ID in cache key if authenticated
	if userID, exists := c.Get("userID"); exists {
		return fmt.Sprintf("cache:user:%s:%s:%s", userID, path, query)
	}

	return fmt.Sprintf("cache:public:%s:%s", path, query)
}

// ProductCacheKeyFunc generates cache key for product-related requests
func ProductCacheKeyFunc(c *gin.Context) string {
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery
	hash := fmt.Sprintf("%x", md5.Sum([]byte(path+query)))
	return fmt.Sprintf("cache:products:%s", hash)
}

// CacheMiddleware creates a caching middleware
func CacheMiddleware(config CacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		rdb := database.GetRedisClient()
		if rdb == nil {
			c.Next()
			return
		}

		key := config.KeyFunc(c)
		ctx := c.Request.Context()

		// Try to get cached response
		cached, err := rdb.Get(ctx, key).Result()
		if err == nil {
			// Cache hit - return cached response
			var cachedResponse CachedResponse
			if json.Unmarshal([]byte(cached), &cachedResponse) == nil {
				// Set headers
				for k, v := range cachedResponse.Headers {
					c.Header(k, v)
				}
				c.Header("X-Cache", "HIT")
				c.Data(cachedResponse.StatusCode, cachedResponse.ContentType, cachedResponse.Body)
				c.Abort()
				return
			}
		}

		// Cache miss - continue with request and cache the response
		c.Header("X-Cache", "MISS")

		// Create a custom response writer to capture the response
		writer := &CacheResponseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
			headers:        make(map[string]string),
		}
		c.Writer = writer

		c.Next()

		// Cache the response if it's successful
		if writer.statusCode >= 200 && writer.statusCode < 300 {
			cachedResponse := CachedResponse{
				StatusCode:  writer.statusCode,
				ContentType: writer.Header().Get("Content-Type"),
				Headers:     writer.headers,
				Body:        writer.body,
			}

			if data, err := json.Marshal(cachedResponse); err == nil {
				rdb.Set(ctx, key, data, config.TTL)
			}
		}
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode  int               `json:"statusCode"`
	ContentType string            `json:"contentType"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
}

// CacheResponseWriter wraps gin.ResponseWriter to capture response data
type CacheResponseWriter struct {
	gin.ResponseWriter
	body       []byte
	headers    map[string]string
	statusCode int
}

func (w *CacheResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

func (w *CacheResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *CacheResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// InvalidateCache invalidates cache entries matching a pattern
func InvalidateCache(pattern string) error {
	rdb := database.GetRedisClient()
	if rdb == nil {
		return fmt.Errorf("redis not available")
	}

	ctx := context.Background()
	keys, err := rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return rdb.Del(ctx, keys...).Err()
	}

	return nil
}

// Common cache configurations
var (
	// Product catalog cache: 5 minutes
	ProductCatalogCache = CacheConfig{
		TTL:     5 * time.Minute,
		KeyFunc: ProductCacheKeyFunc,
	}

	// Category cache: 15 minutes
	CategoryCache = CacheConfig{
		TTL:     15 * time.Minute,
		KeyFunc: DefaultCacheKeyFunc,
	}

	// User profile cache: 2 minutes
	UserProfileCache = CacheConfig{
		TTL:     2 * time.Minute,
		KeyFunc: DefaultCacheKeyFunc,
	}

	// Search results cache: 10 minutes
	SearchCache = CacheConfig{
		TTL:     10 * time.Minute,
		KeyFunc: DefaultCacheKeyFunc,
	}
)

// CacheInvalidationMiddleware invalidates relevant caches after write operations
func CacheInvalidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only invalidate cache for successful write operations
		if c.Request.Method != "GET" && c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			path := c.Request.URL.Path

			// Invalidate product-related caches
			if contains(path, []string{"/products", "/categories"}) {
				go InvalidateCache("cache:products:*")
				go InvalidateCache("cache:public:/api/products*")
				go InvalidateCache("cache:public:/api/categories*")
			}

			// Invalidate user-related caches
			if contains(path, []string{"/users", "/profile"}) {
				if userID, exists := c.Get("userID"); exists {
					go InvalidateCache(fmt.Sprintf("cache:user:%s:*", userID))
				}
			}

			// Invalidate order-related caches
			if contains(path, []string{"/orders"}) {
				if userID, exists := c.Get("userID"); exists {
					go InvalidateCache(fmt.Sprintf("cache:user:%s:/api/orders*", userID))
					go InvalidateCache(fmt.Sprintf("cache:user:%s:/api/users/orders*", userID))
				}
			}
		}
	}
}

func contains(str string, substrings []string) bool {
	for _, substring := range substrings {
		if len(str) >= len(substring) && str[:len(substring)] == substring {
			return true
		}
	}
	return false
}
