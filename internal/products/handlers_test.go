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
)

type ProductHandlerTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *Service
	handler *Handler
	router  *gin.Engine
}

func (suite *ProductHandlerTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.Category{}, &models.Product{}, &models.User{}, &models.Order{}, &models.OrderItem{})
	suite.Require().NoError(err)

	suite.db = db
	suite.service = NewService(db)
	suite.handler = NewHandler(suite.service)

	// Setup router
	suite.router = gin.New()
	SetupRoutes(suite.router, suite.handler)
}

func (suite *ProductHandlerTestSuite) SetupTest() {
	// Clean up tables before each test
	suite.db.Exec("DELETE FROM order_items")
	suite.db.Exec("DELETE FROM orders")
	suite.db.Exec("DELETE FROM products")
	suite.db.Exec("DELETE FROM categories")
	suite.db.Exec("DELETE FROM users")
}

func (suite *ProductHandlerTestSuite) createTestCategory() *models.Category {
	category := &models.Category{
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
	}
	suite.db.Create(category)
	return category
}

func (suite *ProductHandlerTestSuite) createTestProduct(categoryID string, name string, price float64) *models.Product {
	product := &models.Product{
		Name:       name,
		Price:      price,
		SKU:        "SKU-" + name,
		Inventory:  10,
		IsActive:   true,
		CategoryID: categoryID,
		Images:     models.StringArray{"image1.jpg", "image2.jpg"},
	}
	suite.db.Create(product)
	return product
}

func (suite *ProductHandlerTestSuite) TestGetProducts_Success() {
	// Setup test data
	category := suite.createTestCategory()
	suite.createTestProduct(category.ID, "Laptop", 999.99)
	suite.createTestProduct(category.ID, "Phone", 599.99)

	// Make request
	req, _ := http.NewRequest("GET", "/api/products", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Check response data
	data := response.Data.(map[string]interface{})
	products := data["products"].([]interface{})
	assert.Len(suite.T(), products, 2)
	assert.Equal(suite.T(), float64(2), data["total"])
}

func (suite *ProductHandlerTestSuite) TestGetProducts_WithFilters() {
	// Setup test data
	category := suite.createTestCategory()
	suite.createTestProduct(category.ID, "Expensive", 1000.00)
	suite.createTestProduct(category.ID, "Cheap", 50.00)

	// Make request with price filter
	req, _ := http.NewRequest("GET", "/api/products?min_price=100&max_price=800", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), float64(0), data["total"]) // No products in range
}

func (suite *ProductHandlerTestSuite) TestGetProducts_WithPagination() {
	// Setup test data
	category := suite.createTestCategory()
	for i := 1; i <= 5; i++ {
		suite.createTestProduct(category.ID, fmt.Sprintf("Product%d", i), float64(i*100))
	}

	// Make request with pagination
	req, _ := http.NewRequest("GET", "/api/products?page=1&page_size=2", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	data := response.Data.(map[string]interface{})
	products := data["products"].([]interface{})
	assert.Len(suite.T(), products, 2)
	assert.Equal(suite.T(), float64(5), data["total"])
	assert.Equal(suite.T(), float64(3), data["totalPages"])
	assert.True(suite.T(), data["hasNext"].(bool))
	assert.False(suite.T(), data["hasPrevious"].(bool))
}

func (suite *ProductHandlerTestSuite) TestGetProductByID_Success() {
	// Setup test data
	category := suite.createTestCategory()
	product := suite.createTestProduct(category.ID, "TestProduct", 99.99)

	// Make request
	req, _ := http.NewRequest("GET", "/api/products/"+product.ID, nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Check product data
	productData := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), product.ID, productData["id"])
	assert.Equal(suite.T(), "TestProduct", productData["name"])
}

func (suite *ProductHandlerTestSuite) TestGetProductByID_NotFound() {
	// Make request for non-existent product
	req, _ := http.NewRequest("GET", "/api/products/non-existent", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Equal(suite.T(), "Product not found", response.Error.Message)
}

func (suite *ProductHandlerTestSuite) TestGetProductByID_EmptyID() {
	// Make request with empty ID - this will redirect to products list
	req, _ := http.NewRequest("GET", "/api/products/", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Gin redirects trailing slash to non-trailing slash
	assert.Equal(suite.T(), http.StatusMovedPermanently, w.Code)
}

func (suite *ProductHandlerTestSuite) TestSearchProducts_Success() {
	// Setup test data
	category := suite.createTestCategory()
	laptop := suite.createTestProduct(category.ID, "Gaming Laptop", 999.99)
	laptop.Description = "High performance gaming laptop"
	suite.db.Save(laptop)
	
	suite.createTestProduct(category.ID, "Phone", 599.99)

	// Make search request
	req, _ := http.NewRequest("GET", "/api/products/search?q=gaming", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Check search results
	data := response.Data.(map[string]interface{})
	products := data["products"].([]interface{})
	assert.Len(suite.T(), products, 1)
	
	product := products[0].(map[string]interface{})
	assert.Equal(suite.T(), "Gaming Laptop", product["name"])
}

func (suite *ProductHandlerTestSuite) TestSearchProducts_EmptyQuery() {
	// Make search request without query
	req, _ := http.NewRequest("GET", "/api/products/search", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Equal(suite.T(), "Search query is required", response.Error.Message)
}

func (suite *ProductHandlerTestSuite) TestGetCategories_Success() {
	// Setup test data
	category1 := suite.createTestCategory()
	category2 := &models.Category{
		ID:        "cat-2",
		Name:      "Books",
		Slug:      "books",
		IsActive:  true,
		SortOrder: 1,
	}
	suite.db.Create(category2)

	// Make request
	req, _ := http.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Check categories data
	data := response.Data.(map[string]interface{})
	categories := data["categories"].([]interface{})
	assert.Len(suite.T(), categories, 2)
	assert.Equal(suite.T(), float64(2), data["total"])

	// Check first category
	firstCategory := categories[0].(map[string]interface{})
	assert.Equal(suite.T(), category1.ID, firstCategory["id"])
	assert.Equal(suite.T(), "Electronics", firstCategory["name"])
}

func (suite *ProductHandlerTestSuite) TestGetCategoryByID_Success() {
	// Setup test data
	category := suite.createTestCategory()

	// Make request
	req, _ := http.NewRequest("GET", "/api/categories/"+category.ID, nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Check category data
	categoryData := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), category.ID, categoryData["id"])
	assert.Equal(suite.T(), "Electronics", categoryData["name"])
}

func (suite *ProductHandlerTestSuite) TestGetCategoryByID_NotFound() {
	// Make request for non-existent category
	req, _ := http.NewRequest("GET", "/api/categories/non-existent", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Equal(suite.T(), "Category not found", response.Error.Message)
}

func TestProductHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProductHandlerTestSuite))
}