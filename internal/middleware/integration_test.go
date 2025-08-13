package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	// Apply security middleware
	r.Use(SecurityHeadersMiddleware())
	r.Use(InputSanitizationMiddleware())
	r.Use(ValidateContentLength(1024))

	r.POST("/api/test", func(c *gin.Context) {
		body, _ := c.GetRawData()
		c.JSON(200, gin.H{
			"message": "success",
			"body":    string(body),
		})
	})

	// Test 1: Normal request should work with security headers
	normalJSON := `{"name": "John Doe", "email": "john@example.com"}`
	req, _ := http.NewRequest("POST", "/api/test", bytes.NewBufferString(normalJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")

	// Test 2: Request with null bytes should be sanitized
	maliciousJSON := "{\x00\"name\": \"test\x00user\"}"
	req2, _ := http.NewRequest("POST", "/api/test", bytes.NewBufferString(maliciousJSON))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w2.Code)
	// Verify null bytes are removed
	assert.NotContains(t, w2.Body.String(), "\x00")

	// Test 3: Request too large should be rejected
	largeContent := make([]byte, 2048) // Larger than 1024 limit
	for i := range largeContent {
		largeContent[i] = 'A'
	}

	req3, _ := http.NewRequest("POST", "/api/test", bytes.NewBuffer(largeContent))
	req3.Header.Set("Content-Type", "application/json")
	req3.ContentLength = int64(len(largeContent))

	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	assert.Equal(t, 413, w3.Code) // Request Entity Too Large
}

func TestValidationFunctionsIntegration(t *testing.T) {
	// Test comprehensive validation scenarios
	testCases := []struct {
		name     string
		function func() bool
		expected bool
	}{
		{
			name: "Valid email",
			function: func() bool {
				return ValidateEmail("user@example.com")
			},
			expected: true,
		},
		{
			name: "Invalid email",
			function: func() bool {
				return ValidateEmail("invalid-email")
			},
			expected: false,
		},
		{
			name: "Strong password",
			function: func() bool {
				valid, _ := ValidatePassword("StrongPass123!")
				return valid
			},
			expected: true,
		},
		{
			name: "Weak password",
			function: func() bool {
				valid, _ := ValidatePassword("weak")
				return valid
			},
			expected: false,
		},
		{
			name: "Valid phone",
			function: func() bool {
				return ValidatePhone("+1-234-567-8900")
			},
			expected: true,
		},
		{
			name: "Invalid phone",
			function: func() bool {
				return ValidatePhone("123")
			},
			expected: false,
		},
		{
			name: "Valid price",
			function: func() bool {
				return ValidatePrice(99.99)
			},
			expected: true,
		},
		{
			name: "Invalid price",
			function: func() bool {
				return ValidatePrice(-10.00)
			},
			expected: false,
		},
		{
			name: "Valid quantity",
			function: func() bool {
				return ValidateQuantity(5)
			},
			expected: true,
		},
		{
			name: "Invalid quantity",
			function: func() bool {
				return ValidateQuantity(-1)
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.function()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSanitizationIntegration(t *testing.T) {
	// Test various XSS and injection attempts
	maliciousInputs := []struct {
		input            string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			input:            "<script>alert('xss')</script>",
			shouldContain:    []string{"&lt;script&gt;", "&lt;/script&gt;"},
			shouldNotContain: []string{"<script>", "</script>"},
		},
		{
			input:            "javascript:alert('hack')",
			shouldContain:    []string{"alert"},
			shouldNotContain: []string{"javascript:"},
		},
		{
			input:            "<img src=x onerror=alert('xss')>",
			shouldContain:    []string{"&lt;img", "&gt;"},
			shouldNotContain: []string{"<img", "onerror="},
		},
		{
			input:            "Normal text with <b>bold</b>",
			shouldContain:    []string{"Normal text", "&lt;b&gt;", "&lt;/b&gt;"},
			shouldNotContain: []string{"<b>", "</b>"},
		},
	}

	for i, test := range maliciousInputs {
		t.Run(fmt.Sprintf("Sanitization test %d", i+1), func(t *testing.T) {
			result := SanitizeInput(test.input)

			for _, should := range test.shouldContain {
				assert.Contains(t, result, should, "Result should contain: %s", should)
			}

			for _, shouldNot := range test.shouldNotContain {
				assert.NotContains(t, result, shouldNot, "Result should not contain: %s", shouldNot)
			}
		})
	}
}
