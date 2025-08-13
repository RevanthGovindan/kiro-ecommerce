package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"ecommerce-website/internal/models"
)

// MockUserRepository is a mock implementation of the user repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		firstName     string
		lastName      string
		setupMock     func(*MockUserRepository)
		expectedError string
	}{
		{
			name:      "successful registration",
			email:     "test@example.com",
			password:  "password123",
			firstName: "John",
			lastName:  "Doe",
			setupMock: func(repo *MockUserRepository) {
				repo.On("GetByEmail", "test@example.com").Return(nil, gorm.ErrRecordNotFound)
				repo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
			},
		},
		{
			name:      "user already exists",
			email:     "existing@example.com",
			password:  "password123",
			firstName: "John",
			lastName:  "Doe",
			setupMock: func(repo *MockUserRepository) {
				existingUser := &models.User{Email: "existing@example.com"}
				repo.On("GetByEmail", "existing@example.com").Return(existingUser, nil)
			},
			expectedError: "user already exists",
		},
		{
			name:          "invalid email",
			email:         "invalid-email",
			password:      "password123",
			firstName:     "John",
			lastName:      "Doe",
			setupMock:     func(repo *MockUserRepository) {},
			expectedError: "invalid email format",
		},
		{
			name:          "password too short",
			email:         "test@example.com",
			password:      "123",
			firstName:     "John",
			lastName:      "Doe",
			setupMock:     func(repo *MockUserRepository) {},
			expectedError: "password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			service := &Service{
				userRepo:  mockRepo,
				jwtSecret: "test-secret",
			}

			user, err := service.Register(context.Background(), tt.email, tt.password, tt.firstName, tt.lastName)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.firstName, user.FirstName)
				assert.Equal(t, tt.lastName, user.LastName)
				assert.NotEmpty(t, user.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		email         string
		password      string
		setupMock     func(*MockUserRepository)
		expectedError string
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			setupMock: func(repo *MockUserRepository) {
				user := &models.User{
					ID:        "user-123",
					Email:     "test@example.com",
					Password:  string(hashedPassword),
					FirstName: "John",
					LastName:  "Doe",
					IsActive:  true,
				}
				repo.On("GetByEmail", "test@example.com").Return(user, nil)
			},
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMock: func(repo *MockUserRepository) {
				repo.On("GetByEmail", "nonexistent@example.com").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "invalid credentials",
		},
		{
			name:     "incorrect password",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMock: func(repo *MockUserRepository) {
				user := &models.User{
					ID:        "user-123",
					Email:     "test@example.com",
					Password:  string(hashedPassword),
					FirstName: "John",
					LastName:  "Doe",
					IsActive:  true,
				}
				repo.On("GetByEmail", "test@example.com").Return(user, nil)
			},
			expectedError: "invalid credentials",
		},
		{
			name:     "inactive user",
			email:    "test@example.com",
			password: "password123",
			setupMock: func(repo *MockUserRepository) {
				user := &models.User{
					ID:        "user-123",
					Email:     "test@example.com",
					Password:  string(hashedPassword),
					FirstName: "John",
					LastName:  "Doe",
					IsActive:  false,
				}
				repo.On("GetByEmail", "test@example.com").Return(user, nil)
			},
			expectedError: "account is inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			service := &Service{
				userRepo:  mockRepo,
				jwtSecret: "test-secret",
			}

			token, user, err := service.Login(context.Background(), tt.email, tt.password)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, token)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	service := &Service{
		jwtSecret: "test-secret",
	}

	// Create a valid token
	claims := jwt.MapClaims{
		"user_id": "user-123",
		"email":   "test@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	validTokenString, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	// Create an expired token
	expiredClaims := jwt.MapClaims{
		"user_id": "user-123",
		"email":   "test@example.com",
		"exp":     time.Now().Add(-time.Hour).Unix(),
		"iat":     time.Now().Add(-2 * time.Hour).Unix(),
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
			expectedError: "token is expired",
		},
		{
			name:          "invalid token",
			tokenString:   "invalid.token.string",
			expectedError: "invalid token",
		},
		{
			name:          "empty token",
			tokenString:   "",
			expectedError: "token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := service.ValidateToken(tt.tokenString)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}
		})
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	service := &Service{
		jwtSecret: "test-secret",
	}

	user := &models.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	tokenString, err := service.GenerateToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Validate the generated token
	userID, err := service.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userID)
}

func TestAuthService_HashPassword(t *testing.T) {
	service := &Service{}

	password := "testpassword123"
	hashedPassword, err := service.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)

	// Verify the hash
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.NoError(t, err)
}

func TestAuthService_ComparePasswords(t *testing.T) {
	service := &Service{}

	password := "testpassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name           string
		hashedPassword string
		plainPassword  string
		shouldMatch    bool
	}{
		{
			name:           "matching passwords",
			hashedPassword: string(hashedPassword),
			plainPassword:  password,
			shouldMatch:    true,
		},
		{
			name:           "non-matching passwords",
			hashedPassword: string(hashedPassword),
			plainPassword:  "wrongpassword",
			shouldMatch:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := service.ComparePasswords(tt.hashedPassword, tt.plainPassword)
			assert.Equal(t, tt.shouldMatch, match)
		})
	}
}
