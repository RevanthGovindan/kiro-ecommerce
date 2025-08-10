package products

import (
	"ecommerce-website/internal/models"

	"gorm.io/gorm"
)

// TestHelpers provides utility functions for testing
type TestHelpers struct {
	db *gorm.DB
}

// NewTestHelpers creates a new test helpers instance
func NewTestHelpers(db *gorm.DB) *TestHelpers {
	return &TestHelpers{db: db}
}

// CreateTestCategory creates a test category
func (h *TestHelpers) CreateTestCategory(id, name, slug string) *models.Category {
	category := &models.Category{
		ID:       id,
		Name:     name,
		Slug:     slug,
		IsActive: true,
	}
	h.db.Create(category)
	return category
}

// CreateTestProduct creates a test product
func (h *TestHelpers) CreateTestProduct(id, name, sku, categoryID string, price float64, inventory int) *models.Product {
	product := &models.Product{
		ID:         id,
		Name:       name,
		SKU:        sku,
		Price:      price,
		Inventory:  inventory,
		IsActive:   true,
		CategoryID: categoryID,
		Images:     models.StringArray{"test1.jpg", "test2.jpg"},
	}
	h.db.Create(product)
	return product
}

// CreateTestUser creates a test user
func (h *TestHelpers) CreateTestUser(id, email, firstName, lastName string) *models.User {
	user := &models.User{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  "hashedpassword",
		IsActive:  true,
		Role:      "customer",
	}
	h.db.Create(user)
	return user
}

// CleanupTables removes all data from test tables
func (h *TestHelpers) CleanupTables() {
	h.db.Exec("DELETE FROM order_items")
	h.db.Exec("DELETE FROM orders")
	h.db.Exec("DELETE FROM products")
	h.db.Exec("DELETE FROM categories")
	h.db.Exec("DELETE FROM addresses")
	h.db.Exec("DELETE FROM users")
}

// SeedBasicData creates basic test data
func (h *TestHelpers) SeedBasicData() {
	// Create categories
	h.CreateTestCategory("cat-electronics", "Electronics", "electronics")
	h.CreateTestCategory("cat-books", "Books", "books")
	h.CreateTestCategory("cat-clothing", "Clothing", "clothing")

	// Create products
	h.CreateTestProduct("prod-laptop", "Gaming Laptop", "LAPTOP-001", "cat-electronics", 1299.99, 5)
	h.CreateTestProduct("prod-phone", "Smartphone", "PHONE-001", "cat-electronics", 799.99, 10)
	h.CreateTestProduct("prod-book", "Programming Book", "BOOK-001", "cat-books", 49.99, 20)
	h.CreateTestProduct("prod-shirt", "Cotton T-Shirt", "SHIRT-001", "cat-clothing", 29.99, 0) // Out of stock
}

// CountProducts returns the number of active products
func (h *TestHelpers) CountProducts() int64 {
	var count int64
	h.db.Model(&models.Product{}).Where("is_active = ?", true).Count(&count)
	return count
}

// CountCategories returns the number of active categories
func (h *TestHelpers) CountCategories() int64 {
	var count int64
	h.db.Model(&models.Category{}).Where("is_active = ?", true).Count(&count)
	return count
}