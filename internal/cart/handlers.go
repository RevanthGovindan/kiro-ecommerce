package cart

import (
	"net/http"
	"strings"

	"ecommerce-website/internal/models"
	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

// NewHandler creates a new cart handler
func NewHandler() *Handler {
	return &Handler{
		service: NewService(),
	}
}

// GetCart retrieves the current cart
func (h *Handler) GetCart(c *gin.Context) {
	sessionID := h.getOrCreateSessionID(c)
	
	cart, err := h.service.GetCartWithProducts(c.Request.Context(), sessionID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "CART_RETRIEVE_ERROR", "Failed to retrieve cart", err.Error())
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Cart retrieved successfully", cart)
}

// AddItem adds an item to the cart
func (h *Handler) AddItem(c *gin.Context) {
	sessionID := h.getOrCreateSessionID(c)
	
	var req models.AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request data", err.Error())
		return
	}
	
	cart, err := h.service.AddItem(c.Request.Context(), sessionID, req.ProductID, req.Quantity)
	if err != nil {
		errMsg := err.Error()
		// Check if it's an inventory error
		if errMsg == "product is not available" {
			utils.ErrorResponse(c, http.StatusBadRequest, "PRODUCT_NOT_AVAILABLE", "Product is not available", errMsg)
			return
		}
		if strings.Contains(errMsg, "insufficient inventory") {
			utils.ErrorResponse(c, http.StatusBadRequest, "INSUFFICIENT_INVENTORY", errMsg, errMsg)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "CART_ADD_ERROR", "Failed to add item to cart", errMsg)
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Item added to cart successfully", cart)
}

// UpdateItem updates the quantity of an item in the cart
func (h *Handler) UpdateItem(c *gin.Context) {
	sessionID := h.getOrCreateSessionID(c)
	
	var req models.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request data", err.Error())
		return
	}
	
	cart, err := h.service.UpdateItem(c.Request.Context(), sessionID, req.ProductID, req.Quantity)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "item not found in cart" {
			utils.ErrorResponse(c, http.StatusNotFound, "ITEM_NOT_FOUND", "Item not found in cart", errMsg)
			return
		}
		if strings.Contains(errMsg, "insufficient inventory") {
			utils.ErrorResponse(c, http.StatusBadRequest, "INSUFFICIENT_INVENTORY", errMsg, errMsg)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "CART_UPDATE_ERROR", "Failed to update cart item", errMsg)
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Cart item updated successfully", cart)
}

// RemoveItem removes an item from the cart
func (h *Handler) RemoveItem(c *gin.Context) {
	sessionID := h.getOrCreateSessionID(c)
	
	var req models.RemoveItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request data", err.Error())
		return
	}
	
	cart, err := h.service.RemoveItem(c.Request.Context(), sessionID, req.ProductID)
	if err != nil {
		if err.Error() == "item not found in cart" {
			utils.ErrorResponse(c, http.StatusNotFound, "ITEM_NOT_FOUND", "Item not found in cart", err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "CART_REMOVE_ERROR", "Failed to remove item from cart", err.Error())
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Item removed from cart successfully", cart)
}

// ClearCart removes all items from the cart
func (h *Handler) ClearCart(c *gin.Context) {
	sessionID := h.getOrCreateSessionID(c)
	
	if err := h.service.ClearCart(c.Request.Context(), sessionID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "CART_CLEAR_ERROR", "Failed to clear cart", err.Error())
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Cart cleared successfully", nil)
}

// getOrCreateSessionID gets the session ID from cookie or creates a new one
func (h *Handler) getOrCreateSessionID(c *gin.Context) string {
	// Try to get session ID from cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		// Create new session ID
		sessionID = uuid.New().String()
		// Set cookie with 24 hour expiration
		c.SetCookie("session_id", sessionID, 86400, "/", "", false, true)
	}
	return sessionID
}