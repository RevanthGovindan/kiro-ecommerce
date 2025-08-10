package cart

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ecommerce-website/internal/database"
	"ecommerce-website/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Register cart routes
	api := router.Group("/api")
	RegisterRoutes(api)
	
	return router
}

func TestHandler_GetCart(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	router := setupTestRouter()
	
	t.Run("should return empty cart for new session", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/cart", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		items := data["items"].([]interface{})
		assert.Empty(t, items)
		assert.Equal(t, 0.0, data["subtotal"])
		assert.Equal(t, 0.0, data["total"])
	})
	
	t.Run("should return cart with items", func(t *testing.T) {
		product := createTestProduct(t, "Handler Test Product", 25.99, 10)
		
		// Create cart with item
		cartItems := []models.CartItem{
			{
				ProductID: product.ID,
				Quantity:  2,
				Price:     product.Price,
			},
		}
		createTestCart(t, redisClient, "test-session", cartItems)
		
		req, _ := http.NewRequest("GET", "/api/cart", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "test-session"})
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		responseItems := data["items"].([]interface{})
		assert.Len(t, responseItems, 1)
		
		item := responseItems[0].(map[string]interface{})
		assert.Equal(t, product.ID, item["productId"])
		assert.Equal(t, 2.0, item["quantity"])
		assert.Equal(t, product.Price, item["price"])
	})
}

func TestHandler_AddItem(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	router := setupTestRouter()
	
	t.Run("should add item to cart", func(t *testing.T) {
		product := createTestProduct(t, "Add Item Test", 19.99, 5)
		
		requestBody := models.AddItemRequest{
			ProductID: product.ID,
			Quantity:  3,
		}
		
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/api/cart/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		responseItems := data["items"].([]interface{})
		assert.Len(t, responseItems, 1)
		
		item := responseItems[0].(map[string]interface{})
		assert.Equal(t, product.ID, item["productId"])
		assert.Equal(t, 3.0, item["quantity"])
		assert.Equal(t, 59.97, data["subtotal"]) // 19.99 * 3
	})
	
	t.Run("should return error for invalid request", func(t *testing.T) {
		invalidRequest := map[string]interface{}{
			"productId": "",
			"quantity":  0,
		}
		
		jsonBody, _ := json.Marshal(invalidRequest)
		req, _ := http.NewRequest("POST", "/api/cart/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.False(t, response["success"].(bool))
	})
	
	t.Run("should return error for insufficient inventory", func(t *testing.T) {
		product := createTestProduct(t, "Low Stock Add Test", 10.00, 2)
		
		requestBody := models.AddItemRequest{
			ProductID: product.ID,
			Quantity:  5,
		}
		
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/api/cart/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		// Debug: print response body
		t.Logf("Response body: %s", w.Body.String())
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.False(t, response["success"].(bool))
		
		// Handle both error formats
		if errorDetail, ok := response["error"].(map[string]interface{}); ok {
			// Check both message and details fields
			message := errorDetail["message"].(string)
			details := errorDetail["details"].(string)
			assert.True(t, 
				strings.Contains(message, "insufficient inventory") || strings.Contains(details, "insufficient inventory"),
				"Expected 'insufficient inventory' in message '%s' or details '%s'", message, details)
		} else {
			// If it's a 500 error, it might be a different issue
			t.Logf("Unexpected response format: %+v", response)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		}
	})
}

func TestHandler_UpdateItem(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	router := setupTestRouter()
	
	t.Run("should update item quantity", func(t *testing.T) {
		product := createTestProduct(t, "Update Item Test", 15.00, 10)
		
		// First add an item
		cartItems := []models.CartItem{
			{
				ProductID: product.ID,
				Quantity:  2,
				Price:     product.Price,
			},
		}
		createTestCart(t, redisClient, "update-session", cartItems)
		
		// Update the item
		requestBody := models.UpdateItemRequest{
			ProductID: product.ID,
			Quantity:  5,
		}
		
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("PUT", "/api/cart/update", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "update-session"})
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		responseItems := data["items"].([]interface{})
		assert.Len(t, responseItems, 1)
		
		item := responseItems[0].(map[string]interface{})
		assert.Equal(t, 5.0, item["quantity"])
		assert.Equal(t, 75.0, data["subtotal"]) // 15.00 * 5
	})
	
	t.Run("should remove item when quantity is 0", func(t *testing.T) {
		product := createTestProduct(t, "Remove Update Test", 20.00, 5)
		
		// First add an item
		cartItems := []models.CartItem{
			{
				ProductID: product.ID,
				Quantity:  3,
				Price:     product.Price,
			},
		}
		createTestCart(t, redisClient, "remove-session", cartItems)
		
		// Update quantity to 0
		requestBody := models.UpdateItemRequest{
			ProductID: product.ID,
			Quantity:  0,
		}
		
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("PUT", "/api/cart/update", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "remove-session"})
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		responseItems := data["items"].([]interface{})
		assert.Empty(t, responseItems)
		assert.Equal(t, 0.0, data["subtotal"])
	})
	
	t.Run("should return error for item not in cart", func(t *testing.T) {
		product := createTestProduct(t, "Not In Cart Update", 10.00, 5)
		
		requestBody := models.UpdateItemRequest{
			ProductID: product.ID,
			Quantity:  2,
		}
		
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("PUT", "/api/cart/update", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.False(t, response["success"].(bool))
		if errorDetail, ok := response["error"].(map[string]interface{}); ok {
			message := errorDetail["message"].(string)
			details := errorDetail["details"].(string)
			assert.True(t, 
				strings.Contains(message, "Item not found in cart") || strings.Contains(details, "Item not found in cart"),
				"Expected 'Item not found in cart' in message '%s' or details '%s'", message, details)
		}
	})
}

func TestHandler_RemoveItem(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	router := setupTestRouter()
	
	t.Run("should remove item from cart", func(t *testing.T) {
		product1 := createTestProduct(t, "Remove Test 1", 10.00, 5)
		product2 := createTestProduct(t, "Remove Test 2", 20.00, 5)
		
		// Create cart with two items
		cartItems := []models.CartItem{
			{
				ProductID: product1.ID,
				Quantity:  2,
				Price:     product1.Price,
			},
			{
				ProductID: product2.ID,
				Quantity:  1,
				Price:     product2.Price,
			},
		}
		createTestCart(t, redisClient, "remove-item-session", cartItems)
		
		// Remove first item
		requestBody := models.RemoveItemRequest{
			ProductID: product1.ID,
		}
		
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("DELETE", "/api/cart/remove", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "remove-item-session"})
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		responseItems := data["items"].([]interface{})
		assert.Len(t, responseItems, 1)
		
		item := responseItems[0].(map[string]interface{})
		assert.Equal(t, product2.ID, item["productId"])
		assert.Equal(t, 20.0, data["subtotal"])
	})
	
	t.Run("should return error for item not in cart", func(t *testing.T) {
		product := createTestProduct(t, "Not In Cart Remove", 10.00, 5)
		
		requestBody := models.RemoveItemRequest{
			ProductID: product.ID,
		}
		
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("DELETE", "/api/cart/remove", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.False(t, response["success"].(bool))
		if errorDetail, ok := response["error"].(map[string]interface{}); ok {
			message := errorDetail["message"].(string)
			details := errorDetail["details"].(string)
			assert.True(t, 
				strings.Contains(message, "Item not found in cart") || strings.Contains(details, "Item not found in cart"),
				"Expected 'Item not found in cart' in message '%s' or details '%s'", message, details)
		}
	})
}

func TestHandler_ClearCart(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	router := setupTestRouter()
	
	t.Run("should clear cart", func(t *testing.T) {
		product := createTestProduct(t, "Clear Cart Test", 15.00, 5)
		
		// Create cart with item
		cartItems := []models.CartItem{
			{
				ProductID: product.ID,
				Quantity:  3,
				Price:     product.Price,
			},
		}
		createTestCart(t, redisClient, "clear-session", cartItems)
		
		req, _ := http.NewRequest("DELETE", "/api/cart/clear", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "clear-session"})
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Cart cleared successfully", response["message"])
	})
}