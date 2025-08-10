package users

import (
	"ecommerce-website/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestHelpers provides utility functions for testing
type TestHelpers struct {
	db *gorm.DB
}

// NewTestHelpers creates a new TestHelpers instance
func NewTestHelpers(db *gorm.DB) *TestHelpers {
	return &TestHelpers{db: db}
}

// CreateTestUser creates a test user in the database
func (h *TestHelpers) CreateTestUser(email, firstName, lastName string) *models.User {
	user := &models.User{
		Email:     email,
		Password:  "hashedpassword",
		FirstName: firstName,
		LastName:  lastName,
		Phone:     stringPtr("1234567890"),
		Role:      "customer",
		IsActive:  true,
	}
	h.db.Create(user)
	return user
}

// CreateTestAddress creates a test address for a user
func (h *TestHelpers) CreateTestAddress(userID, addressType string, isDefault bool) *models.Address {
	address := &models.Address{
		UserID:     userID,
		Type:       addressType,
		FirstName:  "John",
		LastName:   "Doe",
		Address1:   "123 Main St",
		City:       "New York",
		State:      "NY",
		PostalCode: "10001",
		Country:    "US",
		IsDefault:  isDefault,
	}
	h.db.Create(address)
	return address
}

// CreateTestOrder creates a test order for a user
func (h *TestHelpers) CreateTestOrder(userID, status string, total float64) *models.Order {
	order := &models.Order{
		UserID:   userID,
		Status:   status,
		Subtotal: total - 10, // Assume $10 tax/shipping
		Tax:      5.0,
		Shipping: 5.0,
		Total:    total,
	}
	h.db.Create(order)
	return order
}

// CleanupTestData removes all test data from the database
func (h *TestHelpers) CleanupTestData() {
	h.db.Exec("DELETE FROM order_items")
	h.db.Exec("DELETE FROM orders")
	h.db.Exec("DELETE FROM addresses")
	h.db.Exec("DELETE FROM users")
}

// GetUserByEmail retrieves a user by email for testing
func (h *TestHelpers) GetUserByEmail(email string) *models.User {
	var user models.User
	h.db.Where("email = ?", email).First(&user)
	return &user
}

// GetAddressesByUserID retrieves all addresses for a user
func (h *TestHelpers) GetAddressesByUserID(userID string) []models.Address {
	var addresses []models.Address
	h.db.Where("user_id = ?", userID).Find(&addresses)
	return addresses
}

// GetOrdersByUserID retrieves all orders for a user
func (h *TestHelpers) GetOrdersByUserID(userID string) []models.Order {
	var orders []models.Order
	h.db.Where("user_id = ?", userID).Find(&orders)
	return orders
}

// AssertUserExists checks if a user exists in the database
func (h *TestHelpers) AssertUserExists(userID string) bool {
	var count int64
	h.db.Model(&models.User{}).Where("id = ?", userID).Count(&count)
	return count > 0
}

// AssertAddressExists checks if an address exists in the database
func (h *TestHelpers) AssertAddressExists(addressID string) bool {
	var count int64
	h.db.Model(&models.Address{}).Where("id = ?", addressID).Count(&count)
	return count > 0
}

// AssertOrderExists checks if an order exists in the database
func (h *TestHelpers) AssertOrderExists(orderID string) bool {
	var count int64
	h.db.Model(&models.Order{}).Where("id = ?", orderID).Count(&count)
	return count > 0
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.Address{}, &models.Order{}, &models.OrderItem{})
	require.NoError(t, err)

	return db
}

// stringPtr is a helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
