package cart

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers cart routes
func RegisterRoutes(router *gin.RouterGroup) {
	handler := NewHandler()
	
	cartGroup := router.Group("/cart")
	{
		cartGroup.GET("", handler.GetCart)
		cartGroup.POST("/add", handler.AddItem)
		cartGroup.PUT("/update", handler.UpdateItem)
		cartGroup.DELETE("/remove", handler.RemoveItem)
		cartGroup.DELETE("/clear", handler.ClearCart)
	}
}