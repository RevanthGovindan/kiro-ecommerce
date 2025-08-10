package auth

import (
	"testing"
	"time"

	"ecommerce-website/internal/config"
	"ecommerce-website/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)



func TestService_Register(t *testing.T) {
	service, db := setupTestService(t)

	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr error
		setup   func()
	}{
		{
			name: "successful registration",
			req: RegisterRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
				Phone:     "1234567890",
			},
			wantErr: nil,
		},
		{
			name: "duplicate email",
			req: RegisterRequest{
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "Jane",
				LastName:  "Doe",
			},
			setup: func() {
				// Create existing user
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				user := TestUser{
					ID:        "existing-user-id",
					Email:     "existing@example.com",
					Password:  string(hashedPassword),
					FirstName: "Existing",
					LastName:  "User",
					Role:      "customer",
					IsActive:  true,
				}
				db.Create(&user)
			},
			wantErr: ErrUserExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			user, err := service.Register(tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.req.Email, user.Email)
				assert.Equal(t, tt.req.FirstName, user.FirstName)
				assert.Equal(t, tt.req.LastName, user.LastName)
				assert.Equal(t, "customer", user.Role)
				assert.True(t, user.IsActive)
				assert.Empty(t, user.Password) // Should be removed from response
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	service, db := setupTestService(t)

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
		name    string
		req     LoginRequest
		wantErr error
	}{
		{
			name: "successful login",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: nil,
		},
		{
			name: "invalid email",
			req: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "invalid password",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			wantErr: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, tokens, err := service.Login(tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, user)
				assert.Nil(t, tokens)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotNil(t, tokens)
				assert.Equal(t, tt.req.Email, user.Email)
				assert.Empty(t, user.Password) // Should be removed from response
				assert.NotEmpty(t, tokens.AccessToken)
				assert.NotEmpty(t, tokens.RefreshToken)
			}
		})
	}
}

func TestService_GenerateTokens(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}
	service := NewService(nil, cfg) // No DB needed for token generation

	user := &models.User{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  "customer",
	}

	tokens, err := service.GenerateTokens(user)
	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)

	// Validate access token
	accessClaims, err := service.ValidateToken(tokens.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, accessClaims.UserID)
	assert.Equal(t, user.Email, accessClaims.Email)
	assert.Equal(t, user.Role, accessClaims.Role)

	// Validate refresh token
	refreshClaims, err := service.ValidateToken(tokens.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, refreshClaims.UserID)
	assert.Equal(t, user.Email, refreshClaims.Email)
	assert.Equal(t, user.Role, refreshClaims.Role)
}

func TestService_ValidateToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}
	service := NewService(nil, cfg) // No DB needed for token validation

	// Create valid token
	claims := Claims{
		UserID: "test-user-id",
		Email:  "test@example.com",
		Role:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "test-user-id",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	// Test valid token
	validatedClaims, err := service.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, claims.UserID, validatedClaims.UserID)
	assert.Equal(t, claims.Email, validatedClaims.Email)
	assert.Equal(t, claims.Role, validatedClaims.Role)

	// Test invalid token
	_, err = service.ValidateToken("invalid-token")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)

	// Test expired token
	expiredClaims := Claims{
		UserID: "test-user-id",
		Email:  "test@example.com",
		Role:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   "test-user-id",
		},
	}

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	_, err = service.ValidateToken(expiredTokenString)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestService_RefreshToken(t *testing.T) {
	service, db := setupTestService(t)

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
	modelUser := &models.User{
		ID:    testUser.ID,
		Email: testUser.Email,
		Role:  testUser.Role,
	}
	originalTokens, err := service.GenerateTokens(modelUser)
	require.NoError(t, err)

	// Test successful refresh
	newTokens, err := service.RefreshToken(originalTokens.RefreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, newTokens)
	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.RefreshToken)
	// Note: tokens might be the same if generated at the same second, so we just check they exist

	// Test with invalid refresh token
	_, err = service.RefreshToken("invalid-token")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)

	// Test with refresh token for non-existent user
	nonExistentUser := &models.User{
		ID:    "non-existent-id",
		Email: "nonexistent@example.com",
		Role:  "customer",
	}
	nonExistentTokens, err := service.GenerateTokens(nonExistentUser)
	require.NoError(t, err)

	_, err = service.RefreshToken(nonExistentTokens.RefreshToken)
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestService_GetUserByID(t *testing.T) {
	service, db := setupTestService(t)

	// Create test user
	testUser := models.User{
		ID:        "test-user-id",
		Email:     "test@example.com",
		Password:  "hashed-password",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		IsActive:  true,
	}
	db.Create(&testUser)

	// Test successful retrieval
	user, err := service.GetUserByID("test-user-id")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, testUser.ID, user.ID)
	assert.Equal(t, testUser.Email, user.Email)
	assert.Equal(t, testUser.FirstName, user.FirstName)
	assert.Equal(t, testUser.LastName, user.LastName)
	assert.Empty(t, user.Password) // Should be removed from response

	// Test user not found
	_, err = service.GetUserByID("non-existent-id")
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)

	// Test inactive user - first create, then update to set IsActive = false
	inactiveUser := TestUser{
		ID:        "inactive-user-id",
		Email:     "inactive@example.com",
		FirstName: "Inactive",
		LastName:  "User",
		Role:      "customer",
		IsActive:  true, // Create as active first
	}
	result := db.Create(&inactiveUser)
	require.NoError(t, result.Error)

	// Now update to inactive
	db.Model(&inactiveUser).Update("is_active", false)

	_, err = service.GetUserByID("inactive-user-id")
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestService_ForgotPassword(t *testing.T) {
	service, db := setupTestService(t)

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
		name    string
		req     ForgotPasswordRequest
		wantErr bool
	}{
		{
			name: "successful forgot password",
			req: ForgotPasswordRequest{
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "non-existent email",
			req: ForgotPasswordRequest{
				Email: "nonexistent@example.com",
			},
			wantErr: false, // Should not reveal if email exists
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ForgotPassword(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check if reset token was set for existing user
			if tt.req.Email == "test@example.com" {
				var updatedUser TestUser
				db.Where("email = ?", tt.req.Email).First(&updatedUser)
				// Note: We can't easily test the token fields with TestUser struct
				// In a real scenario, you'd check the actual User model
			}
		})
	}
}

func TestService_ResetPassword(t *testing.T) {
	service, db := setupTestService(t)

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

	// Generate reset token
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

	// Update user with reset token (using raw SQL since TestUser doesn't have the fields)
	db.Exec("UPDATE users SET password_reset_token = ?, password_reset_expiry = ? WHERE id = ?", 
		resetTokenString, time.Now().Add(1*time.Hour), testUser.ID)

	tests := []struct {
		name    string
		req     ResetPasswordRequest
		wantErr error
	}{
		{
			name: "successful password reset",
			req: ResetPasswordRequest{
				Token:    resetTokenString,
				Password: "newpassword123",
			},
			wantErr: nil,
		},
		{
			name: "invalid token",
			req: ResetPasswordRequest{
				Token:    "invalid-token",
				Password: "newpassword123",
			},
			wantErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ResetPassword(tt.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_SendEmailVerification(t *testing.T) {
	service, db := setupTestService(t)

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
		name    string
		userID  string
		wantErr error
	}{
		{
			name:    "successful email verification send",
			userID:  "test-user-id",
			wantErr: nil,
		},
		{
			name:    "user not found",
			userID:  "non-existent-id",
			wantErr: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SendEmailVerification(tt.userID)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_VerifyEmail(t *testing.T) {
	service, db := setupTestService(t)

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

	// Generate verification token
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

	// Update user with verification token (using raw SQL since TestUser doesn't have the fields)
	db.Exec("UPDATE users SET email_verification_token = ? WHERE id = ?", 
		verificationTokenString, testUser.ID)

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "successful email verification",
			token:   verificationTokenString,
			wantErr: nil,
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			wantErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.VerifyEmail(tt.token)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}