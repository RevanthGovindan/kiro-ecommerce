package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ecommerce-website/internal/models"
)

// TestDatabase provides utilities for test database operations
type TestDatabase struct {
	DB *gorm.DB
}

// NewTestDatabase creates a new test database instance
func NewTestDatabase(t *testing.T) *TestDatabase {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Cart{},
		&models.CartItem{},
	)
	require.NoError(t, err)

	return &TestDatabase{DB: db}
}

// Cleanup closes the database connection
func (td *TestDatabase) Cleanup() {
	sqlDB, _ := td.DB.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

// SeedTestData populates the database with test data
func (td *TestDatabase) SeedTestData(t *testing.T) {
	// Create test categories
	categories := []models.Category{
		{ID: "cat-1", Name: "Electronics", Slug: "electronics", IsActive: true},
		{ID: "cat-2", Name: "Clothing", Slug: "clothing", IsActive: true},
		{ID: "cat-3", Name: "Books", Slug: "books", IsActive: true},
	}
	for _, cat := range categories {
		require.NoError(t, td.DB.Create(&cat).Error)
	}

	// Create test products
	products := []models.Product{
		{
			ID:          "prod-1",
			Name:        "Gaming Laptop",
			Description: "High-performance gaming laptop",
			Price:       1299.99,
			SKU:         "GAM-LAP-001",
			Inventory:   10,
			CategoryID:  "cat-1",
			IsActive:    true,
		},
		{
			ID:          "prod-2",
			Name:        "Cotton T-Shirt",
			Description: "Comfortable cotton t-shirt",
			Price:       29.99,
			SKU:         "COT-TSH-001",
			Inventory:   50,
			CategoryID:  "cat-2",
			IsActive:    true,
		},
		{
			ID:          "prod-3",
			Name:        "Programming Book",
			Description: "Learn programming fundamentals",
			Price:       49.99,
			SKU:         "PRG-BOK-001",
			Inventory:   25,
			CategoryID:  "cat-3",
			IsActive:    true,
		},
	}
	for _, prod := range products {
		require.NoError(t, td.DB.Create(&prod).Error)
	}

	// Create test users
	users := []models.User{
		{
			ID:        "user-1",
			Email:     "customer@example.com",
			Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
			FirstName: "John",
			LastName:  "Doe",
			Role:      "customer",
			IsActive:  true,
		},
		{
			ID:        "user-2",
			Email:     "admin@example.com",
			Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
			FirstName: "Admin",
			LastName:  "User",
			Role:      "admin",
			IsActive:  true,
		},
	}
	for _, user := range users {
		require.NoError(t, td.DB.Create(&user).Error)
	}
}

// APITestHelper provides utilities for API testing
type APITestHelper struct {
	Router *gin.Engine
	DB     *gorm.DB
}

// NewAPITestHelper creates a new API test helper
func NewAPITestHelper(router *gin.Engine, db *gorm.DB) *APITestHelper {
	gin.SetMode(gin.TestMode)
	return &APITestHelper{
		Router: router,
		DB:     db,
	}
}

// MakeRequest makes an HTTP request and returns the response
func (h *APITestHelper) MakeRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, url, reqBody)

	// Set default content type
	req.Header.Set("Content-Type", "application/json")

	// Set additional headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	h.Router.ServeHTTP(w, req)

	return w
}

// MakeAuthenticatedRequest makes an authenticated HTTP request
func (h *APITestHelper) MakeAuthenticatedRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	return h.MakeRequest(method, url, body, headers)
}

// ParseJSONResponse parses JSON response body
func (h *APITestHelper) ParseJSONResponse(w *httptest.ResponseRecorder, target interface{}) error {
	return json.Unmarshal(w.Body.Bytes(), target)
}

// AssertJSONResponse asserts that the response is valid JSON and matches expected structure
func (h *APITestHelper) AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	require.Equal(t, expectedStatus, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return response
}

// AssertSuccessResponse asserts that the response indicates success
func (h *APITestHelper) AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	response := h.AssertJSONResponse(t, w, expectedStatus)
	require.True(t, response["success"].(bool))
	require.NotNil(t, response["data"])
	return response
}

// AssertErrorResponse asserts that the response indicates an error
func (h *APITestHelper) AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	response := h.AssertJSONResponse(t, w, expectedStatus)
	require.False(t, response["success"].(bool))
	require.NotNil(t, response["error"])
	return response
}

// TestTimer provides utilities for timing operations in tests
type TestTimer struct {
	start time.Time
}

// NewTestTimer creates a new test timer
func NewTestTimer() *TestTimer {
	return &TestTimer{start: time.Now()}
}

// Elapsed returns the elapsed time since the timer was created
func (tt *TestTimer) Elapsed() time.Duration {
	return time.Since(tt.start)
}

// AssertDuration asserts that the elapsed time is within expected bounds
func (tt *TestTimer) AssertDuration(t *testing.T, min, max time.Duration) {
	elapsed := tt.Elapsed()
	require.True(t, elapsed >= min, "Operation completed too quickly: %v < %v", elapsed, min)
	require.True(t, elapsed <= max, "Operation took too long: %v > %v", elapsed, max)
}

// TestDataFactory provides utilities for creating test data
type TestDataFactory struct{}

// NewTestDataFactory creates a new test data factory
func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{}
}

// CreateTestUser creates a test user with default values
func (f *TestDataFactory) CreateTestUser(overrides map[string]interface{}) models.User {
	user := models.User{
		Email:     "test@example.com",
		Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
		FirstName: "Test",
		LastName:  "User",
		Role:      "customer",
		IsActive:  true,
	}

	// Apply overrides
	if email, ok := overrides["email"].(string); ok {
		user.Email = email
	}
	if firstName, ok := overrides["firstName"].(string); ok {
		user.FirstName = firstName
	}
	if lastName, ok := overrides["lastName"].(string); ok {
		user.LastName = lastName
	}
	if role, ok := overrides["role"].(string); ok {
		user.Role = role
	}
	if isActive, ok := overrides["isActive"].(bool); ok {
		user.IsActive = isActive
	}

	return user
}

// CreateTestProduct creates a test product with default values
func (f *TestDataFactory) CreateTestProduct(overrides map[string]interface{}) models.Product {
	product := models.Product{
		Name:        "Test Product",
		Description: "Test product description",
		Price:       99.99,
		SKU:         fmt.Sprintf("TEST-%d", time.Now().UnixNano()),
		Inventory:   10,
		CategoryID:  "cat-1",
		IsActive:    true,
	}

	// Apply overrides
	if name, ok := overrides["name"].(string); ok {
		product.Name = name
	}
	if description, ok := overrides["description"].(string); ok {
		product.Description = description
	}
	if price, ok := overrides["price"].(float64); ok {
		product.Price = price
	}
	if sku, ok := overrides["sku"].(string); ok {
		product.SKU = sku
	}
	if inventory, ok := overrides["inventory"].(int); ok {
		product.Inventory = inventory
	}
	if categoryID, ok := overrides["categoryId"].(string); ok {
		product.CategoryID = categoryID
	}
	if isActive, ok := overrides["isActive"].(bool); ok {
		product.IsActive = isActive
	}

	return product
}

// CreateTestOrder creates a test order with default values
func (f *TestDataFactory) CreateTestOrder(overrides map[string]interface{}) models.Order {
	order := models.Order{
		UserID:   "user-1",
		Status:   "pending",
		Subtotal: 99.99,
		Tax:      8.00,
		Shipping: 5.99,
		Total:    113.98,
	}

	// Apply overrides
	if userID, ok := overrides["userId"].(string); ok {
		order.UserID = userID
	}
	if status, ok := overrides["status"].(string); ok {
		order.Status = status
	}
	if subtotal, ok := overrides["subtotal"].(float64); ok {
		order.Subtotal = subtotal
	}
	if tax, ok := overrides["tax"].(float64); ok {
		order.Tax = tax
	}
	if shipping, ok := overrides["shipping"].(float64); ok {
		order.Shipping = shipping
	}
	if total, ok := overrides["total"].(float64); ok {
		order.Total = total
	}

	return order
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}

	return false
}
