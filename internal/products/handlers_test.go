package products

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-website/internal/auth"
	"ecommerce-website/internal/config"
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
	db          *gorm.DB
	service     *Service
	handler     *Handler
	authService *auth.Service
	router      *gin.Engine
	adminToken  string
	userToken   string
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

	// Setup auth service
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	suite.authService = auth.NewService(db, cfg)

	// Setup router
	suite.router = gin.New()
	SetupRoutes(suite.router, suite.handler, suite.authService)

	// Create test users and tokens
	suite.setupTestUsers()
}

func (suite *ProductHandlerTestSuite) setupTestUsers() {
	// Create admin user
	adminUser := &models.User{
		Email:     "admin@test.com",
		Password:  "$2a$10$test", // hashed password
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
		IsActive:  true,
	}
	suite.db.Create(adminUser)

	// Create regular user
	regularUser := &models.User{
		Email:     "user@test.com",
		Password:  "$2a$10$test", // hashed password
		FirstName: "Regular",
		LastName:  "User",
		Role:      "customer",
		IsActive:  true,
	}
	suite.db.Create(regularUser)

	// Generate tokens
	adminTokens, _ := suite.authService.GenerateTokens(adminUser)
	userTokens, _ := suite.authService.GenerateTokens(regularUser)

	suite.adminToken = adminTokens.AccessToken
	suite.userToken = userTokens.AccessToken
}

func (suite *ProductHandlerTestSuite) SetupTest() {
	// Clean up tables before each test (except users)
	suite.db.Exec("DELETE FROM order_items")
	suite.db.Exec("DELETE FROM orders")
	suite.db.Exec("DELETE FROM products")
	suite.db.Exec("DELETE FROM categories")
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

// Admin Product Management Tests

func (suite *ProductHandlerTestSuite) TestCreateProduct_Success() {
	// Setup test data
	category := suite.createTestCategory()

	// Create request body
	requestBody := CreateProductRequest{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		SKU:         "TEST-SKU-001",
		Inventory:   50,
		CategoryID:  category.ID,
		Images:      []string{"image1.jpg", "image2.jpg"},
		Specifications: map[string]interface{}{
			"color": "red",
			"size":  "large",
		},
	}

	jsonBody, _ := json.Marshal(requestBody)

	// Make request with admin token
	req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Check product data
	productData := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), "Test Product", productData["name"])
	assert.Equal(suite.T(), "TEST-SKU-001", productData["sku"])
	assert.Equal(suite.T(), 99.99, productData["price"])
}

func (suite *ProductHandlerTestSuite) TestCreateProduct_Unauthorized() {
	category := suite.createTestCategory()

	requestBody := CreateProductRequest{
		Name:       "Test Product",
		Price:      99.99,
		SKU:        "TEST-SKU-001",
		CategoryID: category.ID,
	}

	jsonBody, _ := json.Marshal(requestBody)

	// Make request without token
	req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *ProductHandlerTestSuite) TestCreateProduct_Forbidden() {
	category := suite.createTestCategory()

	requestBody := CreateProductRequest{
		Name:       "Test Product",
		Price:      99.99,
		SKU:        "TEST-SKU-001",
		CategoryID: category.ID,
	}

	jsonBody, _ := json.Marshal(requestBody)

	// Make request with user token (not admin)
	req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.userToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *ProductHandlerTestSuite) TestCreateProduct_DuplicateSKU() {
	category := suite.createTestCategory()
	suite.createTestProduct(category.ID, "Existing Product", 99.99)

	requestBody := CreateProductRequest{
		Name:       "New Product",
		Price:      199.99,
		SKU:        "SKU-Existing Product", // Same SKU as existing product
		CategoryID: category.ID,
	}

	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Equal(suite.T(), "Product with this SKU already exists", response.Error.Message)
}

func (suite *ProductHandlerTestSuite) TestUpdateProduct_Success() {
	category := suite.createTestCategory()
	product := suite.createTestProduct(category.ID, "Original Product", 99.99)

	// Update request
	updatePrice := 149.99
	updateName := "Updated Product"
	requestBody := UpdateProductRequest{
		Name:  &updateName,
		Price: &updatePrice,
	}

	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("PUT", "/api/admin/products/"+product.ID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Check updated data
	productData := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), "Updated Product", productData["name"])
	assert.Equal(suite.T(), 149.99, productData["price"])
}

func (suite *ProductHandlerTestSuite) TestUpdateProduct_NotFound() {
	updateName := "Updated Product"
	requestBody := UpdateProductRequest{
		Name: &updateName,
	}

	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("PUT", "/api/admin/products/non-existent", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *ProductHandlerTestSuite) TestDeleteProduct_Success() {
	category := suite.createTestCategory()
	product := suite.createTestProduct(category.ID, "Product to Delete", 99.99)

	req, _ := http.NewRequest("DELETE", "/api/admin/products/"+product.ID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Verify product is soft deleted
	var deletedProduct models.Product
	err = suite.db.Unscoped().Where("id = ?", product.ID).First(&deletedProduct).Error
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), deletedProduct.DeletedAt)
}

func (suite *ProductHandlerTestSuite) TestDeleteProduct_NotFound() {
	req, _ := http.NewRequest("DELETE", "/api/admin/products/non-existent", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *ProductHandlerTestSuite) TestUpdateInventory_Success() {
	category := suite.createTestCategory()
	product := suite.createTestProduct(category.ID, "Product", 99.99)

	requestBody := UpdateInventoryRequest{
		Inventory: 100,
	}

	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("PUT", "/api/admin/products/"+product.ID+"/inventory", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Check updated inventory
	productData := response.Data.(map[string]interface{})
	assert.Equal(suite.T(), float64(100), productData["inventory"])
}

func (suite *ProductHandlerTestSuite) TestUpdateInventory_NegativeValue() {
	category := suite.createTestCategory()
	product := suite.createTestProduct(category.ID, "Product", 99.99)

	requestBody := UpdateInventoryRequest{
		Inventory: -5,
	}

	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("PUT", "/api/admin/products/"+product.ID+"/inventory", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Equal(suite.T(), "Inventory cannot be negative", response.Error.Message)
}

func (suite *ProductHandlerTestSuite) TestGetAllProductsAdmin_Success() {
	category := suite.createTestCategory()

	// Create active product
	suite.createTestProduct(category.ID, "Active Product", 99.99)

	// Create inactive product
	inactiveProduct := suite.createTestProduct(category.ID, "Inactive Product", 199.99)
	inactiveProduct.IsActive = false
	suite.db.Save(inactiveProduct)

	req, _ := http.NewRequest("GET", "/api/admin/products", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Check that both active and inactive products are returned
	data := response.Data.(map[string]interface{})
	products := data["products"].([]interface{})
	assert.Len(suite.T(), products, 2)
	assert.Equal(suite.T(), float64(2), data["total"])
}

func (suite *ProductHandlerTestSuite) TestGetAllProductsAdmin_WithFilters() {
	category := suite.createTestCategory()

	// Create active product
	suite.createTestProduct(category.ID, "Active Product", 99.99)

	// Create inactive product
	inactiveProduct := suite.createTestProduct(category.ID, "Inactive Product", 199.99)
	inactiveProduct.IsActive = false
	suite.db.Save(inactiveProduct)

	// Filter for only active products
	req, _ := http.NewRequest("GET", "/api/admin/products?is_active=true", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Check that only active product is returned
	data := response.Data.(map[string]interface{})
	products := data["products"].([]interface{})
	assert.Len(suite.T(), products, 1)
	assert.Equal(suite.T(), float64(1), data["total"])

	product := products[0].(map[string]interface{})
	assert.Equal(suite.T(), "Active Product", product["name"])
}

func (suite *ProductHandlerTestSuite) TestCreateCategory_Success() {
	// Prepare request body
	reqBody := CreateCategoryRequest{
		Name:        "Electronics",
		Slug:        "electronics",
		Description: stringPtr("Electronic devices and accessories"),
		IsActive:    boolPtr(true),
		SortOrder:   intPtr(0),
	}

	body, _ := json.Marshal(reqBody)

	// Make request
	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Category created successfully", response.Message)

	// Verify category was created in database
	var category models.Category
	err = suite.db.Where("slug = ?", "electronics").First(&category).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Electronics", category.Name)
	assert.Equal(suite.T(), "electronics", category.Slug)
	assert.True(suite.T(), category.IsActive)
}

func (suite *ProductHandlerTestSuite) TestCreateCategory_ValidationError() {
	// Test missing required fields
	reqBody := CreateCategoryRequest{
		Name: "", // Missing name
		Slug: "test-slug",
	}

	body, _ := json.Marshal(reqBody)

	// Make request
	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	// Gin's binding validation catches this first
	assert.Equal(suite.T(), "INVALID_REQUEST", response.Error.Code)
}

func (suite *ProductHandlerTestSuite) TestCreateCategory_DuplicateSlug() {
	// Create existing category
	existingCategory := &models.Category{
		Name:     "Existing Category",
		Slug:     "electronics",
		IsActive: true,
	}
	suite.db.Create(existingCategory)

	// Try to create category with same slug
	reqBody := CreateCategoryRequest{
		Name: "New Electronics",
		Slug: "electronics", // Duplicate slug
	}

	body, _ := json.Marshal(reqBody)

	// Make request
	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Equal(suite.T(), "CATEGORY_EXISTS", response.Error.Code)
}

func (suite *ProductHandlerTestSuite) TestCreateCategory_WithParent() {
	// Create parent category
	parentCategory := &models.Category{
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
	}
	suite.db.Create(parentCategory)

	// Create child category
	reqBody := CreateCategoryRequest{
		Name:     "Smartphones",
		Slug:     "smartphones",
		ParentID: &parentCategory.ID,
	}

	body, _ := json.Marshal(reqBody)

	// Make request
	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// Verify category was created with parent
	var category models.Category
	err = suite.db.Preload("Parent").Where("slug = ?", "smartphones").First(&category).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Smartphones", category.Name)
	assert.NotNil(suite.T(), category.Parent)
	assert.Equal(suite.T(), "Electronics", category.Parent.Name)
}

func (suite *ProductHandlerTestSuite) TestCreateCategory_Unauthorized() {
	// Prepare request body
	reqBody := CreateCategoryRequest{
		Name: "Test Category",
		Slug: "test-category",
	}

	body, _ := json.Marshal(reqBody)

	// Make request without auth token
	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// Helper functions for tests (using existing ones from service_test.go)

func TestProductHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProductHandlerTestSuite))
}
