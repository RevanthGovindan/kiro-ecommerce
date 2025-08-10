package payments

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-website/internal/auth"
	"ecommerce-website/internal/config"
	"ecommerce-website/internal/database"
	"ecommerce-website/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIntegrationTest(t *testing.T) (*gin.Engine, *auth.Service, string) {
	cfg := &config.Config{
		JWTSecret:      "test-secret",
		RazorpayKeyID:  "test_key_id",
		RazorpaySecret: "test_secret",
	}

	err := database.InitializeTest(cfg)
	require.NoError(t, err)

	// Create test user
	user := models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FirstName: "Test",
		LastName:  "User",
		Role:      "customer",
		IsActive:  true,
	}
	err = database.GetDB().Create(&user).Error
	require.NoError(t, err)

	// Create auth service and generate token
	authService := auth.NewService(database.GetDB(), cfg)
	tokens, err := authService.GenerateTokens(&user)
	require.NoError(t, err)

	// Create payment service and handler
	paymentService := NewService(database.GetDB(), cfg.RazorpayKeyID, cfg.RazorpaySecret)
	paymentHandler := NewHandler(paymentService)

	// Setup router
	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetupRoutes(r, paymentHandler, authService)

	return r, authService, tokens.AccessToken
}

func TestPaymentIntegration_CreateOrderFlow(t *testing.T) {
	r, _, token := setupIntegrationTest(t)

	// Create test order
	user := models.User{
		Email:     "test2@example.com",
		Password:  "hashedpassword",
		FirstName: "Test2",
		LastName:  "User2",
		Role:      "customer",
		IsActive:  true,
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

	// Test create payment order
	createReq := CreateOrderRequest{
		OrderID:     order.ID,
		Amount:      110.0,
		Currency:    "INR",
		Description: "Test payment integration",
	}

	jsonBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/payments/create-order", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail because we don't have real Razorpay credentials
	// But the request should be properly formatted and reach the service
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"].(map[string]interface{})["code"], "PAYMENT_ORDER_FAILED")
}

func TestPaymentIntegration_GetPaymentStatus(t *testing.T) {
	r, _, token := setupIntegrationTest(t)

	// Test get payment status for non-existent order
	req, _ := http.NewRequest("GET", "/api/payments/status/non-existent-order", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "PAYMENT_NOT_FOUND", response["error"].(map[string]interface{})["code"])
}

func TestPaymentIntegration_WebhookEndpoint(t *testing.T) {
	r, _, _ := setupIntegrationTest(t)

	// Test webhook with unknown event (should succeed)
	webhookPayload := map[string]interface{}{
		"event": "unknown.event",
		"payload": map[string]interface{}{
			"test": "data",
		},
	}

	jsonBody, _ := json.Marshal(webhookPayload)
	req, _ := http.NewRequest("POST", "/api/payments/webhook", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestPaymentIntegration_AuthenticationRequired(t *testing.T) {
	r, _, _ := setupIntegrationTest(t)

	// Test create order without authentication
	createReq := CreateOrderRequest{
		OrderID:     "test-order",
		Amount:      110.0,
		Currency:    "INR",
		Description: "Test payment",
	}

	jsonBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/payments/create-order", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "MISSING_TOKEN", response["error"].(map[string]interface{})["code"])
}
