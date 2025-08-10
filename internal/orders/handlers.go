package orders

import (
	"fmt"
	"net/http"
	"strconv"

	"ecommerce-website/internal/models"
	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServiceInterface
}

// NewHandler creates a new orders handler
func NewHandler(service ServiceInterface) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateOrder handles POST /api/orders/create
func (h *Handler) CreateOrder(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	// Validate required address fields
	if err := validateOrderAddress(req.ShippingAddress); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_SHIPPING_ADDRESS", "Invalid shipping address", err.Error())
		return
	}

	if err := validateOrderAddress(req.BillingAddress); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BILLING_ADDRESS", "Invalid billing address", err.Error())
		return
	}

	// Create order
	order, err := h.service.CreateOrder(c.Request.Context(), userID.(string), &req)
	if err != nil {
		// Check for specific error types
		switch err.Error() {
		case "cart is empty":
			utils.ErrorResponse(c, http.StatusBadRequest, "EMPTY_CART", "Cart is empty", nil)
		default:
			if contains(err.Error(), "insufficient inventory") {
				utils.ErrorResponse(c, http.StatusConflict, "INSUFFICIENT_INVENTORY", err.Error(), nil)
			} else if contains(err.Error(), "no longer available") {
				utils.ErrorResponse(c, http.StatusConflict, "PRODUCT_UNAVAILABLE", err.Error(), nil)
			} else {
				utils.ErrorResponse(c, http.StatusInternalServerError, "ORDER_CREATION_FAILED", "Failed to create order", err.Error())
			}
		}
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Order created successfully", order)
}

// GetOrder handles GET /api/orders/:id
func (h *Handler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ORDER_ID", "Order ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	var userIDStr string
	if exists {
		userIDStr = userID.(string)
	}

	// Check if user is admin
	userRole, _ := c.Get("userRole")
	isAdmin := userRole == "admin"

	// If not admin, must provide user ID for filtering
	if !isAdmin && !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Admin can view any order, regular users can only view their own
	var filterUserID string
	if !isAdmin {
		filterUserID = userIDStr
	}

	order, err := h.service.GetOrder(orderID, filterUserID)
	if err != nil {
		if err.Error() == "order not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "GET_ORDER_FAILED", "Failed to get order", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Order retrieved successfully", order)
}

// GetUserOrders handles GET /api/orders (for authenticated users)
func (h *Handler) GetUserOrders(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	orders, total, err := h.service.GetUserOrders(userID.(string), page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "GET_ORDERS_FAILED", "Failed to get orders", err.Error())
		return
	}

	// Calculate pagination info
	totalPages := (int(total) + limit - 1) / limit

	response := gin.H{
		"orders": orders,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Orders retrieved successfully", response)
}

// GetAllOrders handles GET /api/admin/orders (admin only)
func (h *Handler) GetAllOrders(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	userID := c.Query("userId")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	orders, total, err := h.service.GetAllOrders(page, limit, status, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "GET_ORDERS_FAILED", "Failed to get orders", err.Error())
		return
	}

	// Calculate pagination info
	totalPages := (int(total) + limit - 1) / limit

	response := gin.H{
		"orders": orders,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Orders retrieved successfully", response)
}

// GetAllCustomers handles GET /api/admin/customers (admin only)
func (h *Handler) GetAllCustomers(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	customers, total, err := h.service.GetAllCustomers(page, limit, search)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "GET_CUSTOMERS_FAILED", "Failed to get customers", err.Error())
		return
	}

	// Calculate pagination info
	totalPages := (int(total) + limit - 1) / limit

	response := gin.H{
		"customers": customers,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Customers retrieved successfully", response)
}

// UpdateOrderStatus handles PUT /api/admin/orders/:id/status (admin only)
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ORDER_ID", "Order ID is required", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	order, err := h.service.UpdateOrderStatus(orderID, req.Status)
	if err != nil {
		if err.Error() == "order not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found", nil)
		} else if contains(err.Error(), "invalid order status") {
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_STATUS", err.Error(), nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_STATUS_FAILED", "Failed to update order status", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Order status updated successfully", order)
}

// validateOrderAddress validates required fields in an order address
func validateOrderAddress(addr models.OrderAddress) error {
	if addr.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if addr.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	if addr.Address1 == "" {
		return fmt.Errorf("address line 1 is required")
	}
	if addr.City == "" {
		return fmt.Errorf("city is required")
	}
	if addr.State == "" {
		return fmt.Errorf("state is required")
	}
	if addr.PostalCode == "" {
		return fmt.Errorf("postal code is required")
	}
	if addr.Country == "" {
		return fmt.Errorf("country is required")
	}
	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0)))
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
