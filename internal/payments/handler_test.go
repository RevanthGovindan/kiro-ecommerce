package payments

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-website/internal/config"
	"ecommerce-website/internal/database"
	"ecommerce-website/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandler(t *testing.T) (*Handler, *gin.Engine) {
	cfg := &config.Config{}
	err := database.InitializeTest(cfg)
	require.NoError(t, err)

	service := NewService(database.GetDB(), "test_key_id", "test_secret")
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	return handler, r
}

func TestHandler_CreateOrder(t *testing.T) {
	handler, r := setupTestHandler(t)

	// Create test user and order
	user := models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FirstName: "Test",
		LastName:  "User",
	}
	err := database.GetDB().Create(&user).Error
	require.NoError(t, err)

	order := models.Order{
		UserID:   user.ID,
		Status:   "pending",
		Subtotal: 100.0,
		Total:    110.0,
	}
	err = database.GetDB().Create(&order).Error
	require.NoError(t, err)

	r.POST("/payments/create-order", handler.CreateOrder)

	tests := []struct {
		name           string
		requestBody    CreateOrderRequest
		expectedStatus int
	}{
		{
			name: "valid request",
			requestBody: CreateOrderRequest{
				OrderID:     order.ID,
				Amount:      110.0,
				Currency:    "INR",
				Description: "Test payment",
			},
			expectedStatus: http.StatusInternalServerError, // Will fail without real Razorpay credentials
		},
		{
			name: "invalid request - missing order ID",
			requestBody: CreateOrderRequest{
				Amount:      110.0,
				Currency:    "INR",
				Description: "Test payment",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid request - missing amount",
			requestBody: CreateOrderRequest{
				OrderID:     order.ID,
				Currency:    "INR",
				Description: "Test payment",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/payments/create-order", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_VerifyPayment(t *testing.T) {
	handler, r := setupTestHandler(t)
	r.POST("/payments/verify", handler.VerifyPayment)

	tests := []struct {
		name           string
		requestBody    VerifyPaymentRequest
		expectedStatus int
	}{
		{
			name: "valid request format",
			requestBody: VerifyPaymentRequest{
				RazorpayOrderID:   "order_test123",
				RazorpayPaymentID: "pay_test123",
				RazorpaySignature: "test_signature",
			},
			expectedStatus: http.StatusBadRequest, // Will fail verification
		},
		{
			name: "invalid request - missing order ID",
			requestBody: VerifyPaymentRequest{
				RazorpayPaymentID: "pay_test123",
				RazorpaySignature: "test_signature",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid request - missing payment ID",
			requestBody: VerifyPaymentRequest{
				RazorpayOrderID:   "order_test123",
				RazorpaySignature: "test_signature",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid request - missing signature",
			requestBody: VerifyPaymentRequest{
				RazorpayOrderID:   "order_test123",
				RazorpayPaymentID: "pay_test123",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/payments/verify", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_GetPaymentStatus(t *testing.T) {
	handler, r := setupTestHandler(t)
	r.GET("/payments/status/:orderId", handler.GetPaymentStatus)

	tests := []struct {
		name           string
		orderID        string
		expectedStatus int
	}{
		{
			name:           "valid order ID format",
			orderID:        "test-order-id",
			expectedStatus: http.StatusNotFound, // Payment won't exist
		},
		{
			name:           "empty order ID",
			orderID:        "",
			expectedStatus: http.StatusNotFound, // Route won't match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/payments/status/" + tt.orderID
			req, _ := http.NewRequest("GET", url, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_HandleWebhook(t *testing.T) {
	handler, r := setupTestHandler(t)
	r.POST("/payments/webhook", handler.HandleWebhook)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid webhook payload",
			requestBody: map[string]interface{}{
				"event": "payment.captured",
				"payload": map[string]interface{}{
					"payment": map[string]interface{}{
						"id":       "pay_test123",
						"order_id": "order_test123",
					},
				},
			},
			expectedStatus: http.StatusInternalServerError, // Will fail because payment doesn't exist
		},
		{
			name: "unknown event",
			requestBody: map[string]interface{}{
				"event": "unknown.event",
			},
			expectedStatus: http.StatusOK, // Should succeed for unknown events
		},
		{
			name:           "invalid JSON",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.requestBody != nil {
				jsonBody, _ := json.Marshal(tt.requestBody)
				req, _ = http.NewRequest("POST", "/payments/webhook", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest("POST", "/payments/webhook", bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
