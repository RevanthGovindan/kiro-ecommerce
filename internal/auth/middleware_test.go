package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce-website/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestMiddleware(t *testing.T) *Service {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}
	service := NewService(db, cfg)

	return service
}

func createTestToken(t *testing.T, userID, email, role string, expiry time.Duration) string {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	return tokenString
}

func TestAuthMiddleware(t *testing.T) {
	service := setupTestMiddleware(t)
	gin.SetMode(gin.TestMode)

	// Create a test handler that requires authentication
	testHandler := func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		userEmail, _ := c.Get("user_email")
		userRole, _ := c.Get("user_role")

		c.JSON(http.StatusOK, gin.H{
			"user_id":    userID,
			"user_email": userEmail,
			"user_role":  userRole,
		})
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + createTestToken(t, "user-123", "test@example.com", "customer", 15*time.Minute),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "MISSING_TOKEN",
		},
		{
			name:           "invalid authorization format",
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN_FORMAT",
		},
		{
			name:           "missing Bearer prefix",
			authHeader:     createTestToken(t, "user-123", "test@example.com", "customer", 15*time.Minute),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN_FORMAT",
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN",
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + createTestToken(t, "user-123", "test@example.com", "customer", -1*time.Hour),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			router.Use(service.AuthMiddleware())
			router.GET("/test", testHandler)

			c.Request = httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}

			router.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorDetail := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorDetail["code"])
			} else {
				// For successful authentication, check if user context is set
				assert.Equal(t, "user-123", response["user_id"])
				assert.Equal(t, "test@example.com", response["user_email"])
				assert.Equal(t, "customer", response["user_role"])
			}
		})
	}
}

func TestAdminMiddleware(t *testing.T) {
	service := setupTestMiddleware(t)
	gin.SetMode(gin.TestMode)

	// Create a test handler that requires admin access
	testHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	}

	tests := []struct {
		name           string
		userRole       string
		setContext     bool
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "admin user",
			userRole:       "admin",
			setContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "customer user",
			userRole:       "customer",
			setContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedError:  "INSUFFICIENT_PERMISSIONS",
		},
		{
			name:           "missing user context",
			setContext:     false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "MISSING_USER_CONTEXT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			// Setup middleware that sets user context
			router.Use(func(c *gin.Context) {
				if tt.setContext {
					c.Set("user_role", tt.userRole)
				}
				c.Next()
			})
			router.Use(service.AdminMiddleware())
			router.GET("/admin", testHandler)

			c.Request = httptest.NewRequest("GET", "/admin", nil)
			router.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorDetail := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorDetail["code"])
			} else {
				assert.Equal(t, "admin access granted", response["message"])
			}
		})
	}
}

func TestOptionalAuthMiddleware(t *testing.T) {
	service := setupTestMiddleware(t)
	gin.SetMode(gin.TestMode)

	// Create a test handler that works with or without authentication
	testHandler := func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if exists {
			c.JSON(http.StatusOK, gin.H{
				"authenticated": true,
				"user_id":       userID,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"authenticated": false,
			})
		}
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedAuth   bool
		expectedStatus int
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + createTestToken(t, "user-123", "test@example.com", "customer", 15*time.Minute),
			expectedAuth:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no authorization header",
			authHeader:     "",
			expectedAuth:   false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid token format",
			authHeader:     "InvalidFormat token",
			expectedAuth:   false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedAuth:   false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + createTestToken(t, "user-123", "test@example.com", "customer", -1*time.Hour),
			expectedAuth:   false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			router.Use(service.OptionalAuthMiddleware())
			router.GET("/test", testHandler)

			c.Request = httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}

			router.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedAuth, response["authenticated"])
			if tt.expectedAuth {
				assert.Equal(t, "user-123", response["user_id"])
			}
		})
	}
}