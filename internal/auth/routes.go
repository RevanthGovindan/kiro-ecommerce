package auth

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures authentication routes
func SetupRoutes(router *gin.Engine, handler *Handler, authService *Service) {
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/admin/login", handler.AdminLogin)
		auth.POST("/refresh", handler.RefreshToken)
		auth.POST("/logout", handler.Logout)
		auth.POST("/forgot-password", handler.ForgotPassword)
		auth.POST("/reset-password", handler.ResetPassword)
		auth.GET("/verify-email", handler.VerifyEmail)
		auth.GET("/me", authService.AuthMiddleware(), handler.Me)
		auth.POST("/resend-verification", authService.AuthMiddleware(), handler.ResendEmailVerification)
	}
}
