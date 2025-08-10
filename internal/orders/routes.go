package orders

import (
	"ecommerce-website/internal/auth"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up the order routes
func SetupRoutes(r *gin.Engine, handler *Handler, authService *auth.Service) {
	api := r.Group("/api")

	// Public routes (none for orders)

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(authService.AuthMiddleware())
	{
		// User order routes
		protected.POST("/orders/create", handler.CreateOrder)
		protected.GET("/orders/:id", handler.GetOrder)
		protected.GET("/orders", handler.GetUserOrders)
	}

	// Admin routes (require admin role)
	admin := api.Group("/admin")
	admin.Use(authService.AuthMiddleware())
	admin.Use(authService.AdminMiddleware())
	{
		admin.GET("/orders", handler.GetAllOrders)
		admin.PUT("/orders/:id/status", handler.UpdateOrderStatus)
	}
}
