package database

import (
	"testing"

	"ecommerce-website/internal/config"
	"ecommerce-website/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseConnection(t *testing.T) {
	// Load test configuration
	cfg := &config.Config{
		DatabaseURL: "postgres://user:password@localhost:5433/ecommerce?sslmode=disable",
	}

	// Initialize database
	err := Initialize(cfg)
	assert.NoError(t, err, "Database initialization should not fail")
	assert.NotNil(t, DB, "Database connection should be established")

	// Test basic query
	var count int64
	err = DB.Model(&models.User{}).Count(&count).Error
	assert.NoError(t, err, "Should be able to query users table")

	// Clean up
	Close()
}

func TestModelsRelationships(t *testing.T) {
	// Load test configuration
	cfg := &config.Config{
		DatabaseURL: "postgres://user:password@localhost:5433/ecommerce?sslmode=disable",
	}

	// Initialize database
	err := Initialize(cfg)
	assert.NoError(t, err)
	defer Close()

	// Test fetching a user with addresses
	var user models.User
	err = DB.Preload("Addresses").Where("email = ?", "customer@example.com").First(&user).Error
	assert.NoError(t, err, "Should be able to fetch user with addresses")
	assert.NotEmpty(t, user.Addresses, "User should have addresses")

	// Test fetching a product with category
	var product models.Product
	err = DB.Preload("Category").First(&product).Error
	assert.NoError(t, err, "Should be able to fetch product with category")
	assert.NotEmpty(t, product.Category.Name, "Product should have a category")

	// Test fetching categories with products
	var category models.Category
	err = DB.Preload("Products").Where("slug = ?", "smartphones").First(&category).Error
	assert.NoError(t, err, "Should be able to fetch category with products")
	assert.NotEmpty(t, category.Products, "Category should have products")
}