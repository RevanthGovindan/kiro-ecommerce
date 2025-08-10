package orders

import (
	"context"
	"testing"
	"time"

	"ecommerce-website/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
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

	return db
}

func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FirstName: "Test",
		LastName:  "User",
		Role:      "customer",
		IsActive:  true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestCategory(t *testing.T, db *gorm.DB) *models.Category {
	category := &models.Category{
		Name:     "Test Category",
		Slug:     "test-category",
		IsActive: true,
	}
	err := db.Create(category).Error
	require.NoError(t, err)
	return category
}

func createTestProduct(t *testing.T, db *gorm.DB, categoryID string, inventory int) *models.Product {
	product := &models.Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		SKU:         "TEST-SKU-" + time.Now().Format("20060102150405"),
		Inventory:   inventory,
		IsActive:    true,
		CategoryID:  categoryID,
	}
	err := db.Create(product).Error
	require.NoError(t, err)
	return product
}

func TestService_CreateOrder(t *testing.T) {
	db := setupTestDB(t)
	mockCartService := new(MockCartService)
	service := NewServiceWithCartService(db, mockCartService)

	// Create test data
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	product := createTestProduct(t, db, category.ID, 10)

	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		request     *CreateOrderRequest
		setupMock   func()
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful order creation",
			userID: user.ID,
			request: &CreateOrderRequest{
				SessionID: "test-session-id",
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
			setupMock: func() {
				cart := &models.Cart{
					SessionID: "test-session-id",
					Items: []models.CartItem{
						{
							ProductID: product.ID,
							Quantity:  2,
							Product:   *product,
						},
					},
				}
				mockCartService.On("GetCartWithProducts", mock.Anything, "test-session-id").Return(cart, nil)
				mockCartService.On("ClearCart", mock.Anything, "test-session-id").Return(nil)
			},
			expectError: false,
		},
		{
			name:   "empty cart error",
			userID: user.ID,
			request: &CreateOrderRequest{
				SessionID: "empty-session-id",
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
			setupMock: func() {
				cart := &models.Cart{
					SessionID: "empty-session-id",
					Items:     []models.CartItem{},
				}
				mockCartService.On("GetCartWithProducts", mock.Anything, "empty-session-id").Return(cart, nil)
			},
			expectError: true,
			errorMsg:    "cart is empty",
		},
		{
			name:   "insufficient inventory error",
			userID: user.ID,
			request: &CreateOrderRequest{
				SessionID: "insufficient-inventory-session",
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
			setupMock: func() {
				cart := &models.Cart{
					SessionID: "insufficient-inventory-session",
					Items: []models.CartItem{
						{
							ProductID: product.ID,
							Quantity:  15, // More than available inventory (10)
							Product:   *product,
						},
					},
				}
				mockCartService.On("GetCartWithProducts", mock.Anything, "insufficient-inventory-session").Return(cart, nil)
			},
			expectError: true,
			errorMsg:    "insufficient inventory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockCartService.ExpectedCalls = nil
			tt.setupMock()

			order, err := service.CreateOrder(ctx, tt.userID, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				assert.Equal(t, tt.userID, order.UserID)
				assert.Equal(t, "pending", order.Status)
				assert.Greater(t, order.Total, 0.0)
			}

			// Verify mock expectations
			mockCartService.AssertExpectations(t)
		})
	}
}

func TestService_GetOrder(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create test data
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	product := createTestProduct(t, db, category.ID, 10)

	// Create test order
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

	// Create order item
	orderItem := &models.OrderItem{
		OrderID:   order.ID,
		ProductID: product.ID,
		Quantity:  1,
		Price:     99.99,
		Total:     99.99,
	}
	err = db.Create(orderItem).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		orderID     string
		userID      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "get order successfully",
			orderID:     order.ID,
			userID:      user.ID,
			expectError: false,
		},
		{
			name:        "order not found",
			orderID:     "non-existent-id",
			userID:      user.ID,
			expectError: true,
			errorMsg:    "order not found",
		},
		{
			name:        "get order as admin (no user filter)",
			orderID:     order.ID,
			userID:      "", // Empty userID means admin access
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetOrder(tt.orderID, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.orderID, result.ID)
			}
		})
	}
}

func TestService_GetUserOrders(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create test data
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	product := createTestProduct(t, db, category.ID, 10)

	// Create multiple test orders
	for i := 0; i < 5; i++ {
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

		// Create order item
		orderItem := &models.OrderItem{
			OrderID:   order.ID,
			ProductID: product.ID,
			Quantity:  1,
			Price:     99.99,
			Total:     99.99,
		}
		err = db.Create(orderItem).Error
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		userID        string
		page          int
		limit         int
		expectedCount int
		expectedTotal int64
	}{
		{
			name:          "get first page",
			userID:        user.ID,
			page:          1,
			limit:         3,
			expectedCount: 3,
			expectedTotal: 5,
		},
		{
			name:          "get second page",
			userID:        user.ID,
			page:          2,
			limit:         3,
			expectedCount: 2,
			expectedTotal: 5,
		},
		{
			name:          "get all orders",
			userID:        user.ID,
			page:          1,
			limit:         10,
			expectedCount: 5,
			expectedTotal: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders, total, err := service.GetUserOrders(tt.userID, tt.page, tt.limit)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(orders))
			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}

func TestService_UpdateOrderStatus(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create test data
	user := createTestUser(t, db)

	// Create test order
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
		name        string
		orderID     string
		status      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "update to processing",
			orderID:     order.ID,
			status:      "processing",
			expectError: false,
		},
		{
			name:        "update to shipped",
			orderID:     order.ID,
			status:      "shipped",
			expectError: false,
		},
		{
			name:        "invalid status",
			orderID:     order.ID,
			status:      "invalid",
			expectError: true,
			errorMsg:    "invalid order status",
		},
		{
			name:        "order not found",
			orderID:     "non-existent-id",
			status:      "processing",
			expectError: true,
			errorMsg:    "order not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.UpdateOrderStatus(tt.orderID, tt.status)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.status, result.Status)
			}
		})
	}
}

func TestService_GetAllOrders(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create test data
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	product := createTestProduct(t, db, category.ID, 10)

	// Create orders with different statuses
	statuses := []string{"pending", "processing", "shipped", "delivered"}
	for _, status := range statuses {
		order := &models.Order{
			UserID:   user.ID,
			Status:   status,
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

		// Create order item
		orderItem := &models.OrderItem{
			OrderID:   order.ID,
			ProductID: product.ID,
			Quantity:  1,
			Price:     99.99,
			Total:     99.99,
		}
		err = db.Create(orderItem).Error
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		page          int
		limit         int
		status        string
		expectedCount int
		expectedTotal int64
	}{
		{
			name:          "get all orders",
			page:          1,
			limit:         10,
			status:        "",
			expectedCount: 4,
			expectedTotal: 4,
		},
		{
			name:          "filter by pending status",
			page:          1,
			limit:         10,
			status:        "pending",
			expectedCount: 1,
			expectedTotal: 1,
		},
		{
			name:          "filter by processing status",
			page:          1,
			limit:         10,
			status:        "processing",
			expectedCount: 1,
			expectedTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders, total, err := service.GetAllOrders(tt.page, tt.limit, tt.status, "")

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(orders))
			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}

func TestService_GetAllCustomers(t *testing.T) {
	// Use a unique database for this test to avoid interference
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

	service := NewService(db)

	// Create test customers
	customer1 := &models.User{
		Email:     "john.doe@example.com",
		Password:  "hashedpassword",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		IsActive:  true,
	}
	customer2 := &models.User{
		Email:     "jane.smith@example.com",
		Password:  "hashedpassword",
		FirstName: "Jane",
		LastName:  "Smith",
		Role:      "customer",
		IsActive:  true,
	}
	admin := &models.User{
		Email:     "admin@example.com",
		Password:  "hashedpassword",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
		IsActive:  true,
	}
	inactiveCustomer := &models.User{
		Email:     "inactive@example.com",
		Password:  "hashedpassword",
		FirstName: "Inactive",
		LastName:  "Customer",
		Role:      "customer",
		IsActive:  false,
	}

	require.NoError(t, db.Create(customer1).Error)
	require.NoError(t, db.Create(customer2).Error)
	require.NoError(t, db.Create(admin).Error)
	require.NoError(t, db.Create(inactiveCustomer).Error)

	// Explicitly set inactive customer to false
	db.Model(inactiveCustomer).Update("is_active", false)

	tests := []struct {
		name          string
		page          int
		limit         int
		search        string
		expectedCount int
		expectedTotal int64
	}{
		{
			name:          "get all customers without search",
			page:          1,
			limit:         10,
			search:        "",
			expectedCount: 2, // Only active customers, not admin
			expectedTotal: 2,
		},
		{
			name:          "search by first name",
			page:          1,
			limit:         10,
			search:        "john",
			expectedCount: 1,
			expectedTotal: 1,
		},
		{
			name:          "search by last name",
			page:          1,
			limit:         10,
			search:        "smith",
			expectedCount: 1,
			expectedTotal: 1,
		},
		{
			name:          "search by email",
			page:          1,
			limit:         10,
			search:        "jane.smith",
			expectedCount: 1,
			expectedTotal: 1,
		},
		{
			name:          "no matching customers",
			page:          1,
			limit:         10,
			search:        "nonexistent",
			expectedCount: 0,
			expectedTotal: 0,
		},
		{
			name:          "pagination test",
			page:          1,
			limit:         1,
			search:        "",
			expectedCount: 1,
			expectedTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customers, total, err := service.GetAllCustomers(tt.page, tt.limit, tt.search)

			assert.NoError(t, err)

			assert.Equal(t, tt.expectedCount, len(customers))
			assert.Equal(t, tt.expectedTotal, total)

			// Verify all returned users are customers and active
			for _, customer := range customers {
				assert.Equal(t, "customer", customer.Role)
				assert.True(t, customer.IsActive)
				assert.Empty(t, customer.Password) // Password should be removed
			}
		})
	}
}

func TestService_UpdateOrderStatusWithEmailNotification(t *testing.T) {
	db := setupTestDB(t)

	// Create mock email service
	mockEmailService := &MockEmailService{}

	// Create service with mock dependencies
	service := NewServiceWithDependencies(db, &MockCartService{}, mockEmailService)

	// Create test data
	user := createTestUser(t, db)
	order := &models.Order{
		UserID:   user.ID,
		Status:   "pending",
		Subtotal: 100.0,
		Total:    100.0,
	}
	require.NoError(t, db.Create(order).Error)

	tests := []struct {
		name           string
		orderID        string
		newStatus      string
		mockSetup      func()
		expectedError  bool
		expectedStatus string
	}{
		{
			name:      "successful status update with email notification",
			orderID:   order.ID,
			newStatus: "processing",
			mockSetup: func() {
				mockEmailService.On("SendOrderStatusUpdate", mock.AnythingOfType("*models.Order"), "pending", "processing").Return(nil)
			},
			expectedError:  false,
			expectedStatus: "processing",
		},
		{
			name:      "status update with email failure (should not fail)",
			orderID:   order.ID,
			newStatus: "shipped",
			mockSetup: func() {
				mockEmailService.On("SendOrderStatusUpdate", mock.AnythingOfType("*models.Order"), "processing", "shipped").Return(assert.AnError)
			},
			expectedError:  false,
			expectedStatus: "shipped",
		},
		{
			name:          "invalid status",
			orderID:       order.ID,
			newStatus:     "invalid_status",
			mockSetup:     func() {},
			expectedError: true,
		},
		{
			name:          "non-existent order",
			orderID:       "non-existent-id",
			newStatus:     "processing",
			mockSetup:     func() {},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockEmailService.ExpectedCalls = nil
			tt.mockSetup()

			updatedOrder, err := service.UpdateOrderStatus(tt.orderID, tt.newStatus)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, updatedOrder)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updatedOrder)
				assert.Equal(t, tt.expectedStatus, updatedOrder.Status)
			}

			// Verify mock expectations
			mockEmailService.AssertExpectations(t)
		})
	}
}

// Mock email service for testing
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendOrderStatusUpdate(order *models.Order, oldStatus, newStatus string) error {
	args := m.Called(order, oldStatus, newStatus)
	return args.Error(0)
}
