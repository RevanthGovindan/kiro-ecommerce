package orders

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func setupIntegrationTest(t *testing.T) (*gorm.DB, *gin.Engine, *auth.Service, *MockCartService) {
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

	// Setup mock cart service for integration tests
	mockCartService := &MockCartService{}

	// Setup orders service and handler
	ordersService := NewServiceWithCartService(db, mockCartService)
	ordersHandler := NewHandler(ordersService)

	// Setup routes
	SetupRoutes(router, ordersHandler, authService)

	return db, router, authService, mockCartService
}

func createIntegrationTestUser(t *testing.T, db *gorm.DB) (*models.User, string) {
	user := &models.User{
		Email:     "test@example.com",
		Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		FirstName: "Test",
		LastName:  "User",
		Role:      "customer",
		IsActive:  true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create auth service to generate token
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	authService := auth.NewService(db, cfg)

	token, err := authService.GenerateTokens(user)
	require.NoError(t, err)

	return user, token.AccessToken
}

func createIntegrationTestProduct(t *testing.T, db *gorm.DB) *models.Product {
	// Create category first
	category := &models.Category{
		Name:     "Test Category",
		Slug:     "test-category",
		IsActive: true,
	}
	err := db.Create(category).Error
	require.NoError(t, err)

	product := &models.Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		SKU:         "TEST-SKU-" + time.Now().Format("20060102150405"),
		Inventory:   10,
		IsActive:    true,
		CategoryID:  category.ID,
	}
	err = db.Create(product).Error
	require.NoError(t, err)

	return product
}

func TestIntegration_CreateOrder(t *testing.T) {
	db, router, _, mockCartService := setupIntegrationTest(t)
	user, token := createIntegrationTestUser(t, db)
	product := createIntegrationTestProduct(t, db)

	tests := []struct {
		name           string
		requestBody    CreateOrderRequest
		authToken      string
		setupMock      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "empty cart error (expected)",
			requestBody: CreateOrderRequest{
				SessionID: "test-session",
				ShippingAddress: models.OrderAddress{
					FirstName:  "John",
					LastName:   "Doe",
					Address1:   "123 Main St",
					City:       "Anytown",
					State:      "CA",
					PostalCode: "12345",
					Country:    "US",
				},
				BillingAddress: models.OrderAddress{
					FirstName:  "John",
					LastName:   "Doe",
					Address1:   "123 Main St",
					City:       "Anytown",
					State:      "CA",
					PostalCode: "12345",
					Country:    "US",
				},
				PaymentIntentID: "pi_test123",
			},
			authToken: token,
			setupMock: func() {
				cart := &models.Cart{
					SessionID: "test-session",
					Items:     []models.CartItem{},
				}
				mockCartService.On("GetCartWithProducts", mock.Anything, "test-session").Return(cart, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "EMPTY_CART",
		},
		{
			name: "missing auth token",
			requestBody: CreateOrderRequest{
				SessionID: "test-session",
				ShippingAddress: models.OrderAddress{
					FirstName:  "John",
					LastName:   "Doe",
					Address1:   "123 Main St",
					City:       "Anytown",
					State:      "CA",
					PostalCode: "12345",
					Country:    "US",
				},
				BillingAddress: models.OrderAddress{
					FirstName:  "John",
					LastName:   "Doe",
					Address1:   "123 Main St",
					City:       "Anytown",
					State:      "CA",
					PostalCode: "12345",
					Country:    "US",
				},
				PaymentIntentID: "pi_test123",
			},
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid shipping address",
			requestBody: CreateOrderRequest{
				SessionID: "test-session",
				ShippingAddress: models.OrderAddress{
					// Missing required fields
					FirstName: "John",
				},
				BillingAddress: models.OrderAddress{
					FirstName:  "John",
					LastName:   "Doe",
					Address1:   "123 Main St",
					City:       "Anytown",
					State:      "CA",
					PostalCode: "12345",
					Country:    "US",
				},
				PaymentIntentID: "pi_test123",
			},
			authToken:      token,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockCartService.ExpectedCalls = nil
			if tt.setupMock != nil {
				tt.setupMock()
			}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/orders/create", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				errorObj, exists := response["error"]
				require.True(t, exists)

				errorMap := errorObj.(map[string]interface{})
				code := errorMap["code"].(string)
				assert.Equal(t, tt.expectedError, code)
			}

			// Verify mock expectations if mock was set up
			if tt.setupMock != nil {
				mockCartService.AssertExpectations(t)
			}
		})
	}

	// Clean up
	_ = user
	_ = product
}

func TestIntegration_GetOrder(t *testing.T) {
	db, router, _, _ := setupIntegrationTest(t)
	user, token := createIntegrationTestUser(t, db)

	// Create a test order directly in the database
	order := &models.Order{
		UserID:   user.ID,
		Status:   "pending",
		Subtotal: 99.99,
		Tax:      0,
		Shipping: 0,
		Total:    99.99,
		ShippingAddress: models.OrderAddress{
			FirstName:  "John",
			LastName:   "Doe",
			Address1:   "123 Main St",
			City:       "Anytown",
			State:      "CA",
			PostalCode: "12345",
			Country:    "US",
		},
		BillingAddress: models.OrderAddress{
			FirstName:  "John",
			LastName:   "Doe",
			Address1:   "123 Main St",
			City:       "Anytown",
			State:      "CA",
			PostalCode: "12345",
			Country:    "US",
		},
		PaymentIntentID: "pi_test123",
	}
	err := db.Create(order).Error
	require.NoError(t, err)

	tests := []struct {
		name           string
		orderID        string
		authToken      string
		expectedStatus int
	}{
		{
			name:           "successful order retrieval",
			orderID:        order.ID,
			authToken:      token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "order not found",
			orderID:        "non-existent-id",
			authToken:      token,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing auth token",
			orderID:        order.ID,
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest("GET", "/api/orders/"+tt.orderID, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// If successful, check response structure
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				data, exists := response["data"]
				require.True(t, exists)

				orderData := data.(map[string]interface{})
				assert.Equal(t, order.ID, orderData["id"])
				assert.Equal(t, "pending", orderData["status"])
			}
		})
	}
}

func TestIntegration_GetUserOrders(t *testing.T) {
	db, router, _, _ := setupIntegrationTest(t)
	user, token := createIntegrationTestUser(t, db)

	// Create multiple test orders
	for i := 0; i < 3; i++ {
		order := &models.Order{
			UserID:   user.ID,
			Status:   "pending",
			Subtotal: 99.99,
			Tax:      0,
			Shipping: 0,
			Total:    99.99,
			ShippingAddress: models.OrderAddress{
				FirstName:  "John",
				LastName:   "Doe",
				Address1:   "123 Main St",
				City:       "Anytown",
				State:      "CA",
				PostalCode: "12345",
				Country:    "US",
			},
			BillingAddress: models.OrderAddress{
				FirstName:  "John",
				LastName:   "Doe",
				Address1:   "123 Main St",
				City:       "Anytown",
				State:      "CA",
				PostalCode: "12345",
				Country:    "US",
			},
			PaymentIntentID: "pi_test123",
		}
		err := db.Create(order).Error
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		queryParams    string
		authToken      string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "get all user orders",
			queryParams:    "",
			authToken:      token,
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "get orders with pagination",
			queryParams:    "?page=1&limit=2",
			authToken:      token,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "missing auth token",
			queryParams:    "",
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest("GET", "/api/orders"+tt.queryParams, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// If successful, check response structure
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				data, exists := response["data"]
				require.True(t, exists)

				dataMap := data.(map[string]interface{})
				orders := dataMap["orders"].([]interface{})
				assert.Equal(t, tt.expectedCount, len(orders))

				pagination := dataMap["pagination"].(map[string]interface{})
				assert.Equal(t, float64(3), pagination["total"])
			}
		})
	}
}
