package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-website/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupIntegrationTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate the schema using TestUser for SQLite compatibility
	err = db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup authentication
	cfg := &config.Config{JWTSecret: "test-secret-key"}
	authService := NewService(db, cfg)
	authHandler := NewHandler(authService)

	// Setup routes
	SetupRoutes(router, authHandler, authService)

	return router, db
}

func TestAuthenticationFlow_Integration(t *testing.T) {
	router, _ := setupIntegrationTest(t)

	// Test user registration
	registerReq := RegisterRequest{
		Email:     "integration@test.com",
		Password:  "password123",
		FirstName: "Integration",
		LastName:  "Test",
		Phone:     "1234567890",
	}

	registerBody, _ := json.Marshal(registerReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var registerResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &registerResponse)
	require.NoError(t, err)
	assert.True(t, registerResponse["success"].(bool))

	// Test user login
	loginReq := LoginRequest{
		Email:    "integration@test.com",
		Password: "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)
	assert.True(t, loginResponse["success"].(bool))

	// Extract tokens
	data := loginResponse["data"].(map[string]interface{})
	tokens := data["tokens"].(map[string]interface{})
	accessToken := tokens["access_token"].(string)
	refreshToken := tokens["refresh_token"].(string)

	// Test authenticated endpoint
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var meResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &meResponse)
	require.NoError(t, err)
	assert.True(t, meResponse["success"].(bool))

	// Test token refresh
	refreshReq := map[string]string{
		"refresh_token": refreshToken,
	}

	refreshBody, _ := json.Marshal(refreshReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var refreshResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &refreshResponse)
	require.NoError(t, err)
	assert.True(t, refreshResponse["success"].(bool))

	// Test logout
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/auth/logout", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var logoutResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &logoutResponse)
	require.NoError(t, err)
	assert.True(t, logoutResponse["success"].(bool))
}

func TestPasswordResetFlow_Integration(t *testing.T) {
	router, db := setupIntegrationTest(t)

	// Create a test user first
	user := TestUser{
		ID:        "reset-user-id",
		Email:     "reset@test.com",
		Password:  "hashedpassword",
		FirstName: "Reset",
		LastName:  "Test",
		Role:      "customer",
		IsActive:  true,
	}
	db.Create(&user)

	// Test forgot password
	forgotReq := ForgotPasswordRequest{
		Email: "reset@test.com",
	}

	forgotBody, _ := json.Marshal(forgotReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/forgot-password", bytes.NewBuffer(forgotBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var forgotResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &forgotResponse)
	require.NoError(t, err)
	assert.True(t, forgotResponse["success"].(bool))

	// Get the reset token from database (in real scenario, this would be sent via email)
	var updatedUser TestUser
	db.Where("email = ?", "reset@test.com").First(&updatedUser)
	require.NotNil(t, updatedUser.PasswordResetToken)

	// Test password reset
	resetReq := ResetPasswordRequest{
		Token:    *updatedUser.PasswordResetToken,
		Password: "newpassword123",
	}

	resetBody, _ := json.Marshal(resetReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/auth/reset-password", bytes.NewBuffer(resetBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resetResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resetResponse)
	require.NoError(t, err)
	assert.True(t, resetResponse["success"].(bool))

	// Verify user can login with new password
	loginReq := LoginRequest{
		Email:    "reset@test.com",
		Password: "newpassword123",
	}

	loginBody, _ := json.Marshal(loginReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)
	assert.True(t, loginResponse["success"].(bool))
}