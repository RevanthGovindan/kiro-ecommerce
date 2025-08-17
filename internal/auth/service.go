package auth

import (
	"errors"
	"time"

	"ecommerce-website/internal/config"
	"ecommerce-website/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	config *config.Config
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Phone     string `json:"phone,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserExists         = errors.New("user with this email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token has expired")
)

func NewService(db *gorm.DB, config *config.Config) *Service {
	return &Service{
		db:     db,
		config: config,
	}
}

// Register creates a new user account
func (s *Service) Register(req RegisterRequest) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := models.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     &req.Phone,
		Role:      "customer",
		IsActive:  true,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// Send email verification
	if err := s.SendEmailVerification(user.ID); err != nil {
		// Log error but don't fail registration
		// In production, you might want to handle this differently
	}

	// Remove password from response
	user.Password = ""
	return &user, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(req LoginRequest) (*models.User, *TokenPair, error) {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokens, err := s.GenerateTokens(&user)
	if err != nil {
		return nil, nil, err
	}

	// Remove password from response
	user.Password = ""
	return &user, tokens, nil
}

// GenerateTokens creates access and refresh tokens for a user
func (s *Service) GenerateTokens(user *models.User) (*TokenPair, error) {
	// Access token (15 minutes)
	accessClaims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Refresh token (7 days)
	refreshClaims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken generates new tokens using a valid refresh token
func (s *Service) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// Get user from database to ensure they still exist and are active
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", claims.UserID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Generate new tokens
	return s.GenerateTokens(&user)
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Remove password from response
	user.Password = ""
	return &user, nil
}

// ForgotPassword generates a password reset token and sends reset email
func (s *Service) ForgotPassword(req ForgotPasswordRequest) error {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Don't reveal if email exists or not for security
			return nil
		}
		return err
	}

	// Generate reset token (using JWT for simplicity, but could use random token)
	resetClaims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // 1 hour expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	resetToken := jwt.NewWithClaims(jwt.SigningMethodHS256, resetClaims)
	resetTokenString, err := resetToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return err
	}

	// Store reset token and expiry in database
	expiryTime := time.Now().Add(1 * time.Hour)
	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"password_reset_token":  resetTokenString,
		"password_reset_expiry": expiryTime,
	}).Error; err != nil {
		return err
	}

	// TODO: Send email with reset link containing the token
	// For now, we'll just store the token in the database
	// In a real implementation, you would send an email here

	return nil
}

// ResetPassword resets user password using a valid reset token
func (s *Service) ResetPassword(req ResetPasswordRequest) error {
	// Validate reset token
	claims, err := s.ValidateToken(req.Token)
	if err != nil {
		return ErrInvalidToken
	}

	// Check if token is for password reset
	if claims.Role != "password_reset" {
		return ErrInvalidToken
	}

	// Find user and verify reset token
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ? AND password_reset_token = ?",
		claims.UserID, true, req.Token).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidToken
		}
		return err
	}

	// Check if token has expired
	if user.PasswordResetExpiry != nil && time.Now().After(*user.PasswordResetExpiry) {
		return ErrExpiredToken
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password and clear reset token
	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"password":              string(hashedPassword),
		"password_reset_token":  nil,
		"password_reset_expiry": nil,
	}).Error; err != nil {
		return err
	}

	return nil
}

// SendEmailVerification generates and sends email verification token
func (s *Service) SendEmailVerification(userID string) error {
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Don't send if already verified
	if user.EmailVerified {
		return nil
	}

	// Generate verification token
	verificationClaims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   "email_verification",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hour expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	verificationToken := jwt.NewWithClaims(jwt.SigningMethodHS256, verificationClaims)
	verificationTokenString, err := verificationToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return err
	}

	// Store verification token in database
	if err := s.db.Model(&user).Update("email_verification_token", verificationTokenString).Error; err != nil {
		return err
	}

	// TODO: Send verification email
	// For now, we'll just store the token in the database
	// In a real implementation, you would send an email here

	return nil
}

// VerifyEmail verifies user email using verification token
func (s *Service) VerifyEmail(token string) error {
	// Validate verification token
	claims, err := s.ValidateToken(token)
	if err != nil {
		return ErrInvalidToken
	}

	// Check if token is for email verification
	if claims.Role != "email_verification" {
		return ErrInvalidToken
	}

	// Find user and verify token
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ? AND email_verification_token = ?",
		claims.UserID, true, token).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidToken
		}
		return err
	}

	// Mark email as verified and clear verification token
	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"email_verified":           true,
		"email_verification_token": nil,
	}).Error; err != nil {
		return err
	}

	return nil
}

// AdminLogin authenticates an admin using environment credentials
func (s *Service) AdminLogin(req AdminLoginRequest) (*models.User, *TokenPair, error) {
	// Check against environment credentials
	if req.Email != s.config.AdminEmail || req.Password != s.config.AdminPassword {
		return nil, nil, ErrInvalidCredentials
	}

	// Create a virtual admin user for token generation
	adminUser := &models.User{
		ID:        "admin-user",
		Email:     s.config.AdminEmail,
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
		IsActive:  true,
	}

	// Generate tokens
	tokens, err := s.GenerateTokens(adminUser)
	if err != nil {
		return nil, nil, err
	}

	return adminUser, tokens, nil
}
