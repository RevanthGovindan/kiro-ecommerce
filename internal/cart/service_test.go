package cart

import (
	"context"
	"testing"

	"ecommerce-website/internal/config"
	"ecommerce-website/internal/database"
	"ecommerce-website/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Setup test database
	cfg := &config.Config{
		DatabaseURL: ":memory:",
	}
	
	if err := database.InitializeTest(cfg); err != nil {
		panic("Failed to initialize test database: " + err.Error())
	}
	
	m.Run()
}

func TestService_GetCart(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	service := NewService()
	ctx := context.Background()
	sessionID := "test-session-1"
	
	t.Run("should return empty cart when cart doesn't exist", func(t *testing.T) {
		cart, err := service.GetCart(ctx, sessionID)
		
		require.NoError(t, err)
		assert.Equal(t, sessionID, cart.SessionID)
		assert.Empty(t, cart.Items)
		assert.Equal(t, 0.0, cart.Subtotal)
		assert.Equal(t, 0.0, cart.Total)
	})
	
	t.Run("should return existing cart", func(t *testing.T) {
		// Create test product
		product := createTestProduct(t, "Test Product", 10.99, 5)
		
		// Create test cart
		items := []models.CartItem{
			{
				ProductID: product.ID,
				Quantity:  2,
				Price:     product.Price,
			},
		}
		expectedCart := createTestCart(t, redisClient, sessionID, items)
		
		cart, err := service.GetCart(ctx, sessionID)
		
		require.NoError(t, err)
		assert.Equal(t, expectedCart.SessionID, cart.SessionID)
		assert.Len(t, cart.Items, 1)
		assert.Equal(t, expectedCart.Items[0].ProductID, cart.Items[0].ProductID)
		assert.Equal(t, expectedCart.Items[0].Quantity, cart.Items[0].Quantity)
		assert.Equal(t, expectedCart.Subtotal, cart.Subtotal)
	})
}

func TestService_AddItem(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	service := NewService()
	ctx := context.Background()
	sessionID := "test-session-2"
	
	t.Run("should add new item to empty cart", func(t *testing.T) {
		product := createTestProduct(t, "Test Product 1", 15.99, 10)
		
		cart, err := service.AddItem(ctx, sessionID, product.ID, 3)
		
		require.NoError(t, err)
		assert.Equal(t, sessionID, cart.SessionID)
		assert.Len(t, cart.Items, 1)
		assert.Equal(t, product.ID, cart.Items[0].ProductID)
		assert.Equal(t, 3, cart.Items[0].Quantity)
		assert.Equal(t, product.Price, cart.Items[0].Price)
		assert.Equal(t, 47.97, cart.Items[0].Total) // 15.99 * 3
		assert.Equal(t, 47.97, cart.Subtotal)
		assert.Equal(t, 47.97, cart.Total)
	})
	
	t.Run("should add quantity to existing item", func(t *testing.T) {
		product := createTestProduct(t, "Test Product 2", 20.00, 10)
		
		// Add item first time
		cart, err := service.AddItem(ctx, sessionID, product.ID, 2)
		require.NoError(t, err)
		
		// Add same item again
		cart, err = service.AddItem(ctx, sessionID, product.ID, 1)
		require.NoError(t, err)
		
		// Should have combined quantities
		item := cart.FindItem(product.ID)
		require.NotNil(t, item)
		assert.Equal(t, 3, item.Quantity)
		assert.Equal(t, 60.0, item.Total) // 20.00 * 3
	})
	
	t.Run("should fail when product is not active", func(t *testing.T) {
		product := createTestProduct(t, "Inactive Product", 10.00, 5)
		product.IsActive = false
		database.GetDB().Save(product)
		
		_, err := service.AddItem(ctx, sessionID, product.ID, 1)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "product is not available")
	})
	
	t.Run("should fail when insufficient inventory", func(t *testing.T) {
		product := createTestProduct(t, "Low Stock Product", 10.00, 2)
		
		_, err := service.AddItem(ctx, sessionID, product.ID, 5)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient inventory")
	})
	
	t.Run("should fail when adding to existing item exceeds inventory", func(t *testing.T) {
		product := createTestProduct(t, "Limited Product", 10.00, 5)
		
		// Add 3 items first
		_, err := service.AddItem(ctx, sessionID, product.ID, 3)
		require.NoError(t, err)
		
		// Try to add 3 more (total would be 6, but only 5 available)
		_, err = service.AddItem(ctx, sessionID, product.ID, 3)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient inventory")
	})
}

func TestService_UpdateItem(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	service := NewService()
	ctx := context.Background()
	sessionID := "test-session-3"
	
	t.Run("should update item quantity", func(t *testing.T) {
		product := createTestProduct(t, "Update Test Product", 25.00, 10)
		
		// Add item first
		cart, err := service.AddItem(ctx, sessionID, product.ID, 2)
		require.NoError(t, err)
		
		// Update quantity
		cart, err = service.UpdateItem(ctx, sessionID, product.ID, 4)
		require.NoError(t, err)
		
		item := cart.FindItem(product.ID)
		require.NotNil(t, item)
		assert.Equal(t, 4, item.Quantity)
		assert.Equal(t, 100.0, item.Total) // 25.00 * 4
		assert.Equal(t, 100.0, cart.Subtotal)
	})
	
	t.Run("should remove item when quantity is 0", func(t *testing.T) {
		uniqueSessionID := "test-session-update-remove"
		product := createTestProduct(t, "Remove Test Product", 15.00, 5)
		
		// Add item first
		cart, err := service.AddItem(ctx, uniqueSessionID, product.ID, 2)
		require.NoError(t, err)
		
		// Update quantity to 0
		cart, err = service.UpdateItem(ctx, uniqueSessionID, product.ID, 0)
		require.NoError(t, err)
		
		item := cart.FindItem(product.ID)
		assert.Nil(t, item)
		assert.Empty(t, cart.Items)
	})
	
	t.Run("should fail when item not in cart", func(t *testing.T) {
		uniqueSessionID := "test-session-update-not-found"
		product := createTestProduct(t, "Not In Cart Product", 10.00, 5)
		
		_, err := service.UpdateItem(ctx, uniqueSessionID, product.ID, 2)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item not found in cart")
	})
	
	t.Run("should fail when insufficient inventory", func(t *testing.T) {
		uniqueSessionID := "test-session-update-inventory"
		product := createTestProduct(t, "Limited Update Product", 10.00, 3)
		
		// Add item first
		_, err := service.AddItem(ctx, uniqueSessionID, product.ID, 2)
		require.NoError(t, err)
		
		// Try to update to quantity that exceeds inventory
		_, err = service.UpdateItem(ctx, uniqueSessionID, product.ID, 5)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient inventory")
	})
}

func TestService_RemoveItem(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	service := NewService()
	ctx := context.Background()
	sessionID := "test-session-4"
	
	t.Run("should remove item from cart", func(t *testing.T) {
		product1 := createTestProduct(t, "Remove Product 1", 10.00, 5)
		product2 := createTestProduct(t, "Remove Product 2", 20.00, 5)
		
		// Add both items
		_, err := service.AddItem(ctx, sessionID, product1.ID, 2)
		require.NoError(t, err)
		_, err = service.AddItem(ctx, sessionID, product2.ID, 1)
		require.NoError(t, err)
		
		// Remove first item
		cart, err := service.RemoveItem(ctx, sessionID, product1.ID)
		require.NoError(t, err)
		
		assert.Len(t, cart.Items, 1)
		assert.Equal(t, product2.ID, cart.Items[0].ProductID)
		assert.Equal(t, 20.0, cart.Subtotal)
	})
	
	t.Run("should fail when item not in cart", func(t *testing.T) {
		product := createTestProduct(t, "Not In Cart Remove", 10.00, 5)
		
		_, err := service.RemoveItem(ctx, sessionID, product.ID)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item not found in cart")
	})
}

func TestService_ClearCart(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	service := NewService()
	ctx := context.Background()
	sessionID := "test-session-5"
	
	t.Run("should clear cart", func(t *testing.T) {
		product := createTestProduct(t, "Clear Test Product", 10.00, 5)
		
		// Add item first
		_, err := service.AddItem(ctx, sessionID, product.ID, 2)
		require.NoError(t, err)
		
		// Clear cart
		err = service.ClearCart(ctx, sessionID)
		require.NoError(t, err)
		
		// Verify cart is empty
		cart, err := service.GetCart(ctx, sessionID)
		require.NoError(t, err)
		assert.Empty(t, cart.Items)
		assert.Equal(t, 0.0, cart.Subtotal)
	})
}

func TestService_GetCartWithProducts(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	service := NewService()
	ctx := context.Background()
	sessionID := "test-session-6"
	
	t.Run("should populate product details", func(t *testing.T) {
		product := createTestProduct(t, "Product Details Test", 30.00, 5)
		
		// Add item
		_, err := service.AddItem(ctx, sessionID, product.ID, 1)
		require.NoError(t, err)
		
		// Get cart with products
		cart, err := service.GetCartWithProducts(ctx, sessionID)
		require.NoError(t, err)
		
		assert.Len(t, cart.Items, 1)
		assert.Equal(t, product.ID, cart.Items[0].Product.ID)
		assert.Equal(t, product.Name, cart.Items[0].Product.Name)
		assert.Equal(t, product.Price, cart.Items[0].Product.Price)
	})
}