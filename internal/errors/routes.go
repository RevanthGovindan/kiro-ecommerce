package errors

import (
	"ecommerce-website/internal/auth"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up error handling routes
func SetupRoutes(r *gin.Engine, handler *Handler, authService *auth.Service) {
	// Public error logging endpoint (for client-side errors)
	r.POST("/api/errors/client", handler.LogClientError)

	// Public health check endpoint
	r.GET("/api/health", handler.HealthCheck)

	// Admin-only monitoring endpoints
	admin := r.Group("/api/admin/monitoring")
	admin.Use(authService.AuthMiddleware())
	admin.Use(authService.AdminMiddleware())
	{
		admin.GET("/metrics", handler.GetErrorMetrics)
		admin.GET("/alerts", handler.GetAlerts)
		admin.GET("/system", handler.GetSystemMetrics)
		admin.POST("/alerts/:id/resolve", handler.ResolveAlert)
	}
}
