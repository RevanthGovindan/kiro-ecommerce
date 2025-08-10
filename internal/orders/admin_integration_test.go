package orders

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-website/internal/auth"
	"ecommerce-website/internal/config"
	"ecommerce-website/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIntegration_AdminOrderManagement(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
	)
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup auth service
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	authService := auth.NewService(db, cfg)

	// Create mock email service
	mockEmailService := &MockEmailService{}

	// Setup orders service with mocks
	mockCartService := &MockCartService{}
	ordersService := NewServiceWithDependencies(db, mockCartService, mockEmailService)
	ordersHandler := NewHandler(ordersService)

	// Setup routes
	SetupRoutes(router, ordersHandler, authService)

	// Create test admin user
	adminUser := &models.User{
		Email:     "admin@example.com",
		Password:  "hashedpassword",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
		IsActive:  true,
	}
	require.NoError(t, db.Create(adminUser).Error)

	// Create test customers
	customer1 := &models.User{
		Email:     "customer1@example.com",
		Password:  "hashedpassword",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		IsActive:  true,
	}
	customer2 := &models.User{
		Email:     "customer2@example.com",
		Password:  "hashedpassword",
		FirstName: "Jane",
		LastName:  "Smith",
		Role:      "customer",
		IsActive:  true,
	}
	require.NoError(t, db.Create(customer1).Error)
	require.NoError(t, db.Create(customer2).Error)

	// Create test orders
	order1 := &models.Order{
		UserID:   customer1.ID,
		Status:   "pending",
		Subtotal: 100.0,
		Total:    100.0,
	}
	order2 := &models.Order{
		UserID:   customer2.ID,
		Status:   "processing",
		Subtotal: 200.0,
		Total:    200.0,
	}
	require.NoError(t, db.Create(order1).Error)
	require.NoError(t, db.Create(order2).Error)

	// Generate admin JWT token
	adminTokens, err := authService.GenerateTokens(adminUser)
	require.NoError(t, err)
	adminToken := adminTokens.AccessToken

	t.Run("GET /api/admin/orders - success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/admin/orders?page=1&limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		orders := data["orders"].([]interface{})
		assert.Equal(t, 2, len(orders))
	})

	t.Run("GET /api/admin/orders with status filter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/admin/orders?status=pending", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		orders := data["orders"].([]interface{})
		assert.Equal(t, 1, len(orders))
	})

	t.Run("GET /api/admin/customers - success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/admin/customers?page=1&limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		customers := data["customers"].([]interface{})
		assert.Equal(t, 2, len(customers)) // Only customers, not admin
	})

	t.Run("GET /api/admin/customers with search", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/admin/customers?search=john", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		customers := data["customers"].([]interface{})
		assert.Equal(t, 1, len(customers))
	})

	t.Run("PUT /api/admin/orders/:id/status - success with email notification", func(t *testing.T) {
		// Setup mock for email notification
		mockEmailService.On("SendOrderStatusUpdate", mock.AnythingOfType("*models.Order"), "pending", "shipped").Return(nil)

		requestBody := map[string]string{
			"status": "shipped",
		}
		body, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("PUT", "/api/admin/orders/"+order1.ID+"/status", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "shipped", data["status"])

		// Verify email service was called
		mockEmailService.AssertExpectations(t)
	})

	t.Run("PUT /api/admin/orders/:id/status - invalid status", func(t *testing.T) {
		requestBody := map[string]string{
			"status": "invalid_status",
		}
		body, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("PUT", "/api/admin/orders/"+order2.ID+"/status", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
	})

	t.Run("Unauthorized access to admin endpoints", func(t *testing.T) {
		// Create customer token
		customerTokens, err := authService.GenerateTokens(customer1)
		require.NoError(t, err)
		customerToken := customerTokens.AccessToken

		// Test admin orders endpoint
		req, _ := http.NewRequest("GET", "/api/admin/orders", nil)
		req.Header.Set("Authorization", "Bearer "+customerToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		// Test admin customers endpoint
		req, _ = http.NewRequest("GET", "/api/admin/customers", nil)
		req.Header.Set("Authorization", "Bearer "+customerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
