package auth

import (
	"net/http"

	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", err.Error())
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		switch err {
		case ErrUserExists:
			utils.ErrorResponse(c, http.StatusConflict, "USER_EXISTS", "User with this email already exists", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "REGISTRATION_FAILED", "Failed to create user account", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", gin.H{
		"user": user,
	})
}

// Login handles user authentication
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", err.Error())
		return
	}

	user, tokens, err := h.service.Login(req)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "LOGIN_FAILED", "Login failed", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", gin.H{
		"user":   user,
		"tokens": tokens,
	})
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Refresh token is required", err.Error())
		return
	}

	tokens, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		switch err {
		case ErrInvalidToken:
			utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired refresh token", nil)
		case ErrUserNotFound:
			utils.ErrorResponse(c, http.StatusUnauthorized, "USER_NOT_FOUND", "User account not found or inactive", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "TOKEN_REFRESH_FAILED", "Failed to refresh token", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Token refreshed successfully", gin.H{
		"tokens": tokens,
	})
}

// Logout handles user logout (client-side token invalidation)
func (h *Handler) Logout(c *gin.Context) {
	// In a JWT-based system, logout is typically handled client-side by removing the token
	// For server-side logout, you would need to maintain a blacklist of tokens
	utils.SuccessResponse(c, http.StatusOK, "Logout successful", gin.H{
		"message": "Please remove the token from client storage",
	})
}

// Me returns the current user's information
func (h *Handler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "MISSING_USER_CONTEXT", "User context not found", nil)
		return
	}

	user, err := h.service.GetUserByID(userID.(string))
	if err != nil {
		switch err {
		case ErrUserNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "USER_FETCH_FAILED", "Failed to fetch user information", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User information retrieved successfully", gin.H{
		"user": user,
	})
}

// ForgotPassword handles password reset requests
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", err.Error())
		return
	}

	err := h.service.ForgotPassword(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FORGOT_PASSWORD_FAILED", "Failed to process password reset request", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password reset instructions sent to your email", gin.H{
		"message": "If an account with that email exists, you will receive password reset instructions",
	})
}

// ResetPassword handles password reset with token
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", err.Error())
		return
	}

	err := h.service.ResetPassword(req)
	if err != nil {
		switch err {
		case ErrInvalidToken:
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_TOKEN", "Invalid or expired reset token", nil)
		case ErrExpiredToken:
			utils.ErrorResponse(c, http.StatusBadRequest, "EXPIRED_TOKEN", "Reset token has expired", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "RESET_PASSWORD_FAILED", "Failed to reset password", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password reset successfully", gin.H{
		"message": "Your password has been reset successfully",
	})
}

// VerifyEmail handles email verification
func (h *Handler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_TOKEN", "Verification token is required", nil)
		return
	}

	err := h.service.VerifyEmail(token)
	if err != nil {
		switch err {
		case ErrInvalidToken:
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_TOKEN", "Invalid or expired verification token", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "EMAIL_VERIFICATION_FAILED", "Failed to verify email", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Email verified successfully", gin.H{
		"message": "Your email has been verified successfully",
	})
}

// ResendEmailVerification resends email verification
func (h *Handler) ResendEmailVerification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "MISSING_USER_CONTEXT", "User context not found", nil)
		return
	}

	err := h.service.SendEmailVerification(userID.(string))
	if err != nil {
		switch err {
		case ErrUserNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "EMAIL_VERIFICATION_FAILED", "Failed to send verification email", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Verification email sent", gin.H{
		"message": "Verification email has been sent to your email address",
	})
}

// AdminLogin handles admin authentication using environment credentials
func (h *Handler) AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", err.Error())
		return
	}

	user, tokens, err := h.service.AdminLogin(req)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid admin credentials", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "LOGIN_FAILED", "Admin login failed", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Admin login successful", gin.H{
		"user":   user,
		"tokens": tokens,
	})
}
