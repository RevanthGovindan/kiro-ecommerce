package cart

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-website/internal/database"
	"ecommerce-website/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCartIntegration(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	router := setupTestRouter()
	
	// Create test products
	product1 := createTestProduct(t, "Integration Product 1", 29.99, 10)
	product2 := createTestProduct(t, "Integration Product 2", 49.99, 5)
	product3 := createTestProduct(t, "Integration Product 3", 19.99, 3)
	
	t.Run("complete cart workflow", func(t *testing.T) {
		// Step 1: Get empty cart
		req, _ := http.NewRequest("GET", "/api/cart", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data := response["data"].(map[string]interface{})
		assert.Empty(t, data["items"])
		assert.Equal(t, 0.0, data["subtotal"])
		
		// Extract session ID from cookie
		cookies := w.Result().Cookies()
		var sessionID string
		for _, cookie := range cookies {
			if cookie.Name == "session_id" {
				sessionID = cookie.Value
				break
			}
		}
		require.NotEmpty(t, sessionID, "Session ID should be set")
		
		// Step 2: Add first product
		addReq1 := models.AddItemRequest{
			ProductID: product1.ID,
			Quantity:  2,
		}
		jsonBody, _ := json.Marshal(addReq1)
		req, _ = http.NewRequest("POST", "/api/cart/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data = response["data"].(map[string]interface{})
		items := data["items"].([]interface{})
		assert.Len(t, items, 1)
		assert.Equal(t, 59.98, data["subtotal"]) // 29.99 * 2
		
		// Step 3: Add second product
		addReq2 := models.AddItemRequest{
			ProductID: product2.ID,
			Quantity:  1,
		}
		jsonBody, _ = json.Marshal(addReq2)
		req, _ = http.NewRequest("POST", "/api/cart/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data = response["data"].(map[string]interface{})
		items = data["items"].([]interface{})
		assert.Len(t, items, 2)
		assert.Equal(t, 109.97, data["subtotal"]) // 59.98 + 49.99
		
		// Step 4: Add more of first product (should combine)
		addReq3 := models.AddItemRequest{
			ProductID: product1.ID,
			Quantity:  1,
		}
		jsonBody, _ = json.Marshal(addReq3)
		req, _ = http.NewRequest("POST", "/api/cart/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data = response["data"].(map[string]interface{})
		items = data["items"].([]interface{})
		assert.Len(t, items, 2) // Still 2 items, but first one should have quantity 3
		assert.Equal(t, 139.96, data["subtotal"]) // (29.99 * 3) + 49.99
		
		// Verify first item has quantity 3
		var product1Item map[string]interface{}
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			if itemMap["productId"] == product1.ID {
				product1Item = itemMap
				break
			}
		}
		require.NotNil(t, product1Item)
		assert.Equal(t, 3.0, product1Item["quantity"])
		
		// Step 5: Update quantity of second product
		updateReq := models.UpdateItemRequest{
			ProductID: product2.ID,
			Quantity:  3,
		}
		jsonBody, _ = json.Marshal(updateReq)
		req, _ = http.NewRequest("PUT", "/api/cart/update", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data = response["data"].(map[string]interface{})
		assert.Equal(t, 239.94, data["subtotal"]) // (29.99 * 3) + (49.99 * 3)
		
		// Step 6: Add third product
		addReq4 := models.AddItemRequest{
			ProductID: product3.ID,
			Quantity:  2,
		}
		jsonBody, _ = json.Marshal(addReq4)
		req, _ = http.NewRequest("POST", "/api/cart/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data = response["data"].(map[string]interface{})
		items = data["items"].([]interface{})
		assert.Len(t, items, 3)
		assert.Equal(t, 279.92, data["subtotal"]) // 239.94 + (19.99 * 2)
		
		// Step 7: Remove second product
		removeReq := models.RemoveItemRequest{
			ProductID: product2.ID,
		}
		jsonBody, _ = json.Marshal(removeReq)
		req, _ = http.NewRequest("DELETE", "/api/cart/remove", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data = response["data"].(map[string]interface{})
		items = data["items"].([]interface{})
		assert.Len(t, items, 2)
		assert.Equal(t, 129.95, data["subtotal"]) // (29.99 * 3) + (19.99 * 2)
		
		// Step 8: Update third product quantity to 0 (should remove it)
		updateReq2 := models.UpdateItemRequest{
			ProductID: product3.ID,
			Quantity:  0,
		}
		jsonBody, _ = json.Marshal(updateReq2)
		req, _ = http.NewRequest("PUT", "/api/cart/update", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data = response["data"].(map[string]interface{})
		items = data["items"].([]interface{})
		assert.Len(t, items, 1)
		assert.Equal(t, 89.97, data["subtotal"]) // 29.99 * 3
		
		// Step 9: Clear cart
		req, _ = http.NewRequest("DELETE", "/api/cart/clear", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Step 10: Verify cart is empty
		req, _ = http.NewRequest("GET", "/api/cart", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data = response["data"].(map[string]interface{})
		items = data["items"].([]interface{})
		assert.Empty(t, items)
		assert.Equal(t, 0.0, data["subtotal"])
	})
}

func TestCartInventoryConstraints(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	router := setupTestRouter()
	
	// Create product with limited inventory
	product := createTestProduct(t, "Limited Stock Product", 25.00, 3)
	
	t.Run("inventory constraints workflow", func(t *testing.T) {
		// Step 1: Add maximum available quantity
		addReq := models.AddItemRequest{
			ProductID: product.ID,
			Quantity:  3,
		}
		jsonBody, _ := json.Marshal(addReq)
		req, _ := http.NewRequest("POST", "/api/cart/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Extract session ID
		cookies := w.Result().Cookies()
		var sessionID string
		for _, cookie := range cookies {
			if cookie.Name == "session_id" {
				sessionID = cookie.Value
				break
			}
		}
		
		// Step 2: Try to add more (should fail)
		addReq2 := models.AddItemRequest{
			ProductID: product.ID,
			Quantity:  1,
		}
		jsonBody, _ = json.Marshal(addReq2)
		req, _ = http.NewRequest("POST", "/api/cart/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.False(t, response["success"].(bool))
		if errorDetail, ok := response["error"].(map[string]interface{}); ok {
			assert.Contains(t, errorDetail["message"].(string), "insufficient inventory")
		} else if message, ok := response["message"].(string); ok {
			assert.Contains(t, message, "insufficient inventory")
		}
		
		// Step 3: Try to update to exceed inventory (should fail)
		updateReq := models.UpdateItemRequest{
			ProductID: product.ID,
			Quantity:  5,
		}
		jsonBody, _ = json.Marshal(updateReq)
		req, _ = http.NewRequest("PUT", "/api/cart/update", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.False(t, response["success"].(bool))
		if errorDetail, ok := response["error"].(map[string]interface{}); ok {
			assert.Contains(t, errorDetail["message"].(string), "insufficient inventory")
		} else if message, ok := response["message"].(string); ok {
			assert.Contains(t, message, "insufficient inventory")
		}
		
		// Step 4: Update to valid quantity (should succeed)
		updateReq2 := models.UpdateItemRequest{
			ProductID: product.ID,
			Quantity:  2,
		}
		jsonBody, _ = json.Marshal(updateReq2)
		req, _ = http.NewRequest("PUT", "/api/cart/update", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, 50.0, data["subtotal"]) // 25.00 * 2
	})
}

func TestCartPersistence(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer cleanupTestData(t, redisClient)
	
	// Override the global Redis client for testing
	originalClient := database.RedisClient
	database.RedisClient = redisClient
	defer func() { database.RedisClient = originalClient }()
	
	router := setupTestRouter()
	product := createTestProduct(t, "Persistence Test Product", 15.99, 10)
	
	t.Run("cart should persist across requests", func(t *testing.T) {
		sessionID := "persistence-test-session"
		
		// Add item to cart
		addReq := models.AddItemRequest{
			ProductID: product.ID,
			Quantity:  2,
		}
		jsonBody, _ := json.Marshal(addReq)
		req, _ := http.NewRequest("POST", "/api/cart/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify cart persists by getting it in a new request
		req, _ = http.NewRequest("GET", "/api/cart", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data := response["data"].(map[string]interface{})
		items := data["items"].([]interface{})
		assert.Len(t, items, 1)
		
		item := items[0].(map[string]interface{})
		assert.Equal(t, product.ID, item["productId"])
		assert.Equal(t, 2.0, item["quantity"])
		assert.Equal(t, 31.98, data["subtotal"]) // 15.99 * 2
	})
	
	t.Run("cart should be isolated by session", func(t *testing.T) {
		session1 := "isolation-test-session-1"
		session2 := "isolation-test-session-2"
		
		// Add item to first session
		addReq := models.AddItemRequest{
			ProductID: product.ID,
			Quantity:  1,
		}
		jsonBody, _ := json.Marshal(addReq)
		req, _ := http.NewRequest("POST", "/api/cart/add", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: session1})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Check second session has empty cart
		req, _ = http.NewRequest("GET", "/api/cart", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: session2})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data := response["data"].(map[string]interface{})
		items := data["items"].([]interface{})
		assert.Empty(t, items)
		assert.Equal(t, 0.0, data["subtotal"])
		
		// Verify first session still has item
		req, _ = http.NewRequest("GET", "/api/cart", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: session1})
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data = response["data"].(map[string]interface{})
		items = data["items"].([]interface{})
		assert.Len(t, items, 1)
	})
}