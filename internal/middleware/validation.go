package middleware

import (
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"

	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
)

// SanitizeInput sanitizes string input to prevent XSS attacks
func SanitizeInput(input string) string {
	// HTML escape the input
	sanitized := html.EscapeString(input)

	// Remove potentially dangerous characters
	sanitized = strings.ReplaceAll(sanitized, "<script>", "")
	sanitized = strings.ReplaceAll(sanitized, "</script>", "")
	sanitized = strings.ReplaceAll(sanitized, "javascript:", "")
	sanitized = strings.ReplaceAll(sanitized, "vbscript:", "")
	sanitized = strings.ReplaceAll(sanitized, "onload=", "")
	sanitized = strings.ReplaceAll(sanitized, "onerror=", "")

	return strings.TrimSpace(sanitized)
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePassword validates password strength
func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters long"
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	if !hasUpper {
		return false, "Password must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "Password must contain at least one lowercase letter"
	}
	if !hasNumber {
		return false, "Password must contain at least one number"
	}
	if !hasSpecial {
		return false, "Password must contain at least one special character"
	}

	return true, ""
}

// ValidatePhone validates phone number format
func ValidatePhone(phone string) bool {
	// Remove all non-digit characters
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	// Check if it's a valid length (10-15 digits)
	return len(digits) >= 10 && len(digits) <= 15
}

// ValidatePrice validates price format
func ValidatePrice(price float64) bool {
	return price >= 0 && price <= 999999.99
}

// ValidateQuantity validates quantity
func ValidateQuantity(quantity int) bool {
	return quantity >= 0 && quantity <= 10000
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict transport security (HTTPS only)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content security policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://checkout.razorpay.com; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self' https:; frame-src https://api.razorpay.com")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// InputSanitizationMiddleware sanitizes request body inputs
func InputSanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only process JSON requests
		if c.GetHeader("Content-Type") == "application/json" {
			// Get the raw body
			body, err := c.GetRawData()
			if err != nil {
				utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Failed to read request body", nil)
				c.Abort()
				return
			}

			// Basic sanitization - remove null bytes and control characters
			sanitizedBody := strings.ReplaceAll(string(body), "\x00", "")
			sanitizedBody = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(sanitizedBody, "")

			// Set the sanitized body back
			c.Request.Body = io.NopCloser(strings.NewReader(sanitizedBody))
		}

		c.Next()
	}
}

// ValidateContentLength validates request content length
func ValidateContentLength(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			utils.ErrorResponse(c, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE",
				"Request body too large", gin.H{"maxSize": maxSize})
			c.Abort()
			return
		}
		c.Next()
	}
}
