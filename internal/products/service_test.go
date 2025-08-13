package products

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"ecommerce-website/internal/models"
)

// MockProductRepository is a mock implementation of the product repository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(id string) (*models.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetBySKU(sku string) (*models.Product, error) {
	args := m.Called(sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) List(filters ProductFilters) ([]*models.Product, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]*models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) Update(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductRepository) Search(query string, filters ProductFilters) ([]*models.Product, int64, error) {
	args := m.Called(query, filters)
	return args.Get(0).([]*models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) UpdateInventory(id string, quantity int) error {
	args := m.Called(id, quantity)
	return args.Error(0)
}

func TestProductService_CreateProduct(t *testing.T) {
	tests := []struct {
		name          string
		product       *models.Product
		setupMock     func(*MockProductRepository)
		expectedError string
	}{
		{
			name: "successful product creation",
			product: &models.Product{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				SKU:         "TEST-001",
				Inventory:   10,
				CategoryID:  "cat-123",
				IsActive:    true,
			},
			setupMock: func(repo *MockProductRepository) {
				repo.On("GetBySKU", "TEST-001").Return(nil, gorm.ErrRecordNotFound)
				repo.On("Create", mock.AnythingOfType("*models.Product")).Return(nil)
			},
		},
		{
			name: "duplicate SKU",
			product: &models.Product{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				SKU:         "EXISTING-001",
				Inventory:   10,
				CategoryID:  "cat-123",
				IsActive:    true,
			},
			setupMock: func(repo *MockProductRepository) {
				existingProduct := &models.Product{SKU: "EXISTING-001"}
				repo.On("GetBySKU", "EXISTING-001").Return(existingProduct, nil)
			},
			expectedError: "product with SKU already exists",
		},
		{
			name: "invalid price",
			product: &models.Product{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       -10.00,
				SKU:         "TEST-002",
				Inventory:   10,
				CategoryID:  "cat-123",
				IsActive:    true,
			},
			setupMock:     func(repo *MockProductRepository) {},
			expectedError: "price must be greater than 0",
		},
		{
			name: "empty name",
			product: &models.Product{
				Name:        "",
				Description: "Test Description",
				Price:       99.99,
				SKU:         "TEST-003",
				Inventory:   10,
				CategoryID:  "cat-123",
				IsActive:    true,
			},
			setupMock:     func(repo *MockProductRepository) {},
			expectedError: "product name is required",
		},
		{
			name: "empty SKU",
			product: &models.Product{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				SKU:         "",
				Inventory:   10,
				CategoryID:  "cat-123",
				IsActive:    true,
			},
			setupMock:     func(repo *MockProductRepository) {},
			expectedError: "SKU is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			tt.setupMock(mockRepo)

			service := &Service{
				productRepo: mockRepo,
			}

			err := service.CreateProduct(context.Background(), tt.product)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.product.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_GetProduct(t *testing.T) {
	tests := []struct {
		name          string
		productID     string
		setupMock     func(*MockProductRepository)
		expectedError string
	}{
		{
			name:      "successful product retrieval",
			productID: "prod-123",
			setupMock: func(repo *MockProductRepository) {
				product := &models.Product{
					ID:          "prod-123",
					Name:        "Test Product",
					Description: "Test Description",
					Price:       99.99,
					SKU:         "TEST-001",
					Inventory:   10,
					IsActive:    true,
				}
				repo.On("GetByID", "prod-123").Return(product, nil)
			},
		},
		{
			name:      "product not found",
			productID: "nonexistent",
			setupMock: func(repo *MockProductRepository) {
				repo.On("GetByID", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "product not found",
		},
		{
			name:          "empty product ID",
			productID:     "",
			setupMock:     func(repo *MockProductRepository) {},
			expectedError: "product ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			tt.setupMock(mockRepo)

			service := &Service{
				productRepo: mockRepo,
			}

			product, err := service.GetProduct(context.Background(), tt.productID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, product)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, tt.productID, product.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	tests := []struct {
		name          string
		productID     string
		updates       *models.Product
		setupMock     func(*MockProductRepository)
		expectedError string
	}{
		{
			name:      "successful product update",
			productID: "prod-123",
			updates: &models.Product{
				Name:        "Updated Product",
				Description: "Updated Description",
				Price:       149.99,
			},
			setupMock: func(repo *MockProductRepository) {
				existingProduct := &models.Product{
					ID:          "prod-123",
					Name:        "Test Product",
					Description: "Test Description",
					Price:       99.99,
					SKU:         "TEST-001",
					Inventory:   10,
					IsActive:    true,
				}
				repo.On("GetByID", "prod-123").Return(existingProduct, nil)
				repo.On("Update", mock.AnythingOfType("*models.Product")).Return(nil)
			},
		},
		{
			name:      "product not found",
			productID: "nonexistent",
			updates: &models.Product{
				Name: "Updated Product",
			},
			setupMock: func(repo *MockProductRepository) {
				repo.On("GetByID", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "product not found",
		},
		{
			name:      "invalid price update",
			productID: "prod-123",
			updates: &models.Product{
				Price: -50.00,
			},
			setupMock: func(repo *MockProductRepository) {
				existingProduct := &models.Product{
					ID:          "prod-123",
					Name:        "Test Product",
					Description: "Test Description",
					Price:       99.99,
					SKU:         "TEST-001",
					Inventory:   10,
					IsActive:    true,
				}
				repo.On("GetByID", "prod-123").Return(existingProduct, nil)
			},
			expectedError: "price must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			tt.setupMock(mockRepo)

			service := &Service{
				productRepo: mockRepo,
			}

			err := service.UpdateProduct(context.Background(), tt.productID, tt.updates)

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

func TestProductService_UpdateInventory(t *testing.T) {
	tests := []struct {
		name          string
		productID     string
		quantity      int
		setupMock     func(*MockProductRepository)
		expectedError string
	}{
		{
			name:      "successful inventory update",
			productID: "prod-123",
			quantity:  5,
			setupMock: func(repo *MockProductRepository) {
				product := &models.Product{
					ID:        "prod-123",
					Inventory: 10,
					IsActive:  true,
				}
				repo.On("GetByID", "prod-123").Return(product, nil)
				repo.On("UpdateInventory", "prod-123", 5).Return(nil)
			},
		},
		{
			name:      "insufficient inventory",
			productID: "prod-123",
			quantity:  -15,
			setupMock: func(repo *MockProductRepository) {
				product := &models.Product{
					ID:        "prod-123",
					Inventory: 10,
					IsActive:  true,
				}
				repo.On("GetByID", "prod-123").Return(product, nil)
			},
			expectedError: "insufficient inventory",
		},
		{
			name:      "product not found",
			productID: "nonexistent",
			quantity:  5,
			setupMock: func(repo *MockProductRepository) {
				repo.On("GetByID", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "product not found",
		},
		{
			name:      "inactive product",
			productID: "prod-123",
			quantity:  5,
			setupMock: func(repo *MockProductRepository) {
				product := &models.Product{
					ID:        "prod-123",
					Inventory: 10,
					IsActive:  false,
				}
				repo.On("GetByID", "prod-123").Return(product, nil)
			},
			expectedError: "product is not active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			tt.setupMock(mockRepo)

			service := &Service{
				productRepo: mockRepo,
			}

			err := service.UpdateInventory(context.Background(), tt.productID, tt.quantity)

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

func TestProductService_SearchProducts(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		filters       ProductFilters
		setupMock     func(*MockProductRepository)
		expectedCount int
	}{
		{
			name:  "successful search",
			query: "laptop",
			filters: ProductFilters{
				Page:     1,
				PageSize: 10,
			},
			setupMock: func(repo *MockProductRepository) {
				products := []*models.Product{
					{ID: "prod-1", Name: "Gaming Laptop"},
					{ID: "prod-2", Name: "Business Laptop"},
				}
				repo.On("Search", "laptop", mock.AnythingOfType("ProductFilters")).Return(products, int64(2), nil)
			},
			expectedCount: 2,
		},
		{
			name:  "no results",
			query: "nonexistent",
			filters: ProductFilters{
				Page:     1,
				PageSize: 10,
			},
			setupMock: func(repo *MockProductRepository) {
				products := []*models.Product{}
				repo.On("Search", "nonexistent", mock.AnythingOfType("ProductFilters")).Return(products, int64(0), nil)
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepository)
			tt.setupMock(mockRepo)

			service := &Service{
				productRepo: mockRepo,
			}

			products, total, err := service.SearchProducts(context.Background(), tt.query, tt.filters)

			assert.NoError(t, err)
			assert.Len(t, products, tt.expectedCount)
			assert.Equal(t, int64(tt.expectedCount), total)

			mockRepo.AssertExpectations(t)
		})
	}
}
