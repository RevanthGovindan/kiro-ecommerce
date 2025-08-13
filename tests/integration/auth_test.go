//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ecommerce-website/internal/auth"
	"ecommerce-website/internal/models"
)

func setupAuthTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, func()) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup test database
	db, cleanup := setupTestDB(t)

	// Create auth service
	authService := auth.NewService(db, "test-jwt-secret")

	// Setup router
	router := gin.New()
	authGroup := router.Group("/api/auth")
	auth.RegisterRoutes(authGroup, authService)

	return router, db, cleanup
}

func TestAuthIntegration_Register(t *testing.T) {
	router, db, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		checkDB        bool
	}{
		{
			name: "successful registration",
			payload: map[string]interface{}{
				"email":     "test@example.com",
				"password":  "password123",
				"firstName": "John",
				"lastName":  "Doe",
			},
			expectedStatus: http.StatusCreated,
			checkDB:        true,
		},
		{
			name: "duplicate email",
			payload: map[string]interface{}{
				"email":     "test@example.com", // Same email as above
				"password":  "password123",
				"firstName": "Jane",
				"lastName":  "Doe",
			},
			expectedStatus: http.StatusConflict,
			checkDB:        false,
		},
		{
			name: "invalid email",
			payload: map[string]interface{}{
				"email":     "invalid-email",
				"password":  "password123",
				"firstName": "John",
				"lastName":  "Doe",
			},
			expectedStatus: http.StatusBadRequest,
			checkDB:        false,
		},
		{
			name: "password too short",
			payload: map[string]interface{}{
				"email":     "short@example.com",
				"password":  "123",
				"firstName": "John",
				"lastName":  "Doe",
			},
			expectedStatus: http.StatusBadRequest,
			checkDB:        false,
		},
		{
			name: "missing required fields",
			payload: map[string]interface{}{
				"email":    "missing@example.com",
				"password": "password123",
				// Missing firstName and lastName
			},
			expectedStatus: http.StatusBadRequest,
			checkDB:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkDB {
				// Verify user was created in database
				var user models.User
				err := db.Where("email = ?", tt.payload["email"]).First(&user).Error
				require.NoError(t, err)
				assert.Equal(t, tt.payload["email"], user.Email)
				assert.Equal(t, tt.payload["firstName"], user.FirstName)
				assert.Equal(t, tt.payload["lastName"], user.LastName)
				assert.NotEmpty(t, user.ID)
				assert.True(t, user.IsActive)
			}

			// Check response structure
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusCreated {
				assert.True(t, response["success"].(bool))
				assert.NotNil(t, response["data"])
				userData := response["data"].(map[string]interface{})
				assert.NotEmpty(t, userData["token"])
				assert.NotNil(t, userData["user"])
			} else {
				assert.False(t, response["success"].(bool))
				assert.NotNil(t, response["error"])
			}
		})
	}
}

func TestAuthIntegration_Login(t *testing.T) {
	router, db, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	// Create a test user
	testUser := &models.User{
		Email:     "login@example.com",
		Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
	}
	require.NoError(t, db.Create(testUser).Error)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "successful login",
			payload: map[string]interface{}{
				"email":    "login@example.com",
				"password": "password",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid email",
			payload: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid password",
			payload: map[string]interface{}{
				"email":    "login@example.com",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "missing email",
			payload: map[string]interface{}{
				"password": "password",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			payload: map[string]interface{}{
				"email": "login@example.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response structure
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response["success"].(bool))
				assert.NotNil(t, response["data"])
				userData := response["data"].(map[string]interface{})
				assert.NotEmpty(t, userData["token"])
				assert.NotNil(t, userData["user"])
			} else {
				assert.False(t, response["success"].(bool))
				assert.NotNil(t, response["error"])
			}
		})
	}
}

func TestAuthIntegration_ProtectedRoute(t *testing.T) {
	router, db, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	// Add a protected route for testing
	authService := auth.NewService(db, "test-jwt-secret")
	protected := router.Group("/api/protected")
	protected.Use(auth.AuthMiddleware(authService))
	protected.GET("/profile", func(c *gin.Context) {
		userID := c.GetString("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	// Create a test user and get token
	testUser := &models.User{
		Email:     "protected@example.com",
		Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
	}
	require.NoError(t, db.Create(testUser).Error)

	token, err := authService.GenerateToken(testUser)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "valid token",
			token:          token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			token:          "invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/protected/profile", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, testUser.ID, response["user_id"])
			}
		})
	}
}
