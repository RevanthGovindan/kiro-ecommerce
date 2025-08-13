package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSecurityHeadersMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(SecurityHeadersMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkInputSanitizationMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(InputSanitizationMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	jsonData := `{"name": "test user", "email": "test@example.com"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkSanitizeInput(b *testing.B) {
	input := "<script>alert('xss')</script>Normal text with some HTML <img src='test.jpg'>"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeInput(input)
	}
}

func BenchmarkValidateEmail(b *testing.B) {
	email := "test.user+tag@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateEmail(email)
	}
}

func BenchmarkValidatePassword(b *testing.B) {
	password := "SecurePassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidatePassword(password)
	}
}

func TestRateLimitMiddlewarePerformance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test without rate limiting
	r1 := gin.New()
	r1.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	// Test with rate limiting (but without Redis for this test)
	r2 := gin.New()
	r2.Use(RateLimitMiddleware(RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
		KeyFunc:  DefaultKeyFunc,
	}))
	r2.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)

	// Benchmark without rate limiting
	start := time.Now()
	for i := 0; i < 1000; i++ {
		w := httptest.NewRecorder()
		r1.ServeHTTP(w, req)
	}
	withoutRateLimit := time.Since(start)

	// Benchmark with rate limiting
	start = time.Now()
	for i := 0; i < 1000; i++ {
		w := httptest.NewRecorder()
		r2.ServeHTTP(w, req)
	}
	withRateLimit := time.Since(start)

	// Rate limiting should not add excessive overhead when Redis is not available
	// Allow up to 200% overhead since we're doing Redis connection attempts
	overhead := float64(withRateLimit-withoutRateLimit) / float64(withoutRateLimit)
	assert.Less(t, overhead, 2.0, "Rate limiting middleware adds too much overhead: %f", overhead)

	t.Logf("Without rate limiting: %v", withoutRateLimit)
	t.Logf("With rate limiting: %v", withRateLimit)
	t.Logf("Overhead: %.2f%%", overhead*100)
}

func TestCacheMiddlewarePerformance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test without caching
	r1 := gin.New()
	r1.GET("/test", func(c *gin.Context) {
		// Simulate some processing time
		time.Sleep(1 * time.Millisecond)
		c.JSON(200, gin.H{"message": "test", "timestamp": time.Now().Unix()})
	})

	// Test with caching (but without Redis for this test)
	r2 := gin.New()
	r2.Use(CacheMiddleware(CacheConfig{
		TTL:     5 * time.Minute,
		KeyFunc: DefaultCacheKeyFunc,
	}))
	r2.GET("/test", func(c *gin.Context) {
		// Simulate some processing time
		time.Sleep(1 * time.Millisecond)
		c.JSON(200, gin.H{"message": "test", "timestamp": time.Now().Unix()})
	})

	req, _ := http.NewRequest("GET", "/test", nil)

	// Benchmark without caching
	start := time.Now()
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		r1.ServeHTTP(w, req)
	}
	withoutCache := time.Since(start)

	// Benchmark with caching
	start = time.Now()
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		r2.ServeHTTP(w, req)
	}
	withCache := time.Since(start)

	t.Logf("Without caching: %v", withoutCache)
	t.Logf("With caching: %v", withCache)

	// Cache middleware should not add excessive overhead when Redis is not available
	// Allow up to 100% overhead since we're doing Redis connection attempts
	overhead := float64(withCache-withoutCache) / float64(withoutCache)
	assert.Less(t, overhead, 1.0, "Cache middleware adds too much overhead: %f", overhead)
}

func BenchmarkMultipleMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(SecurityHeadersMiddleware())
	r.Use(InputSanitizationMiddleware())
	r.Use(ValidateContentLength(1024 * 1024)) // 1MB
	r.Use(RateLimitMiddleware(GeneralRateLimit))
	r.Use(CacheMiddleware(ProductCatalogCache))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
