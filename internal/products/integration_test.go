package products

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-website/internal/models"
	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ProductIntegrationTestSuite struct {
	suite.Suite
	db      *gorm.DB
	router  *gin.Engine
	service *Service
	handler *Handler
}

func (suite *ProductIntegrationTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.Category{}, &models.Product{}, &models.User{}, &models.Order{}, &models.OrderItem{})
	suite.Require().NoError(err)

	suite.db = db
	// Initialize service and handler
	suite.service = NewService(db)
	suite.handler = NewHandler(suite.service)

	// Setup router
	suite.router = gin.New()
	SetupRoutes(suite.router, suite.handler)
}

func (suite *ProductIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func (suite *ProductIntegrationTestSuite) SetupTest() {
	// Clean up tables before each test
	suite.db.Exec("DELETE FROM order_items")
	suite.db.Exec("DELETE FROM orders")
	suite.db.Exec("DELETE FROM products")
	suite.db.Exec("DELETE FROM categories")
	suite.db.Exec("DELETE FROM users")
}

func (suite *ProductIntegrationTestSuite) seedTestData() {
	// Create categories
	categories := []models.Category{
		{
			Name:      "Electronics",
			Slug:      "electronics",
			IsActive:  true,
			SortOrder: 0,
		},
		{
			Name:      "Books",
			Slug:      "books",
			IsActive:  true,
			SortOrder: 1,
		},
		{
			Name:      "Clothing",
			Slug:      "clothing",
			IsActive:  true,
			SortOrder: 2,
		},
	}

	var categoryIDs []string
	for _, category := range categories {
		suite.db.Create(&category)
		categoryIDs = append(categoryIDs, category.ID)
	}

	// Create products
	products := []models.Product{
		{
			Name:        "Gaming Laptop",
			Description: "High performance gaming laptop with RTX graphics",
			Price:       1299.99,
			SKU:         "LAPTOP-001",
			Inventory:   5,
			IsActive:    true,
			CategoryID:  categoryIDs[0], // Electronics
			Images:      models.StringArray{"laptop1.jpg", "laptop2.jpg"},
		},
		{
			Name:        "Smartphone",
			Description: "Latest smartphone with advanced camera",
			Price:       799.99,
			SKU:         "PHONE-001",
			Inventory:   10,
			IsActive:    true,
			CategoryID:  categoryIDs[0], // Electronics
			Images:      models.StringArray{"phone1.jpg"},
		},
		{
			Name:        "Programming Book",
			Description: "Learn programming with this comprehensive guide",
			Price:       49.99,
			SKU:         "BOOK-001",
			Inventory:   20,
			IsActive:    true,
			CategoryID:  categoryIDs[1], // Books
			Images:      models.StringArray{"book1.jpg"},
		},
		{
			Name:        "Cotton T-Shirt",
			Description: "Comfortable cotton t-shirt",
			Price:       29.99,
			SKU:         "SHIRT-001",
			Inventory:   0, // Out of stock
			IsActive:    true,
			CategoryID:  categoryIDs[2], // Clothing
			Images:      models.StringArray{"shirt1.jpg"},
		},
		{
			Name:        "Inactive Product",
			Description: "This product is inactive",
			Price:       99.99,
			SKU:         "INACTIVE-001",
			Inventory:   5,
			IsActive:    false, // Inactive
			CategoryID:  categoryIDs[0], // Electronics
			Images:      models.StringArray{},
		},
	}

	for i, product := range products {
		suite.db.Create(&product)
		// Explicitly set inactive product after creation to override GORM default
		if i == 4 { // The inactive product
			suite.db.Model(&product).Update("is_active", false)
		}
	}
}

func (suite *ProductIntegrationTestSuite) TestFullProductCatalogWorkflow() {
	// Seed test data
	suite.seedTestData()

	// Test 1: Get all products
	req, _ := http.NewRequest("GET", "/api/products", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	data := response.Data.(map[string]interface{})
	products := data["products"].([]interface{})
	assert.Len(suite.T(), products, 4) // Only active products (inactive should be filtered out)

	// Test 2: Get categories first to get electronics category ID
	req2, _ := http.NewRequest("GET", "/api/categories", nil)
	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusOK, w2.Code)
	var categoriesResponse utils.ApiResponse
	err = json.Unmarshal(w2.Body.Bytes(), &categoriesResponse)
	assert.NoError(suite.T(), err)

	categoriesData := categoriesResponse.Data.(map[string]interface{})
	categories := categoriesData["categories"].([]interface{})
	electronicsCategory := categories[0].(map[string]interface{})
	electronicsID := electronicsCategory["id"].(string)
	
	req, _ = http.NewRequest("GET", "/api/products?category_id="+electronicsID, nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data = response.Data.(map[string]interface{})
	products = data["products"].([]interface{})
	assert.Len(suite.T(), products, 2) // Laptop and Phone

	// Test 3: Filter by price range
	req, _ = http.NewRequest("GET", "/api/products?min_price=50&max_price=800", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data = response.Data.(map[string]interface{})
	products = data["products"].([]interface{})
	assert.Len(suite.T(), products, 1) // Only Phone (799.99)

	// Test 4: Filter by in stock
	req, _ = http.NewRequest("GET", "/api/products?in_stock=true", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data = response.Data.(map[string]interface{})
	products = data["products"].([]interface{})
	assert.Len(suite.T(), products, 3) // Laptop, Phone, Book (shirt is out of stock)

	// Test 5: Search products
	req, _ = http.NewRequest("GET", "/api/products/search?q=gaming", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data = response.Data.(map[string]interface{})
	products = data["products"].([]interface{})
	assert.Len(suite.T(), products, 1) // Gaming Laptop

	// Test 6: Get specific product (get the first product from the list)
	firstProduct := products[0].(map[string]interface{})
	productID := firstProduct["id"].(string)
	
	req, _ = http.NewRequest("GET", "/api/products/"+productID, nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	productData := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), productID, productData["id"])
	assert.Equal(suite.T(), "Gaming Laptop", productData["name"])
	
	// Check category is preloaded
	category := productData["category"].(map[string]interface{})
	assert.Equal(suite.T(), "Electronics", category["name"])

	// Test 7: Get all categories
	req, _ = http.NewRequest("GET", "/api/categories", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data = response.Data.(map[string]interface{})
	categories = data["categories"].([]interface{})
	assert.Len(suite.T(), categories, 3)
	assert.Equal(suite.T(), float64(3), data["total"])

	// Test 8: Get specific category (get the first category from the list)
	firstCategoryForTest := categories[0].(map[string]interface{})
	categoryID := firstCategoryForTest["id"].(string)
	
	req, _ = http.NewRequest("GET", "/api/categories/"+categoryID, nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	categoryData := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), categoryID, categoryData["id"])
	assert.Equal(suite.T(), "Electronics", categoryData["name"])
}

func (suite *ProductIntegrationTestSuite) TestPaginationWorkflow() {
	// Create many products for pagination testing
	
	// Create a category first
	category := models.Category{
		Name:     "Test Category",
		Slug:     "test",
		IsActive: true,
	}
	suite.db.Create(&category)

	// Create 25 products
	for i := 1; i <= 25; i++ {
		product := models.Product{
			Name:       fmt.Sprintf("Product %03d", i),
			Price:      float64(i * 10),
			SKU:        fmt.Sprintf("SKU-%03d", i),
			Inventory:  10,
			IsActive:   true,
			CategoryID: category.ID,
		}
		suite.db.Create(&product)
	}

	// Test first page
	req, _ := http.NewRequest("GET", "/api/products?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data := response.Data.(map[string]interface{})
	products := data["products"].([]interface{})
	assert.Len(suite.T(), products, 10)
	assert.Equal(suite.T(), float64(25), data["total"])
	assert.Equal(suite.T(), float64(3), data["totalPages"])
	assert.True(suite.T(), data["hasNext"].(bool))
	assert.False(suite.T(), data["hasPrevious"].(bool))

	// Test middle page
	req, _ = http.NewRequest("GET", "/api/products?page=2&page_size=10", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data = response.Data.(map[string]interface{})
	products = data["products"].([]interface{})
	assert.Len(suite.T(), products, 10)
	assert.True(suite.T(), data["hasNext"].(bool))
	assert.True(suite.T(), data["hasPrevious"].(bool))

	// Test last page
	req, _ = http.NewRequest("GET", "/api/products?page=3&page_size=10", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data = response.Data.(map[string]interface{})
	products = data["products"].([]interface{})
	assert.Len(suite.T(), products, 5) // Remaining 5 products
	assert.False(suite.T(), data["hasNext"].(bool))
	assert.True(suite.T(), data["hasPrevious"].(bool))
}

func (suite *ProductIntegrationTestSuite) TestSortingWorkflow() {
	suite.seedTestData()

	// Test sorting by price ascending
	req, _ := http.NewRequest("GET", "/api/products?sort_by=price&sort_order=asc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data := response.Data.(map[string]interface{})
	products := data["products"].([]interface{})
	
	// Check that products are sorted by price ascending
	firstProduct := products[0].(map[string]interface{})
	lastProduct := products[len(products)-1].(map[string]interface{})
	
	assert.True(suite.T(), firstProduct["price"].(float64) <= lastProduct["price"].(float64))

	// Test sorting by name
	req, _ = http.NewRequest("GET", "/api/products?sort_by=name&sort_order=asc", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data = response.Data.(map[string]interface{})
	products = data["products"].([]interface{})
	
	// Check that products are sorted by name
	firstProduct = products[0].(map[string]interface{})
	assert.Equal(suite.T(), "Cotton T-Shirt", firstProduct["name"]) // Alphabetically first
}

func TestProductIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ProductIntegrationTestSuite))
}