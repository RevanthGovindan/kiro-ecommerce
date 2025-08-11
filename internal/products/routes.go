package products

import (
	"ecommerce-website/internal/auth"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures product and category routes
func SetupRoutes(router *gin.Engine, handler *Handler, authService *auth.Service) {
	// Product routes
	products := router.Group("/api/products")
	{
		products.GET("", handler.GetProducts)
		products.GET("/search", handler.SearchProducts)
		products.GET("/advanced-search", handler.AdvancedSearchProducts)
		products.GET("/suggestions", handler.GetSearchSuggestions)
		products.GET("/:id", handler.GetProductByID)
	}

	// Category routes
	categories := router.Group("/api/categories")
	{
		categories.GET("", handler.GetCategories)
		categories.GET("/:id", handler.GetCategoryByID)
	}

	// Admin product routes
	adminProducts := router.Group("/api/admin/products")
	adminProducts.Use(authService.AuthMiddleware())
	adminProducts.Use(authService.AdminMiddleware())
	{
		adminProducts.GET("", handler.GetAllProductsAdmin)
		adminProducts.POST("", handler.CreateProduct)
		adminProducts.PUT("/:id", handler.UpdateProduct)
		adminProducts.DELETE("/:id", handler.DeleteProduct)
		adminProducts.PUT("/:id/inventory", handler.UpdateInventory)
	}
}
