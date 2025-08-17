package orders

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-website/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of the orders service
type MockService struct {
	mock.Mock
}

// Ensure MockService implements ServiceInterface
var _ ServiceInterface = (*MockService)(nil)

func (m *MockService) CreateOrder(ctx context.Context, userID string, req *CreateOrderRequest) (*models.Order, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockService) GetOrder(orderID string, userID string) (*models.Order, error) {
	args := m.Called(orderID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockService) GetUserOrders(userID string, page, limit int) ([]models.Order, int64, error) {
	args := m.Called(userID, page, limit)
	return args.Get(0).([]models.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) GetAllOrders(page, limit int, status, userID string) ([]models.Order, int64, error) {
	args := m.Called(page, limit, status, userID)
	return args.Get(0).([]models.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) UpdateOrderStatus(orderID string, status string) (*models.Order, error) {
	args := m.Called(orderID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockService) GetAllCustomers(page, limit int, search string) ([]models.User, int64, error) {
	args := m.Called(page, limit, search)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestHandler_CreateOrder(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter()

	// Setup route with middleware simulation
	router.POST("/api/orders/create", func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		handler.CreateOrder(c)
	})

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful order creation",
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
			mockSetup: func() {
				order := &models.Order{
					ID:       "test-order-id",
					UserID:   "test-user-id",
					Status:   "pending",
					Subtotal: 99.99,
					Total:    99.99,
				}
				mockService.On("CreateOrder", mock.Anything, "test-user-id", mock.AnythingOfType("*orders.CreateOrderRequest")).Return(order, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "empty cart error",
			requestBody: CreateOrderRequest{
				SessionID: "empty-session",
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
			mockSetup: func() {
				mockService.On("CreateOrder", mock.Anything, "test-user-id", mock.AnythingOfType("*orders.CreateOrderRequest")).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/orders/create", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_GetOrder(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter()

	// Setup route with middleware simulation
	router.GET("/api/orders/:id", func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Set("user_role", "customer")
		handler.GetOrder(c)
	})

	tests := []struct {
		name           string
		orderID        string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:    "successful order retrieval",
			orderID: "test-order-id",
			mockSetup: func() {
				order := &models.Order{
					ID:       "test-order-id",
					UserID:   "test-user-id",
					Status:   "pending",
					Subtotal: 99.99,
					Total:    99.99,
				}
				mockService.On("GetOrder", "test-order-id", "test-user-id").Return(order, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "order not found",
			orderID: "non-existent-id",
			mockSetup: func() {
				mockService.On("GetOrder", "non-existent-id", "test-user-id").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// Create request
			req, _ := http.NewRequest("GET", "/api/orders/"+tt.orderID, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_GetUserOrders(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter()

	// Setup route with middleware simulation
	router.GET("/api/orders", func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		handler.GetUserOrders(c)
	})

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:        "successful orders retrieval",
			queryParams: "?page=1&limit=10",
			mockSetup: func() {
				orders := []models.Order{
					{
						ID:       "order1",
						UserID:   "test-user-id",
						Status:   "pending",
						Subtotal: 99.99,
						Total:    99.99,
					},
				}
				mockService.On("GetUserOrders", "test-user-id", 1, 10).Return(orders, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "default pagination",
			queryParams: "",
			mockSetup: func() {
				orders := []models.Order{}
				mockService.On("GetUserOrders", "test-user-id", 1, 10).Return(orders, int64(0), nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// Create request
			req, _ := http.NewRequest("GET", "/api/orders"+tt.queryParams, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_UpdateOrderStatus(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter()

	// Setup route with middleware simulation (admin)
	router.PUT("/api/admin/orders/:id/status", func(c *gin.Context) {
		c.Set("user_id", "admin-user-id")
		c.Set("user_role", "admin")
		handler.UpdateOrderStatus(c)
	})

	tests := []struct {
		name           string
		orderID        string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:    "successful status update",
			orderID: "test-order-id",
			requestBody: map[string]string{
				"status": "processing",
			},
			mockSetup: func() {
				order := &models.Order{
					ID:       "test-order-id",
					UserID:   "test-user-id",
					Status:   "processing",
					Subtotal: 99.99,
					Total:    99.99,
				}
				mockService.On("UpdateOrderStatus", "test-order-id", "processing").Return(order, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "invalid request body",
			orderID: "test-order-id",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/api/admin/orders/"+tt.orderID+"/status", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestValidateOrderAddress(t *testing.T) {
	tests := []struct {
		name        string
		address     models.OrderAddress
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid address",
			address: models.OrderAddress{
				FirstName:  "John",
				LastName:   "Doe",
				Address1:   "123 Main St",
				City:       "Anytown",
				State:      "CA",
				PostalCode: "12345",
				Country:    "US",
			},
			expectError: false,
		},
		{
			name: "missing first name",
			address: models.OrderAddress{
				LastName:   "Doe",
				Address1:   "123 Main St",
				City:       "Anytown",
				State:      "CA",
				PostalCode: "12345",
				Country:    "US",
			},
			expectError: true,
			errorMsg:    "first name is required",
		},
		{
			name: "missing address",
			address: models.OrderAddress{
				FirstName:  "John",
				LastName:   "Doe",
				City:       "Anytown",
				State:      "CA",
				PostalCode: "12345",
				Country:    "US",
			},
			expectError: true,
			errorMsg:    "address line 1 is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOrderAddress(tt.address)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestHandler_GetAllOrders(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter()

	// Setup route with middleware simulation (admin)
	router.GET("/api/admin/orders", func(c *gin.Context) {
		c.Set("user_id", "admin-user-id")
		c.Set("user_role", "admin")
		handler.GetAllOrders(c)
	})

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:        "successful orders retrieval with filters",
			queryParams: "?page=1&limit=10&status=pending&userId=test-user-id",
			mockSetup: func() {
				orders := []models.Order{
					{
						ID:       "order1",
						UserID:   "test-user-id",
						Status:   "pending",
						Subtotal: 99.99,
						Total:    99.99,
					},
				}
				mockService.On("GetAllOrders", 1, 10, "pending", "test-user-id").Return(orders, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "default pagination without filters",
			queryParams: "",
			mockSetup: func() {
				orders := []models.Order{}
				mockService.On("GetAllOrders", 1, 10, "", "").Return(orders, int64(0), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "service error",
			queryParams: "?page=1&limit=10",
			mockSetup: func() {
				mockService.On("GetAllOrders", 1, 10, "", "").Return([]models.Order{}, int64(0), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// Create request
			req, _ := http.NewRequest("GET", "/api/admin/orders"+tt.queryParams, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_GetAllCustomers(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)
	router := setupTestRouter()

	// Setup route with middleware simulation (admin)
	router.GET("/api/admin/customers", func(c *gin.Context) {
		c.Set("user_id", "admin-user-id")
		c.Set("user_role", "admin")
		handler.GetAllCustomers(c)
	})

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:        "successful customers retrieval with search",
			queryParams: "?page=1&limit=10&search=john",
			mockSetup: func() {
				customers := []models.User{
					{
						ID:        "user1",
						Email:     "john@example.com",
						FirstName: "John",
						LastName:  "Doe",
						Role:      "customer",
						IsActive:  true,
					},
				}
				mockService.On("GetAllCustomers", 1, 10, "john").Return(customers, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "default pagination without search",
			queryParams: "",
			mockSetup: func() {
				customers := []models.User{}
				mockService.On("GetAllCustomers", 1, 10, "").Return(customers, int64(0), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "service error",
			queryParams: "?page=1&limit=10",
			mockSetup: func() {
				mockService.On("GetAllCustomers", 1, 10, "").Return([]models.User{}, int64(0), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "invalid pagination parameters",
			queryParams: "?page=0&limit=200",
			mockSetup: func() {
				customers := []models.User{}
				// Should use corrected values: page=1, limit=10
				mockService.On("GetAllCustomers", 1, 10, "").Return(customers, int64(0), nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// Create request
			req, _ := http.NewRequest("GET", "/api/admin/customers"+tt.queryParams, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}
