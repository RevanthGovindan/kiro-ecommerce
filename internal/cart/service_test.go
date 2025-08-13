package cart

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"ecommerce-website/internal/models"
)

// MockCartRepository is a mock implementation of the cart repository
type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) GetCart(sessionID string) (*models.Cart, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (m *MockCartRepository) SaveCart(cart *models.Cart) error {
	args := m.Called(cart)
	return args.Error(0)
}

func (m *MockCartRepository) DeleteCart(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockCartRepository) SetExpiration(sessionID string, expiration time.Duration) error {
	args := m.Called(sessionID, expiration)
	return args.Error(0)
}

// MockProductRepository for cart service tests
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

func TestCartService_AddToCart(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		productID     string
		quantity      int
		setupMocks    func(*MockCartRepository, *MockProductRepository)
		expectedError string
	}{
		{
			name:      "add new item to empty cart",
			sessionID: "session-123",
			productID: "prod-123",
			quantity:  2,
			setupMocks: func(cartRepo *MockCartRepository, prodRepo *MockProductRepository) {
				// Cart doesn't exist yet
				cartRepo.On("GetCart", "session-123").Return(nil, gorm.ErrRecordNotFound)

				// Product exists and has inventory
				product := &models.Product{
					ID:        "prod-123",
					Name:      "Test Product",
					Price:     99.99,
					Inventory: 10,
					IsActive:  true,
				}
				prodRepo.On("GetByID", "prod-123").Return(product, nil)

				// Save cart
				cartRepo.On("SaveCart", mock.AnythingOfType("*models.Cart")).Return(nil)
				cartRepo.On("SetExpiration", "session-123", mock.AnythingOfType("time.Duration")).Return(nil)
			},
		},
		{
			name:      "add item to existing cart",
			sessionID: "session-123",
			productID: "prod-123",
			quantity:  1,
			setupMocks: func(cartRepo *MockCartRepository, prodRepo *MockProductRepository) {
				// Existing cart with different item
				existingCart := &models.Cart{
					SessionID: "session-123",
					Items: []models.CartItem{
						{
							ProductID: "prod-456",
							Quantity:  1,
							Price:     49.99,
						},
					},
					Total: 49.99,
				}
				cartRepo.On("GetCart", "session-123").Return(existingCart, nil)

				// Product exists and has inventory
				product := &models.Product{
					ID:        "prod-123",
					Name:      "Test Product",
					Price:     99.99,
					Inventory: 10,
					IsActive:  true,
				}
				prodRepo.On("GetByID", "prod-123").Return(product, nil)

				// Save updated cart
				cartRepo.On("SaveCart", mock.AnythingOfType("*models.Cart")).Return(nil)
				cartRepo.On("SetExpiration", "session-123", mock.AnythingOfType("time.Duration")).Return(nil)
			},
		},
		{
			name:      "update quantity of existing item",
			sessionID: "session-123",
			productID: "prod-123",
			quantity:  3,
			setupMocks: func(cartRepo *MockCartRepository, prodRepo *MockProductRepository) {
				// Existing cart with the same item
				existingCart := &models.Cart{
					SessionID: "session-123",
					Items: []models.CartItem{
						{
							ProductID: "prod-123",
							Quantity:  2,
							Price:     99.99,
						},
					},
					Total: 199.98,
				}
				cartRepo.On("GetCart", "session-123").Return(existingCart, nil)

				// Product exists and has inventory
				product := &models.Product{
					ID:        "prod-123",
					Name:      "Test Product",
					Price:     99.99,
					Inventory: 10,
					IsActive:  true,
				}
				prodRepo.On("GetByID", "prod-123").Return(product, nil)

				// Save updated cart
				cartRepo.On("SaveCart", mock.AnythingOfType("*models.Cart")).Return(nil)
				cartRepo.On("SetExpiration", "session-123", mock.AnythingOfType("time.Duration")).Return(nil)
			},
		},
		{
			name:      "product not found",
			sessionID: "session-123",
			productID: "nonexistent",
			quantity:  1,
			setupMocks: func(cartRepo *MockCartRepository, prodRepo *MockProductRepository) {
				prodRepo.On("GetByID", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "product not found",
		},
		{
			name:      "insufficient inventory",
			sessionID: "session-123",
			productID: "prod-123",
			quantity:  15,
			setupMocks: func(cartRepo *MockCartRepository, prodRepo *MockProductRepository) {
				cartRepo.On("GetCart", "session-123").Return(nil, gorm.ErrRecordNotFound)

				product := &models.Product{
					ID:        "prod-123",
					Name:      "Test Product",
					Price:     99.99,
					Inventory: 10,
					IsActive:  true,
				}
				prodRepo.On("GetByID", "prod-123").Return(product, nil)
			},
			expectedError: "insufficient inventory",
		},
		{
			name:      "inactive product",
			sessionID: "session-123",
			productID: "prod-123",
			quantity:  1,
			setupMocks: func(cartRepo *MockCartRepository, prodRepo *MockProductRepository) {
				product := &models.Product{
					ID:        "prod-123",
					Name:      "Test Product",
					Price:     99.99,
					Inventory: 10,
					IsActive:  false,
				}
				prodRepo.On("GetByID", "prod-123").Return(product, nil)
			},
			expectedError: "product is not available",
		},
		{
			name:          "invalid quantity",
			sessionID:     "session-123",
			productID:     "prod-123",
			quantity:      0,
			setupMocks:    func(cartRepo *MockCartRepository, prodRepo *MockProductRepository) {},
			expectedError: "quantity must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCartRepo := new(MockCartRepository)
			mockProdRepo := new(MockProductRepository)
			tt.setupMocks(mockCartRepo, mockProdRepo)

			service := &Service{
				cartRepo:    mockCartRepo,
				productRepo: mockProdRepo,
			}

			err := service.AddToCart(context.Background(), tt.sessionID, tt.productID, tt.quantity)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockCartRepo.AssertExpectations(t)
			mockProdRepo.AssertExpectations(t)
		})
	}
}

func TestCartService_GetCart(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		setupMock     func(*MockCartRepository)
		expectedError string
		expectedItems int
	}{
		{
			name:      "get existing cart",
			sessionID: "session-123",
			setupMock: func(repo *MockCartRepository) {
				cart := &models.Cart{
					SessionID: "session-123",
					Items: []models.CartItem{
						{
							ProductID: "prod-123",
							Quantity:  2,
							Price:     99.99,
						},
						{
							ProductID: "prod-456",
							Quantity:  1,
							Price:     49.99,
						},
					},
					Total: 249.97,
				}
				repo.On("GetCart", "session-123").Return(cart, nil)
			},
			expectedItems: 2,
		},
		{
			name:      "get empty cart",
			sessionID: "session-456",
			setupMock: func(repo *MockCartRepository) {
				repo.On("GetCart", "session-456").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedItems: 0,
		},
		{
			name:          "invalid session ID",
			sessionID:     "",
			setupMock:     func(repo *MockCartRepository) {},
			expectedError: "session ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCartRepository)
			tt.setupMock(mockRepo)

			service := &Service{
				cartRepo: mockRepo,
			}

			cart, err := service.GetCart(context.Background(), tt.sessionID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, cart)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cart)
				assert.Len(t, cart.Items, tt.expectedItems)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCartService_UpdateCartItem(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		productID     string
		quantity      int
		setupMocks    func(*MockCartRepository, *MockProductRepository)
		expectedError string
	}{
		{
			name:      "update existing item quantity",
			sessionID: "session-123",
			productID: "prod-123",
			quantity:  5,
			setupMocks: func(cartRepo *MockCartRepository, prodRepo *MockProductRepository) {
				existingCart := &models.Cart{
					SessionID: "session-123",
					Items: []models.CartItem{
						{
							ProductID: "prod-123",
							Quantity:  2,
							Price:     99.99,
						},
					},
					Total: 199.98,
				}
				cartRepo.On("GetCart", "session-123").Return(existingCart, nil)

				product := &models.Product{
					ID:        "prod-123",
					Inventory: 10,
					IsActive:  true,
				}
				prodRepo.On("GetByID", "prod-123").Return(product, nil)

				cartRepo.On("SaveCart", mock.AnythingOfType("*models.Cart")).Return(nil)
				cartRepo.On("SetExpiration", "session-123", mock.AnythingOfType("time.Duration")).Return(nil)
			},
		},
		{
			name:      "remove item when quantity is 0",
			sessionID: "session-123",
			productID: "prod-123",
			quantity:  0,
			setupMocks: func(cartRepo *MockCartRepository, prodRepo *MockProductRepository) {
				existingCart := &models.Cart{
					SessionID: "session-123",
					Items: []models.CartItem{
						{
							ProductID: "prod-123",
							Quantity:  2,
							Price:     99.99,
						},
						{
							ProductID: "prod-456",
							Quantity:  1,
							Price:     49.99,
						},
					},
					Total: 249.97,
				}
				cartRepo.On("GetCart", "session-123").Return(existingCart, nil)
				cartRepo.On("SaveCart", mock.AnythingOfType("*models.Cart")).Return(nil)
				cartRepo.On("SetExpiration", "session-123", mock.AnythingOfType("time.Duration")).Return(nil)
			},
		},
		{
			name:      "item not in cart",
			sessionID: "session-123",
			productID: "prod-999",
			quantity:  1,
			setupMocks: func(cartRepo *MockCartRepository, prodRepo *MockProductRepository) {
				existingCart := &models.Cart{
					SessionID: "session-123",
					Items: []models.CartItem{
						{
							ProductID: "prod-123",
							Quantity:  2,
							Price:     99.99,
						},
					},
					Total: 199.98,
				}
				cartRepo.On("GetCart", "session-123").Return(existingCart, nil)
			},
			expectedError: "item not found in cart",
		},
		{
			name:      "cart not found",
			sessionID: "nonexistent",
			productID: "prod-123",
			quantity:  1,
			setupMocks: func(cartRepo *MockCartRepository, prodRepo *MockProductRepository) {
				cartRepo.On("GetCart", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "cart not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCartRepo := new(MockCartRepository)
			mockProdRepo := new(MockProductRepository)
			tt.setupMocks(mockCartRepo, mockProdRepo)

			service := &Service{
				cartRepo:    mockCartRepo,
				productRepo: mockProdRepo,
			}

			err := service.UpdateCartItem(context.Background(), tt.sessionID, tt.productID, tt.quantity)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockCartRepo.AssertExpectations(t)
			mockProdRepo.AssertExpectations(t)
		})
	}
}

func TestCartService_ClearCart(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		setupMock     func(*MockCartRepository)
		expectedError string
	}{
		{
			name:      "clear existing cart",
			sessionID: "session-123",
			setupMock: func(repo *MockCartRepository) {
				repo.On("DeleteCart", "session-123").Return(nil)
			},
		},
		{
			name:          "invalid session ID",
			sessionID:     "",
			setupMock:     func(repo *MockCartRepository) {},
			expectedError: "session ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCartRepository)
			tt.setupMock(mockRepo)

			service := &Service{
				cartRepo: mockRepo,
			}

			err := service.ClearCart(context.Background(), tt.sessionID)

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

func TestCartService_CalculateTotal(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name          string
		items         []models.CartItem
		expectedTotal float64
	}{
		{
			name: "calculate total for multiple items",
			items: []models.CartItem{
				{ProductID: "prod-1", Quantity: 2, Price: 99.99},
				{ProductID: "prod-2", Quantity: 1, Price: 49.99},
				{ProductID: "prod-3", Quantity: 3, Price: 19.99},
			},
			expectedTotal: 309.95, // (2 * 99.99) + (1 * 49.99) + (3 * 19.99)
		},
		{
			name:          "calculate total for empty cart",
			items:         []models.CartItem{},
			expectedTotal: 0.0,
		},
		{
			name: "calculate total for single item",
			items: []models.CartItem{
				{ProductID: "prod-1", Quantity: 5, Price: 25.50},
			},
			expectedTotal: 127.50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := service.CalculateTotal(tt.items)
			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}
