package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(SecurityHeadersMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=31536000")
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func TestInputSanitizationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(InputSanitizationMiddleware())
	r.POST("/test", func(c *gin.Context) {
		body, _ := c.GetRawData()
		c.JSON(200, gin.H{"body": string(body)})
	})

	// Test with malicious input containing null bytes
	maliciousJSON := `{"name": "test\x00user", "script": "<script>alert('xss')</script>"}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(maliciousJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// Verify null bytes are removed
	assert.NotContains(t, w.Body.String(), "\x00")
}

func TestValidateContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(ValidateContentLength(10)) // 10 bytes max
	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Test with content that exceeds limit
	largeContent := "This content is definitely longer than 10 bytes"
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(largeContent))
	req.ContentLength = int64(len(largeContent))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 413, w.Code) // Request Entity Too Large
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			input:    "javascript:alert('xss')",
			expected: "alert(&#39;xss&#39;)",
		},
		{
			input:    "Normal text",
			expected: "Normal text",
		},
		{
			input:    "<img src=x onerror=alert('xss')>",
			expected: "&lt;img src=x alert(&#39;xss&#39;)&gt;",
		},
	}

	for _, test := range tests {
		result := SanitizeInput(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name+tag@domain.co.uk", true},
		{"invalid-email", false},
		{"@domain.com", false},
		{"user@", false},
		{"", false},
	}

	for _, test := range tests {
		result := ValidateEmail(test.email)
		assert.Equal(t, test.valid, result, "Email: %s", test.email)
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
		message  string
	}{
		{"Password123!", true, ""},
		{"weak", false, "Password must be at least 8 characters long"},
		{"password123!", false, "Password must contain at least one uppercase letter"},
		{"PASSWORD123!", false, "Password must contain at least one lowercase letter"},
		{"Password!", false, "Password must contain at least one number"},
		{"Password123", false, "Password must contain at least one special character"},
	}

	for _, test := range tests {
		valid, message := ValidatePassword(test.password)
		assert.Equal(t, test.valid, valid, "Password: %s", test.password)
		if !test.valid {
			assert.Equal(t, test.message, message)
		}
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		phone string
		valid bool
	}{
		{"1234567890", true},
		{"+1-234-567-8900", true},
		{"(123) 456-7890", true},
		{"123", false},
		{"12345678901234567890", false}, // too long
		{"", false},
	}

	for _, test := range tests {
		result := ValidatePhone(test.phone)
		assert.Equal(t, test.valid, result, "Phone: %s", test.phone)
	}
}

func TestValidatePrice(t *testing.T) {
	tests := []struct {
		price float64
		valid bool
	}{
		{10.99, true},
		{0, true},
		{999999.99, true},
		{-1, false},
		{1000000, false},
	}

	for _, test := range tests {
		result := ValidatePrice(test.price)
		assert.Equal(t, test.valid, result, "Price: %f", test.price)
	}
}

func TestValidateQuantity(t *testing.T) {
	tests := []struct {
		quantity int
		valid    bool
	}{
		{1, true},
		{0, true},
		{10000, true},
		{-1, false},
		{10001, false},
	}

	for _, test := range tests {
		result := ValidateQuantity(test.quantity)
		assert.Equal(t, test.valid, result, "Quantity: %d", test.quantity)
	}
}
