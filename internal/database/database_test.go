package database

import (
	"testing"

	"ecommerce-website/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Address{},
		&models.Category{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
	)
	require.NoError(t, err)

	return db
}

func TestDatabaseConnection(t *testing.T) {
	db := setupTestDB(t)

	// Test basic query
	var count int64
	err := db.Model(&models.User{}).Count(&count).Error
	assert.NoError(t, err, "Should be able to query users table")
}

func TestModelsRelationships(t *testing.T) {
	db := setupTestDB(t)

	// Create test data
	// Create a category
	category := models.Category{
		ID:       "cat-1",
		Name:     "Smartphones",
		Slug:     "smartphones",
		IsActive: true,
	}
	err := db.Create(&category).Error
	require.NoError(t, err)

	// Create a product
	product := models.Product{
		ID:          "prod-1",
		Name:        "iPhone 15",
		Description: "Latest iPhone",
		Price:       999.99,
		SKU:         "IPHONE-15",
		Inventory:   10,
		CategoryID:  "cat-1",
		IsActive:    true,
	}
	err = db.Create(&product).Error
	require.NoError(t, err)

	// Create a user
	user := models.User{
		ID:        "user-1",
		Email:     "customer@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		IsActive:  true,
	}
	err = db.Create(&user).Error
	require.NoError(t, err)

	// Create an address
	address := models.Address{
		ID:         "addr-1",
		UserID:     "user-1",
		Type:       "shipping",
		FirstName:  "John",
		LastName:   "Doe",
		Address1:   "123 Main St",
		City:       "New York",
		State:      "NY",
		PostalCode: "10001",
		Country:    "US",
		IsDefault:  true,
	}
	err = db.Create(&address).Error
	require.NoError(t, err)

	// Test fetching a user with addresses
	var fetchedUser models.User
	err = db.Preload("Addresses").Where("email = ?", "customer@example.com").First(&fetchedUser).Error
	assert.NoError(t, err, "Should be able to fetch user with addresses")
	assert.NotEmpty(t, fetchedUser.Addresses, "User should have addresses")
	assert.Equal(t, "123 Main St", fetchedUser.Addresses[0].Address1)

	// Test fetching a product with category
	var fetchedProduct models.Product
	err = db.Preload("Category").First(&fetchedProduct).Error
	assert.NoError(t, err, "Should be able to fetch product with category")
	assert.NotEmpty(t, fetchedProduct.Category.Name, "Product should have a category")
	assert.Equal(t, "Smartphones", fetchedProduct.Category.Name)

	// Test fetching categories with products
	var fetchedCategory models.Category
	err = db.Preload("Products").Where("slug = ?", "smartphones").First(&fetchedCategory).Error
	assert.NoError(t, err, "Should be able to fetch category with products")
	assert.NotEmpty(t, fetchedCategory.Products, "Category should have products")
	assert.Equal(t, "iPhone 15", fetchedCategory.Products[0].Name)
}
