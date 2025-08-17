package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce-website/internal/config"
	"ecommerce-website/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHandler_Register(t *testing.T) {
	handler, _ := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful registration",
			requestBody: RegisterRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
				Phone:     "1234567890",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid email",
			requestBody: RegisterRequest{
				Email:     "invalid-email",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "missing required fields",
			requestBody: RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				// Missing FirstName and LastName
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "password too short",
			requestBody: RegisterRequest{
				Email:     "test@example.com",
				Password:  "short",
				FirstName: "John",
				LastName:  "Doe",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.Register(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorDetail := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorDetail["code"])
			} else {
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "user")
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	handler, db := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := TestUser{
		ID:        "test-user-id",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		IsActive:  true,
	}
	db.Create(&testUser)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful login",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid email",
			requestBody: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_CREDENTIALS",
		},
		{
			name: "invalid password",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_CREDENTIALS",
		},
		{
			name: "missing email",
			requestBody: LoginRequest{
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.Login(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorDetail := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorDetail["code"])
			} else {
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "user")
				assert.Contains(t, data, "tokens")
			}
		})
	}
}

func TestHandler_RefreshToken(t *testing.T) {
	handler, db := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Create test user
	testUser := TestUser{
		ID:        "test-user-id",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		IsActive:  true,
	}
	db.Create(&testUser)

	// Generate valid refresh token using models.User
	cfg := &config.Config{JWTSecret: "test-secret-key"}
	service := NewService(db, cfg)
	modelUser := &models.User{
		ID:    testUser.ID,
		Email: testUser.Email,
		Role:  testUser.Role,
	}
	tokens, err := service.GenerateTokens(modelUser)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful token refresh",
			requestBody: map[string]string{
				"refresh_token": tokens.RefreshToken,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid refresh token",
			requestBody: map[string]string{
				"refresh_token": "invalid-token",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN",
		},
		{
			name:        "missing refresh token",
			requestBody: map[string]string{
				// Missing refresh_token
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.RefreshToken(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorDetail := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorDetail["code"])
			} else {
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "tokens")
			}
		})
	}
}

func TestHandler_Logout(t *testing.T) {
	handler, _ := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/auth/logout", nil)

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.Contains(t, response, "data")
}

func TestHandler_Me(t *testing.T) {
	handler, db := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Create test user
	testUser := TestUser{
		ID:        "test-user-id",
		Email:     "test@example.com",
		Password:  "hashed-password",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		IsActive:  true,
	}
	db.Create(&testUser)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful user retrieval",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", "test-user-id")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing user context",
			setupContext: func(c *gin.Context) {
				// Don't set user_id
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "MISSING_USER_CONTEXT",
		},
		{
			name: "user not found",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", "non-existent-id")
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "USER_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/auth/me", nil)

			tt.setupContext(c)

			handler.Me(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorDetail := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorDetail["code"])
			} else {
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "data")
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "user")
			}
		})
	}
}

func TestHandler_ForgotPassword(t *testing.T) {
	handler, db := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Create test user
	testUser := TestUser{
		ID:        "test-user-id",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		IsActive:  true,
	}
	db.Create(&testUser)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful forgot password",
			requestBody: ForgotPasswordRequest{
				Email: "test@example.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "non-existent email",
			requestBody: ForgotPasswordRequest{
				Email: "nonexistent@example.com",
			},
			expectedStatus: http.StatusOK, // Should not reveal if email exists
		},
		{
			name: "invalid email format",
			requestBody: ForgotPasswordRequest{
				Email: "invalid-email",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:        "missing email",
			requestBody: map[string]string{
				// Missing email
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/api/auth/forgot-password", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.ForgotPassword(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorDetail := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorDetail["code"])
			} else {
				assert.True(t, response["success"].(bool))
			}
		})
	}
}

func TestHandler_ResetPassword(t *testing.T) {
	handler, db := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Create test user
	testUser := TestUser{
		ID:        "test-user-id",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		IsActive:  true,
	}
	db.Create(&testUser)

	// Generate valid reset token
	resetClaims := Claims{
		UserID: testUser.ID,
		Email:  testUser.Email,
		Role:   "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   testUser.ID,
		},
	}

	resetToken := jwt.NewWithClaims(jwt.SigningMethodHS256, resetClaims)
	resetTokenString, err := resetToken.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	// Update user with reset token
	db.Exec("UPDATE users SET password_reset_token = ?, password_reset_expiry = ? WHERE id = ?",
		resetTokenString, time.Now().Add(1*time.Hour), testUser.ID)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful password reset",
			requestBody: ResetPasswordRequest{
				Token:    resetTokenString,
				Password: "newpassword123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid token",
			requestBody: ResetPasswordRequest{
				Token:    "invalid-token",
				Password: "newpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_TOKEN",
		},
		{
			name: "missing token",
			requestBody: ResetPasswordRequest{
				Password: "newpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "password too short",
			requestBody: ResetPasswordRequest{
				Token:    resetTokenString,
				Password: "short",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/api/auth/reset-password", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.ResetPassword(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorDetail := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorDetail["code"])
			} else {
				assert.True(t, response["success"].(bool))
			}
		})
	}
}

func TestHandler_VerifyEmail(t *testing.T) {
	handler, db := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Create test user
	testUser := TestUser{
		ID:        "test-user-id",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		IsActive:  true,
	}
	db.Create(&testUser)

	// Generate valid verification token
	verificationClaims := Claims{
		UserID: testUser.ID,
		Email:  testUser.Email,
		Role:   "email_verification",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   testUser.ID,
		},
	}

	verificationToken := jwt.NewWithClaims(jwt.SigningMethodHS256, verificationClaims)
	verificationTokenString, err := verificationToken.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	// Update user with verification token
	db.Exec("UPDATE users SET email_verification_token = ? WHERE id = ?",
		verificationTokenString, testUser.ID)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful email verification",
			token:          verificationTokenString,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_TOKEN",
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			url := "/api/auth/verify-email"
			if tt.token != "" {
				url += "?token=" + tt.token
			}
			c.Request = httptest.NewRequest("GET", url, nil)

			handler.VerifyEmail(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorDetail := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorDetail["code"])
			} else {
				assert.True(t, response["success"].(bool))
			}
		})
	}
}

func TestHandler_ResendEmailVerification(t *testing.T) {
	handler, db := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Create test user
	testUser := TestUser{
		ID:        "test-user-id",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		IsActive:  true,
	}
	db.Create(&testUser)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful resend verification",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", "test-user-id")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing user context",
			setupContext: func(c *gin.Context) {
				// Don't set user_id
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "MISSING_USER_CONTEXT",
		},
		{
			name: "user not found",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", "non-existent-id")
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "USER_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/api/auth/resend-verification", nil)

			tt.setupContext(c)

			handler.ResendEmailVerification(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorDetail := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorDetail["code"])
			} else {
				assert.True(t, response["success"].(bool))
			}
		})
	}
}

func TestHandler_AdminLogin(t *testing.T) {
	handler, _ := setupTestHandler(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful admin login",
			requestBody: AdminLoginRequest{
				Email:    "admin@ecommerce.com",
				Password: "admin123456",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid admin credentials",
			requestBody: AdminLoginRequest{
				Email:    "admin@ecommerce.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_CREDENTIALS",
		},
		{
			name: "wrong admin email",
			requestBody: AdminLoginRequest{
				Email:    "wrong@admin.com",
				Password: "admin123456",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_CREDENTIALS",
		},
		{
			name: "invalid email format",
			requestBody: AdminLoginRequest{
				Email:    "invalid-email",
				Password: "admin123456",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "missing password",
			requestBody: AdminLoginRequest{
				Email:    "admin@ecommerce.com",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/api/auth/admin/login", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.AdminLogin(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorDetail := response["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorDetail["code"])
			} else {
				assert.True(t, response["success"].(bool))

				// Verify response contains user and tokens
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "user")
				assert.Contains(t, data, "tokens")

				user := data["user"].(map[string]interface{})
				assert.Equal(t, "admin@ecommerce.com", user["email"])
				assert.Equal(t, "admin", user["role"])

				tokens := data["tokens"].(map[string]interface{})
				assert.Contains(t, tokens, "access_token")
				assert.Contains(t, tokens, "refresh_token")
			}
		})
	}
}
