package orders

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"ecommerce-website/internal/models"
)

// MockOrderRepository is a mock implementation of the order repository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(order *models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(id string) (*models.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByUserID(userID string, page, pageSize int) ([]*models.Order, int64, error) {
	args := m.Called(userID, page, pageSize)
	return args.Get(0).([]*models.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) Update(order *models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) List(filters OrderFilters) ([]*models.Order, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]*models.Order), args.Get(1).(int64), args.Error(2)
}

// MockProductRepository for order service tests
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetByID(id string) (*models.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) UpdateInventory(id string, quantity int) error {
	args := m.Called(id, quantity)
	return args.Error(0)
}

// MockPaymentService for order service tests
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) CreatePaymentIntent(amount float64, currency string) (string, error) {
	args := m.Called(amount, currency)
	return args.String(0), args.Error(1)
}

func (m *MockPaymentService) ConfirmPayment(paymentIntentID string) error {
	args := m.Called(paymentIntentID)
	return args.Error(0)
}

func TestOrderService_CreateOrder(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		orderData     *CreateOrderRequest
		setupMocks    func(*MockOrderRepository, *MockProductRepository, *MockPaymentService)
		expectedError string
	}{
		{
			name:   "successful order creation",
			userID: "user-123",
			orderData: &CreateOrderRequest{
				Items: []OrderItemRequest{
					{ProductID: "prod-1", Quantity: 2},
					{ProductID: "prod-2", Quantity: 1},
				},
				ShippingAddress: models.Address{
					FirstName: "John",
					LastName:  "Doe",
					Address:   "123 Main St",
					City:      "Anytown",
					State:     "CA",
					ZipCode:   "12345",
				},
				BillingAddress: models.Address{
					FirstName: "John",
					LastName:  "Doe",
					Address:   "123 Main St",
					City:      "Anytown",
					State:     "CA",
					ZipCode:   "12345",
				},
			},
			setupMocks: func(orderRepo *MockOrderRepository, prodRepo *MockProductRepository, paymentService *MockPaymentService) {
				// Products exist and have sufficient inventory
				prodRepo.On("GetByID", "prod-1").Return(&models.Product{
					ID: "prod-1", Name: "Product 1", Price: 99.99, Inventory: 10, IsActive: true,
				}, nil)
				prodRepo.On("GetByID", "prod-2").Return(&models.Product{
					ID: "prod-2", Name: "Product 2", Price: 49.99, Inventory: 5, IsActive: true,
				}, nil)

				// Update inventory
				prodRepo.On("UpdateInventory", "prod-1", -2).Return(nil)
				prodRepo.On("UpdateInventory", "prod-2", -1).Return(nil)

				// Create payment intent
				paymentService.On("CreatePaymentIntent", 249.97, "usd").Return("pi_test123", nil)

				// Create order
				orderRepo.On("Create", mock.AnythingOfType("*models.Order")).Return(nil)
			},
		},
		{
			name:   "insufficient inventory",
			userID: "user-123",
			orderData: &CreateOrderRequest{
				Items: []OrderItemRequest{
					{ProductID: "prod-1", Quantity: 15}, // More than available
				},
				ShippingAddress: models.Address{
					FirstName: "John",
					LastName:  "Doe",
					Address:   "123 Main St",
					City:      "Anytown",
					State:     "CA",
					ZipCode:   "12345",
				},
			},
			setupMocks: func(orderRepo *MockOrderRepository, prodRepo *MockProductRepository, paymentService *MockPaymentService) {
				prodRepo.On("GetByID", "prod-1").Return(&models.Product{
					ID: "prod-1", Name: "Product 1", Price: 99.99, Inventory: 10, IsActive: true,
				}, nil)
			},
			expectedError: "insufficient inventory for product",
		},
		{
			name:   "product not found",
			userID: "user-123",
			orderData: &CreateOrderRequest{
				Items: []OrderItemRequest{
					{ProductID: "nonexistent", Quantity: 1},
				},
				ShippingAddress: models.Address{
					FirstName: "John",
					LastName:  "Doe",
					Address:   "123 Main St",
					City:      "Anytown",
					State:     "CA",
					ZipCode:   "12345",
				},
			},
			setupMocks: func(orderRepo *MockOrderRepository, prodRepo *MockProductRepository, paymentService *MockPaymentService) {
				prodRepo.On("GetByID", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "product not found",
		},
		{
			name:   "inactive product",
			userID: "user-123",
			orderData: &CreateOrderRequest{
				Items: []OrderItemRequest{
					{ProductID: "prod-1", Quantity: 1},
				},
				ShippingAddress: models.Address{
					FirstName: "John",
					LastName:  "Doe",
					Address:   "123 Main St",
					City:      "Anytown",
					State:     "CA",
					ZipCode:   "12345",
				},
			},
			setupMocks: func(orderRepo *MockOrderRepository, prodRepo *MockProductRepository, paymentService *MockPaymentService) {
				prodRepo.On("GetByID", "prod-1").Return(&models.Product{
					ID: "prod-1", Name: "Product 1", Price: 99.99, Inventory: 10, IsActive: false,
				}, nil)
			},
			expectedError: "product is not available",
		},
		{
			name:      "empty order items",
			userID:    "user-123",
			orderData: &CreateOrderRequest{Items: []OrderItemRequest{}},
			setupMocks: func(orderRepo *MockOrderRepository, prodRepo *MockProductRepository, paymentService *MockPaymentService) {
			},
			expectedError: "order must contain at least one item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderRepo := new(MockOrderRepository)
			mockProdRepo := new(MockProductRepository)
			mockPaymentService := new(MockPaymentService)
			tt.setupMocks(mockOrderRepo, mockProdRepo, mockPaymentService)

			service := &Service{
				orderRepo:      mockOrderRepo,
				productRepo:    mockProdRepo,
				paymentService: mockPaymentService,
			}

			order, err := service.CreateOrder(context.Background(), tt.userID, tt.orderData)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				assert.Equal(t, tt.userID, order.UserID)
				assert.Equal(t, "pending", order.Status)
				assert.NotEmpty(t, order.ID)
			}

			mockOrderRepo.AssertExpectations(t)
			mockProdRepo.AssertExpectations(t)
			mockPaymentService.AssertExpectations(t)
		})
	}
}

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	tests := []struct {
		name          string
		orderID       string
		newStatus     string
		setupMock     func(*MockOrderRepository)
		expectedError string
	}{
		{
			name:      "successful status update",
			orderID:   "order-123",
			newStatus: "processing",
			setupMock: func(repo *MockOrderRepository) {
				order := &models.Order{
					ID:     "order-123",
					Status: "pending",
					UserID: "user-123",
				}
				repo.On("GetByID", "order-123").Return(order, nil)
				repo.On("Update", mock.AnythingOfType("*models.Order")).Return(nil)
			},
		},
		{
			name:      "order not found",
			orderID:   "nonexistent",
			newStatus: "processing",
			setupMock: func(repo *MockOrderRepository) {
				repo.On("GetByID", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "order not found",
		},
		{
			name:          "invalid status",
			orderID:       "order-123",
			newStatus:     "invalid-status",
			setupMock:     func(repo *MockOrderRepository) {},
			expectedError: "invalid order status",
		},
		{
			name:      "cannot update completed order",
			orderID:   "order-123",
			newStatus: "processing",
			setupMock: func(repo *MockOrderRepository) {
				order := &models.Order{
					ID:     "order-123",
					Status: "completed",
					UserID: "user-123",
				}
				repo.On("GetByID", "order-123").Return(order, nil)
			},
			expectedError: "cannot update completed order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockOrderRepository)
			tt.setupMock(mockRepo)

			service := &Service{
				orderRepo: mockRepo,
			}

			err := service.UpdateOrderStatus(context.Background(), tt.orderID, tt.newStatus)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestOrderService_CalculateOrderTotal(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name          string
		items         []models.OrderItem
		expectedTotal float64
	}{
		{
			name: "calculate total for multiple items",
			items: []models.OrderItem{
				{ProductID: "prod-1", Quantity: 2, Price: 99.99},
				{ProductID: "prod-2", Quantity: 1, Price: 49.99},
				{ProductID: "prod-3", Quantity: 3, Price: 19.99},
			},
			expectedTotal: 309.95, // (2 * 99.99) + (1 * 49.99) + (3 * 19.99)
		},
		{
			name:          "calculate total for empty order",
			items:         []models.OrderItem{},
			expectedTotal: 0.0,
		},
		{
			name: "calculate total for single item",
			items: []models.OrderItem{
				{ProductID: "prod-1", Quantity: 5, Price: 25.50},
			},
			expectedTotal: 127.50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := service.CalculateOrderTotal(tt.items)
			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}

func TestOrderService_ValidateOrderStatus(t *testing.T) {
	service := &Service{}

	validStatuses := []string{"pending", "processing", "shipped", "delivered", "cancelled", "refunded"}
	invalidStatuses := []string{"invalid", "unknown", "", "PENDING", "Processing"}

	for _, status := range validStatuses {
		t.Run("valid status: "+status, func(t *testing.T) {
			assert.True(t, service.ValidateOrderStatus(status))
		})
	}

	for _, status := range invalidStatuses {
		t.Run("invalid status: "+status, func(t *testing.T) {
			assert.False(t, service.ValidateOrderStatus(status))
		})
	}
}
