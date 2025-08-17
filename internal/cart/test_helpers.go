package cart

import (
	"context"
	"encoding/json"
	"testing"

	"ecommerce-website/internal/database"
	"ecommerce-website/internal/models"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// setupTestRedis sets up a test Redis client
func setupTestRedis(t *testing.T) *redis.Client {
	// Use Redis database 1 for testing to avoid conflicts
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})

	// Test connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis is not available, skipping Redis-dependent tests")
	}

	// Clear test database
	err = client.FlushDB(ctx).Err()
	require.NoError(t, err, "Failed to flush test Redis database")

	return client
}

// createTestProduct creates a test product in the database
func createTestProduct(t *testing.T, name string, price float64, inventory int) *models.Product {
	// Create a test category first (or get existing one)
	category := &models.Category{
		Name:     "Test Category",
		Slug:     "test-category",
		IsActive: true,
	}
	// Try to find existing category first
	err := database.GetDB().Where("slug = ?", "test-category").First(category).Error
	if err != nil {
		// Category doesn't exist, create it
		err = database.GetDB().Create(category).Error
		require.NoError(t, err, "Failed to create test category")
	}

	product := &models.Product{
		Name:       name,
		Price:      price,
		SKU:        "TEST-" + name,
		Inventory:  inventory,
		IsActive:   true,
		CategoryID: category.ID,
	}

	err = database.GetDB().Create(product).Error
	require.NoError(t, err, "Failed to create test product")

	return product
}

// createTestCart creates a test cart in Redis
func createTestCart(t *testing.T, client *redis.Client, sessionID string, items []models.CartItem) *models.Cart {
	cart := &models.Cart{
		SessionID: sessionID,
		Items:     items,
	}
	cart.CalculateTotals()

	data, err := json.Marshal(cart)
	require.NoError(t, err, "Failed to marshal test cart")

	ctx := context.Background()
	err = client.Set(ctx, cartKeyPrefix+sessionID, data, cartTTL).Err()
	require.NoError(t, err, "Failed to save test cart to Redis")

	return cart
}

// cleanupTestData cleans up test data from database and Redis
func cleanupTestData(t *testing.T, client *redis.Client) {
	ctx := context.Background()

	// Clear Redis
	err := client.FlushDB(ctx).Err()
	require.NoError(t, err, "Failed to flush test Redis database")

	// Clear database tables
	db := database.GetDB()
	db.Exec("DELETE FROM order_items")
	db.Exec("DELETE FROM orders")
	db.Exec("DELETE FROM products")
	db.Exec("DELETE FROM categories")
	db.Exec("DELETE FROM addresses")
	db.Exec("DELETE FROM users")
}
