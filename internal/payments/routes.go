package payments

import (
	"ecommerce-website/internal/auth"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, handler *Handler, authService *auth.Service) {
	api := r.Group("/api")

	// Payment routes
	payments := api.Group("/payments")
	{
		// Create payment order (requires authentication)
		payments.POST("/create-order", authService.AuthMiddleware(), handler.CreateOrder)

		// Verify payment (requires authentication)
		payments.POST("/verify", authService.AuthMiddleware(), handler.VerifyPayment)

		// Get payment status (requires authentication)
		payments.GET("/status/:orderId", authService.AuthMiddleware(), handler.GetPaymentStatus)

		// Webhook endpoint (no authentication required)
		payments.POST("/webhook", handler.HandleWebhook)
	}
}
