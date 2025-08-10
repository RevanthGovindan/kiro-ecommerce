package payments

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

// CreateOrder creates a new Razorpay order for payment
func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", err.Error())
		return
	}

	// Set default currency if not provided
	if req.Currency == "" {
		req.Currency = "INR"
	}

	payment, err := h.service.CreateOrder(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "PAYMENT_ORDER_FAILED", "Failed to create payment order", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Payment order created successfully", payment)
}

// VerifyPayment verifies the payment signature and updates payment status
func (h *Handler) VerifyPayment(c *gin.Context) {
	var req VerifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", err.Error())
		return
	}

	if err := h.service.VerifyPayment(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "PAYMENT_VERIFICATION_FAILED", "Payment verification failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payment verified successfully", nil)
}

// GetPaymentStatus gets the payment status for an order
func (h *Handler) GetPaymentStatus(c *gin.Context) {
	orderID := c.Param("orderId")
	if orderID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_ORDER_ID", "Order ID is required", nil)
		return
	}

	payment, err := h.service.GetPaymentByOrderID(orderID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payment status retrieved successfully", payment)
}

// HandleWebhook handles Razorpay webhook events
func (h *Handler) HandleWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_WEBHOOK_PAYLOAD", "Invalid webhook payload", err.Error())
		return
	}

	if err := h.service.HandleWebhook(payload); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "WEBHOOK_PROCESSING_FAILED", "Failed to process webhook", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
