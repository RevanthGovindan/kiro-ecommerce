package products

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures product and category routes
func SetupRoutes(router *gin.Engine, handler *Handler) {
	// Product routes
	products := router.Group("/api/products")
	{
		products.GET("", handler.GetProducts)
		products.GET("/search", handler.SearchProducts)
		products.GET("/:id", handler.GetProductByID)
	}

	// Category routes
	categories := router.Group("/api/categories")
	{
		categories.GET("", handler.GetCategories)
		categories.GET("/:id", handler.GetCategoryByID)
	}
}