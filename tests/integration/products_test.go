//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ecommerce-website/internal/models"
	"ecommerce-website/internal/products"
)

func setupProductsTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, func()) {
	gin.SetMode(gin.TestMode)

	db, cleanup := setupTestDB(t)

	// Create test categories
	categories := []models.Category{
		{ID: "cat-1", Name: "Electronics", Slug: "electronics", IsActive: true},
		{ID: "cat-2", Name: "Clothing", Slug: "clothing", IsActive: true},
	}
	for _, cat := range categories {
		require.NoError(t, db.Create(&cat).Error)
	}

	// Create products service
	productsService := products.NewService(db)

	// Setup router
	router := gin.New()
	productsGroup := router.Group("/api/products")
	products.RegisterRoutes(productsGroup, productsService)

	return router, db, cleanup
}

func TestProductsIntegration_GetProducts(t *testing.T) {
	router, db, cleanup := setupProductsTestRouter(t)
	defer cleanup()

	// Create test products
	testProducts := []models.Product{
		{
			ID:          "prod-1",
			Name:        "Laptop",
			Description: "Gaming laptop",
			Price:       999.99,
			SKU:         "LAP-001",
			Inventory:   10,
			CategoryID:  "cat-1",
			IsActive:    true,
		},
		{
			ID:          "prod-2",
			Name:        "T-Shirt",
			Description: "Cotton t-shirt",
			Price:       29.99,
			SKU:         "TSH-001",
			Inventory:   50,
			CategoryID:  "cat-2",
			IsActive:    true,
		},
		{
			ID:          "prod-3",
			Name:        "Inactive Product",
			Description: "This product is inactive",
			Price:       19.99,
			SKU:         "INA-001",
			Inventory:   5,
			CategoryID:  "cat-1",
			IsActive:    false,
		},
	}

	for _, product := range testProducts {
		require.NoError(t, db.Create(&product).Error)
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		checkActive    bool
	}{
		{
			name:           "get all active products",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // Only active products
			checkActive:    true,
		},
		{
			name:           "get products with pagination",
			queryParams:    "?page=1&page_size=1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkActive:    true,
		},
		{
			name:           "filter by category",
			queryParams:    "?category_id=cat-1",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Only active products in cat-1
			checkActive:    true,
		},
		{
			name:           "filter by price range",
			queryParams:    "?min_price=20&max_price=100",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // T-Shirt
			checkActive:    true,
		},
		{
			name:           "sort by price ascending",
			queryParams:    "?sort=price&order=asc",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkActive:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/products"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, response["success"].(bool))
			data := response["data"].(map[string]interface{})
			products := data["products"].([]interface{})
			assert.Len(t, products, tt.expectedCount)

			if tt.checkActive {
				for _, p := range products {
					product := p.(map[string]interface{})
					assert.True(t, product["isActive"].(bool))
				}
			}
		})
	}
}

func TestProductsIntegration_GetProductByID(t *testing.T) {
	router, db, cleanup := setupProductsTestRouter(t)
	defer cleanup()

	// Create test product
	testProduct := models.Product{
		ID:          "prod-123",
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		SKU:         "TEST-001",
		Inventory:   10,
		CategoryID:  "cat-1",
		IsActive:    true,
	}
	require.NoError(t, db.Create(&testProduct).Error)

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
	}{
		{
			name:           "get existing product",
			productID:      "prod-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existent product",
			productID:      "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/products/"+tt.productID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				product := data["product"].(map[string]interface{})
				assert.Equal(t, testProduct.ID, product["id"])
				assert.Equal(t, testProduct.Name, product["name"])
				assert.Equal(t, testProduct.Price, product["price"])
			} else {
				assert.False(t, response["success"].(bool))
				assert.NotNil(t, response["error"])
			}
		})
	}
}

func TestProductsIntegration_SearchProducts(t *testing.T) {
	router, db, cleanup := setupProductsTestRouter(t)
	defer cleanup()

	// Create test products
	testProducts := []models.Product{
		{
			ID:          "prod-1",
			Name:        "Gaming Laptop",
			Description: "High-performance gaming laptop",
			Price:       1299.99,
			SKU:         "GAM-LAP-001",
			Inventory:   5,
			CategoryID:  "cat-1",
			IsActive:    true,
		},
		{
			ID:          "prod-2",
			Name:        "Business Laptop",
			Description: "Professional laptop for business",
			Price:       899.99,
			SKU:         "BUS-LAP-001",
			Inventory:   8,
			CategoryID:  "cat-1",
			IsActive:    true,
		},
		{
			ID:          "prod-3",
			Name:        "Gaming Mouse",
			Description: "RGB gaming mouse",
			Price:       59.99,
			SKU:         "GAM-MOU-001",
			Inventory:   20,
			CategoryID:  "cat-1",
			IsActive:    true,
		},
	}

	for _, product := range testProducts {
		require.NoError(t, db.Create(&product).Error)
	}

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedCount  int
		expectedIDs    []string
	}{
		{
			name:           "search for laptop",
			query:          "laptop",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedIDs:    []string{"prod-1", "prod-2"},
		},
		{
			name:           "search for gaming",
			query:          "gaming",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedIDs:    []string{"prod-1", "prod-3"},
		},
		{
			name:           "search with no results",
			query:          "nonexistent",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedIDs:    []string{},
		},
		{
			name:           "empty search query",
			query:          "",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/products/search?q=%s", tt.query), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				products := data["products"].([]interface{})
				assert.Len(t, products, tt.expectedCount)

				if tt.expectedCount > 0 {
					foundIDs := make([]string, 0, len(products))
					for _, p := range products {
						product := p.(map[string]interface{})
						foundIDs = append(foundIDs, product["id"].(string))
					}
					assert.ElementsMatch(t, tt.expectedIDs, foundIDs)
				}
			} else {
				assert.False(t, response["success"].(bool))
			}
		})
	}
}

func TestProductsIntegration_AdminCreateProduct(t *testing.T) {
	router, db, cleanup := setupProductsTestRouter(t)
	defer cleanup()

	// Add admin routes (simplified for testing)
	adminGroup := router.Group("/api/admin/products")
	// Note: In real implementation, this would have admin authentication middleware
	productsService := products.NewService(db)
	products.RegisterAdminRoutes(adminGroup, productsService)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		checkDB        bool
	}{
		{
			name: "create product successfully",
			payload: map[string]interface{}{
				"name":        "New Product",
				"description": "New product description",
				"price":       199.99,
				"sku":         "NEW-001",
				"inventory":   15,
				"categoryId":  "cat-1",
				"isActive":    true,
			},
			expectedStatus: http.StatusCreated,
			checkDB:        true,
		},
		{
			name: "create product with duplicate SKU",
			payload: map[string]interface{}{
				"name":        "Another Product",
				"description": "Another description",
				"price":       299.99,
				"sku":         "NEW-001", // Same SKU as above
				"inventory":   10,
				"categoryId":  "cat-1",
				"isActive":    true,
			},
			expectedStatus: http.StatusConflict,
			checkDB:        false,
		},
		{
			name: "create product with invalid price",
			payload: map[string]interface{}{
				"name":        "Invalid Product",
				"description": "Invalid description",
				"price":       -10.00,
				"sku":         "INV-001",
				"inventory":   5,
				"categoryId":  "cat-1",
				"isActive":    true,
			},
			expectedStatus: http.StatusBadRequest,
			checkDB:        false,
		},
		{
			name: "create product with missing required fields",
			payload: map[string]interface{}{
				"description": "Missing name and SKU",
				"price":       99.99,
				"inventory":   5,
				"categoryId":  "cat-1",
			},
			expectedStatus: http.StatusBadRequest,
			checkDB:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusCreated {
				assert.True(t, response["success"].(bool))
				assert.NotNil(t, response["data"])

				if tt.checkDB {
					// Verify product was created in database
					var product models.Product
					err := db.Where("sku = ?", tt.payload["sku"]).First(&product).Error
					require.NoError(t, err)
					assert.Equal(t, tt.payload["name"], product.Name)
					assert.Equal(t, tt.payload["sku"], product.SKU)
					assert.Equal(t, tt.payload["price"], product.Price)
				}
			} else {
				assert.False(t, response["success"].(bool))
				assert.NotNil(t, response["error"])
			}
		})
	}
}
