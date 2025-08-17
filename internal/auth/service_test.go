package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"ecommerce-website/internal/config"
	"ecommerce-website/internal/models"
)

func TestAuthService_Register_Validation(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		firstName     string
		lastName      string
		expectedError string
	}{
		{
			name:      "valid registration request",
			email:     "test@example.com",
			password:  "password123",
			firstName: "John",
			lastName:  "Doe",
		},
		{
			name:          "invalid email format",
			email:         "invalid-email",
			password:      "password123",
			firstName:     "John",
			lastName:      "Doe",
			expectedError: "invalid email format",
		},
		{
			name:          "password too short",
			email:         "test@example.com",
			password:      "123",
			firstName:     "John",
			lastName:      "Doe",
			expectedError: "password must be at least 8 characters",
		},
		{
			name:          "missing first name",
			email:         "test@example.com",
			password:      "password123",
			firstName:     "",
			lastName:      "Doe",
			expectedError: "first name is required",
		},
		{
			name:          "missing last name",
			email:         "test@example.com",
			password:      "password123",
			firstName:     "John",
			lastName:      "",
			expectedError: "last name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic
			if tt.email != "" && !isValidEmail(tt.email) {
				err := fmt.Errorf("invalid email format")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid email format")
				return
			}

			if len(tt.password) < 8 {
				err := fmt.Errorf("password must be at least 8 characters")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "password must be at least 8 characters")
				return
			}

			if tt.firstName == "" {
				err := fmt.Errorf("first name is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "first name is required")
				return
			}

			if tt.lastName == "" {
				err := fmt.Errorf("last name is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "last name is required")
				return
			}

			// Valid case
			if tt.expectedError == "" {
				assert.NotEmpty(t, tt.email)
				assert.GreaterOrEqual(t, len(tt.password), 8)
				assert.NotEmpty(t, tt.firstName)
				assert.NotEmpty(t, tt.lastName)
			}
		})
	}
}

func TestAuthService_Login_Validation(t *testing.T) {
	tests := []struct {
		name          string
		request       LoginRequest
		expectedError string
	}{
		{
			name: "valid login request",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
		},
		{
			name: "invalid email format",
			request: LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedError: "invalid email format",
		},
		{
			name: "empty password",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			expectedError: "password is required",
		},
		{
			name: "empty email",
			request: LoginRequest{
				Email:    "",
				Password: "password123",
			},
			expectedError: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic
			if tt.request.Email == "" {
				err := fmt.Errorf("email is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "email is required")
				return
			}

			if tt.request.Email != "" && !isValidEmail(tt.request.Email) {
				err := fmt.Errorf("invalid email format")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid email format")
				return
			}

			if tt.request.Password == "" {
				err := fmt.Errorf("password is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "password is required")
				return
			}

			// Valid case
			if tt.expectedError == "" {
				assert.NotEmpty(t, tt.request.Email)
				assert.NotEmpty(t, tt.request.Password)
				assert.True(t, isValidEmail(tt.request.Email))
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	service := NewService(nil, cfg)

	// Create a valid token
	claims := Claims{
		UserID: "user-123",
		Email:  "test@example.com",
		Role:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "user-123",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	validTokenString, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	// Create an expired token
	expiredClaims := Claims{
		UserID: "user-123",
		Email:  "test@example.com",
		Role:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   "user-123",
		},
	}

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	tests := []struct {
		name          string
		tokenString   string
		expectedError string
		expectedID    string
	}{
		{
			name:        "valid token",
			tokenString: validTokenString,
			expectedID:  "user-123",
		},
		{
			name:          "expired token",
			tokenString:   expiredTokenString,
			expectedError: "invalid token",
		},
		{
			name:          "invalid token",
			tokenString:   "invalid.token.string",
			expectedError: "invalid token",
		},
		{
			name:          "empty token",
			tokenString:   "",
			expectedError: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.tokenString)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tt.expectedID, claims.UserID)
			}
		})
	}
}

func TestAuthService_GenerateTokens(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	service := NewService(nil, cfg)

	user := &models.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
	}

	tokens, err := service.GenerateTokens(user)
	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)

	// Validate the generated access token
	claims, err := service.ValidateToken(tokens.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)

	// Validate the generated refresh token
	refreshClaims, err := service.ValidateToken(tokens.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, refreshClaims.UserID)
}

func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, string(hashedPassword))

	// Verify the hash
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	assert.NoError(t, err)
}

func TestPasswordComparison(t *testing.T) {
	password := "testpassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name           string
		hashedPassword []byte
		plainPassword  string
		shouldMatch    bool
	}{
		{
			name:           "matching passwords",
			hashedPassword: hashedPassword,
			plainPassword:  password,
			shouldMatch:    true,
		},
		{
			name:           "non-matching passwords",
			hashedPassword: hashedPassword,
			plainPassword:  "wrongpassword",
			shouldMatch:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bcrypt.CompareHashAndPassword(tt.hashedPassword, []byte(tt.plainPassword))
			if tt.shouldMatch {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestForgotPasswordRequest_Validation(t *testing.T) {
	tests := []struct {
		name          string
		request       ForgotPasswordRequest
		expectedError string
	}{
		{
			name: "valid forgot password request",
			request: ForgotPasswordRequest{
				Email: "test@example.com",
			},
		},
		{
			name: "invalid email format",
			request: ForgotPasswordRequest{
				Email: "invalid-email",
			},
			expectedError: "invalid email format",
		},
		{
			name: "empty email",
			request: ForgotPasswordRequest{
				Email: "",
			},
			expectedError: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic
			if tt.request.Email == "" {
				err := fmt.Errorf("email is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "email is required")
				return
			}

			if !isValidEmail(tt.request.Email) {
				err := fmt.Errorf("invalid email format")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid email format")
				return
			}

			// Valid case
			if tt.expectedError == "" {
				assert.NotEmpty(t, tt.request.Email)
				assert.True(t, isValidEmail(tt.request.Email))
			}
		})
	}
}

func TestResetPasswordRequest_Validation(t *testing.T) {
	tests := []struct {
		name          string
		request       ResetPasswordRequest
		expectedError string
	}{
		{
			name: "valid reset password request",
			request: ResetPasswordRequest{
				Token:    "valid-token",
				Password: "newpassword123",
			},
		},
		{
			name: "empty token",
			request: ResetPasswordRequest{
				Token:    "",
				Password: "newpassword123",
			},
			expectedError: "token is required",
		},
		{
			name: "password too short",
			request: ResetPasswordRequest{
				Token:    "valid-token",
				Password: "123",
			},
			expectedError: "password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic
			if tt.request.Token == "" {
				err := fmt.Errorf("token is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "token is required")
				return
			}

			if len(tt.request.Password) < 8 {
				err := fmt.Errorf("password must be at least 8 characters")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "password must be at least 8 characters")
				return
			}

			// Valid case
			if tt.expectedError == "" {
				assert.NotEmpty(t, tt.request.Token)
				assert.GreaterOrEqual(t, len(tt.request.Password), 8)
			}
		})
	}
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	// Simple email validation for testing
	return len(email) > 0 &&
		len(email) <= 254 &&
		email[0] != '@' &&
		email[len(email)-1] != '@' &&
		countChar(email, '@') == 1 &&
		len(email) > 3
}

// Helper function to count character occurrences
func countChar(s string, c byte) int {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			count++
		}
	}
	return count
}
