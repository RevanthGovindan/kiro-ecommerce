package users

import (
	"ecommerce-website/internal/auth"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up the user profile management routes
func SetupRoutes(r *gin.Engine, handler *Handler, authService *auth.Service) {
	api := r.Group("/api")

	// User profile routes (require authentication)
	userRoutes := api.Group("/users")
	userRoutes.Use(authService.AuthMiddleware())
	{
		// Profile management
		userRoutes.GET("/profile", handler.GetProfile)
		userRoutes.PUT("/profile", handler.UpdateProfile)

		// Order history
		userRoutes.GET("/orders", handler.GetOrders)

		// Address management
		userRoutes.POST("/addresses", handler.CreateAddress)
		userRoutes.GET("/addresses", handler.GetAddresses)
		userRoutes.GET("/addresses/:id", handler.GetAddress)
		userRoutes.PUT("/addresses/:id", handler.UpdateAddress)
		userRoutes.DELETE("/addresses/:id", handler.DeleteAddress)
	}
}
