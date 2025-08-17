package users

import (
	"net/http"

	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetProfile handles GET /api/users/profile
func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated", nil)
		return
	}

	user, err := h.service.GetProfile(userID.(string))
	if err != nil {
		if err == ErrUserNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "user_not_found", "User not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve user profile", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", user)
}

// UpdateProfile handles PUT /api/users/profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated", nil)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}

	user, err := h.service.UpdateProfile(userID.(string), req)
	if err != nil {
		if err == ErrUserNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "user_not_found", "User not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update user profile", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile updated successfully", user)
}

// GetOrders handles GET /api/users/orders
func (h *Handler) GetOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated", nil)
		return
	}

	orders, err := h.service.GetUserOrders(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve user orders", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Orders retrieved successfully", orders)
}

// CreateAddress handles POST /api/users/addresses
func (h *Handler) CreateAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated", nil)
		return
	}

	var req CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}

	address, err := h.service.CreateAddress(userID.(string), req)
	if err != nil {
		if err == ErrUserNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "user_not_found", "User not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to create address", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Address created successfully", address)
}

// GetAddresses handles GET /api/users/addresses
func (h *Handler) GetAddresses(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated", nil)
		return
	}

	addresses, err := h.service.GetAddresses(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve addresses", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Addresses retrieved successfully", addresses)
}

// GetAddress handles GET /api/users/addresses/:id
func (h *Handler) GetAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated", nil)
		return
	}

	addressID := c.Param("id")
	if addressID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "validation_error", "Address ID is required", nil)
		return
	}

	address, err := h.service.GetAddress(userID.(string), addressID)
	if err != nil {
		if err == ErrAddressNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "address_not_found", "Address not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve address", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address retrieved successfully", address)
}

// UpdateAddress handles PUT /api/users/addresses/:id
func (h *Handler) UpdateAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated", nil)
		return
	}

	addressID := c.Param("id")
	if addressID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "validation_error", "Address ID is required", nil)
		return
	}

	var req UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}

	address, err := h.service.UpdateAddress(userID.(string), addressID, req)
	if err != nil {
		if err == ErrAddressNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "address_not_found", "Address not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update address", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address updated successfully", address)
}

// DeleteAddress handles DELETE /api/users/addresses/:id
func (h *Handler) DeleteAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated", nil)
		return
	}

	addressID := c.Param("id")
	if addressID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "validation_error", "Address ID is required", nil)
		return
	}

	err := h.service.DeleteAddress(userID.(string), addressID)
	if err != nil {
		if err == ErrAddressNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "address_not_found", "Address not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to delete address", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address deleted successfully", nil)
}
