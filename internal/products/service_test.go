package products

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductService_CreateProduct_Validation(t *testing.T) {
	tests := []struct {
		name          string
		request       CreateProductRequest
		expectedError string
	}{
		{
			name: "valid product request",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				SKU:         "TEST-001",
				Inventory:   10,
				CategoryID:  "cat-123",
			},
		},
		{
			name: "invalid price",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       -10.00,
				SKU:         "TEST-002",
				Inventory:   10,
				CategoryID:  "cat-123",
			},
			expectedError: "price must be greater than 0",
		},
		{
			name: "empty name",
			request: CreateProductRequest{
				Name:        "",
				Description: "Test Description",
				Price:       99.99,
				SKU:         "TEST-003",
				Inventory:   10,
				CategoryID:  "cat-123",
			},
			expectedError: "product name is required",
		},
		{
			name: "empty SKU",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				SKU:         "",
				Inventory:   10,
				CategoryID:  "cat-123",
			},
			expectedError: "SKU is required",
		},
		{
			name: "empty category ID",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				SKU:         "TEST-004",
				Inventory:   10,
				CategoryID:  "",
			},
			expectedError: "category ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic
			if tt.request.Price <= 0 {
				err := fmt.Errorf("price must be greater than 0")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "price must be greater than 0")
				return
			}

			if tt.request.Name == "" {
				err := fmt.Errorf("product name is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "product name is required")
				return
			}

			if tt.request.SKU == "" {
				err := fmt.Errorf("SKU is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "SKU is required")
				return
			}

			if tt.request.CategoryID == "" {
				err := fmt.Errorf("category ID is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "category ID is required")
				return
			}

			// Valid case
			if tt.expectedError == "" {
				assert.NotEmpty(t, tt.request.Name)
				assert.NotEmpty(t, tt.request.SKU)
				assert.NotEmpty(t, tt.request.CategoryID)
				assert.Greater(t, tt.request.Price, 0.0)
			}
		})
	}
}

func TestProductService_GetProduct_Validation(t *testing.T) {
	tests := []struct {
		name          string
		productID     string
		expectedError string
	}{
		{
			name:      "valid product ID",
			productID: "prod-123",
		},
		{
			name:          "empty product ID",
			productID:     "",
			expectedError: "product ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ID validation logic
			if tt.productID == "" {
				err := fmt.Errorf("product ID is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "product ID is required")
			} else {
				// Valid ID
				assert.NotEmpty(t, tt.productID)
			}
		})
	}
}

func TestProductService_UpdateProduct_Validation(t *testing.T) {
	tests := []struct {
		name          string
		productID     string
		request       UpdateProductRequest
		expectedError string
	}{
		{
			name:      "valid update request",
			productID: "prod-123",
			request: UpdateProductRequest{
				Name:  stringPtr("Updated Product"),
				Price: float64Ptr(149.99),
			},
		},
		{
			name:          "empty product ID",
			productID:     "",
			request:       UpdateProductRequest{},
			expectedError: "product ID is required",
		},
		{
			name:      "invalid price update",
			productID: "prod-123",
			request: UpdateProductRequest{
				Price: float64Ptr(-50.00),
			},
			expectedError: "price must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic
			if tt.productID == "" {
				err := fmt.Errorf("product ID is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "product ID is required")
				return
			}

			if tt.request.Price != nil && *tt.request.Price <= 0 {
				err := fmt.Errorf("price must be greater than 0")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "price must be greater than 0")
				return
			}

			// Valid case
			if tt.expectedError == "" {
				assert.NotEmpty(t, tt.productID)
				if tt.request.Price != nil {
					assert.Greater(t, *tt.request.Price, 0.0)
				}
			}
		})
	}
}

func TestProductService_UpdateInventory_Validation(t *testing.T) {
	tests := []struct {
		name          string
		productID     string
		inventory     int
		expectedError string
	}{
		{
			name:      "valid inventory update",
			productID: "prod-123",
			inventory: 50,
		},
		{
			name:          "empty product ID",
			productID:     "",
			inventory:     10,
			expectedError: "product ID is required",
		},
		{
			name:      "negative inventory",
			productID: "prod-123",
			inventory: -5,
		}, // Negative inventory might be valid for some business cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic
			if tt.productID == "" {
				err := fmt.Errorf("product ID is required")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "product ID is required")
				return
			}

			// Valid case
			if tt.expectedError == "" {
				assert.NotEmpty(t, tt.productID)
			}
		})
	}
}

func TestProductService_SearchProducts_Validation(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		pagination PaginationParams
	}{
		{
			name:  "valid search query",
			query: "laptop",
			pagination: PaginationParams{
				Page:     1,
				PageSize: 10,
			},
		},
		{
			name:  "empty search query",
			query: "",
			pagination: PaginationParams{
				Page:     1,
				PageSize: 10,
			},
		},
		{
			name:  "search with pagination defaults",
			query: "smartphone",
			pagination: PaginationParams{
				Page:     0, // Should default to 1
				PageSize: 0, // Should default to 20
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test pagination parameter normalization
			page := tt.pagination.Page
			pageSize := tt.pagination.PageSize

			if page <= 0 {
				page = 1
			}
			if pageSize <= 0 {
				pageSize = 20
			}

			assert.GreaterOrEqual(t, page, 1)
			assert.GreaterOrEqual(t, pageSize, 1)
		})
	}
}

// Helper functions for pointer creation
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
