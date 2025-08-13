package search

import (
	"testing"

	"ecommerce-website/internal/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&models.Product{}, &models.Category{})

	return db
}

func TestService_FallbackSearch(t *testing.T) {
	db := setupTestDB()

	// Create test category
	category := models.Category{
		ID:       "cat-1",
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
	}
	db.Create(&category)

	// Create test product
	product := models.Product{
		ID:          "prod-1",
		Name:        "Smartphone",
		Description: "Latest smartphone with advanced features",
		Price:       599.99,
		SKU:         "PHONE-001",
		Inventory:   10,
		CategoryID:  "cat-1",
		IsActive:    true,
	}
	db.Create(&product)

	service := &Service{
		db:             db,
		elasticsearch:  nil,
		fallbackSearch: true,
	}

	// Test search functionality
	searchQuery := "smartphone"
	filters := SearchFilters{
		Search: &searchQuery,
	}
	sort := SearchSort{
		Field: "created_at",
		Order: "desc",
	}

	result, err := service.SearchProducts(filters, sort, 1, 20, false)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.Total)
	assert.Len(t, result.Products, 1)
	assert.Equal(t, "Smartphone", result.Products[0].Name)

	// Test suggestions
	suggestions, err := service.GetSuggestions("smart", 5)
	assert.NoError(t, err)
	assert.Len(t, suggestions, 1)
	assert.Equal(t, "Smartphone", suggestions[0])
}
