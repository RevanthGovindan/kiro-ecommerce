package users

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-website/internal/auth"
	"ecommerce-website/internal/config"
	"ecommerce-website/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserIntegrationTestSuite struct {
	suite.Suite
	db          *gorm.DB
	router      *gin.Engine
	authService *auth.Service
}

func (suite *UserIntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	suite.db = setupTestDB(suite.T())
	suite.authService = auth.NewService(suite.db, cfg)

	// Setup router with routes
	suite.router = gin.New()
	userService := NewService(suite.db)
	userHandler := NewHandler(userService)
	SetupRoutes(suite.router, userHandler, suite.authService)
}

func (suite *UserIntegrationTestSuite) TearDownSuite() {
	// SQLite in-memory database will be automatically cleaned up
}

func (suite *UserIntegrationTestSuite) SetupTest() {
	// Clean up before each test
	suite.db.Exec("DELETE FROM addresses")
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserIntegrationTestSuite) createTestUserAndToken() (*models.User, string) {
	user := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     stringPtr("1234567890"),
		Role:      "customer",
		IsActive:  true,
	}
	suite.db.Create(user)

	// Generate token
	tokens, err := suite.authService.GenerateTokens(user)
	suite.Require().NoError(err)

	return user, tokens.AccessToken
}

func (suite *UserIntegrationTestSuite) TestGetProfileIntegration() {
	user, token := suite.createTestUserAndToken()

	req, _ := http.NewRequest("GET", "/api/users/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), user.Email, data["email"])
	assert.Equal(suite.T(), user.FirstName, data["firstName"])
	assert.Equal(suite.T(), user.LastName, data["lastName"])
}

func (suite *UserIntegrationTestSuite) TestGetProfileUnauthorized() {
	req, _ := http.NewRequest("GET", "/api/users/profile", nil)
	// No authorization header
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *UserIntegrationTestSuite) TestUpdateProfileIntegration() {
	_, token := suite.createTestUserAndToken()

	updateReq := UpdateProfileRequest{
		FirstName: "Jane",
		LastName:  "Smith",
		Phone:     stringPtr("9876543210"),
	}

	jsonData, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/users/profile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), updateReq.FirstName, data["firstName"])
	assert.Equal(suite.T(), updateReq.LastName, data["lastName"])
}

func (suite *UserIntegrationTestSuite) TestGetOrdersIntegration() {
	user, token := suite.createTestUserAndToken()

	// Create test orders
	order1 := &models.Order{
		UserID:   user.ID,
		Status:   "completed",
		Subtotal: 100.00,
		Total:    110.00,
	}
	order2 := &models.Order{
		UserID:   user.ID,
		Status:   "pending",
		Subtotal: 50.00,
		Total:    55.00,
	}
	suite.db.Create(order1)
	suite.db.Create(order2)

	req, _ := http.NewRequest("GET", "/api/users/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)
}

func (suite *UserIntegrationTestSuite) TestCreateAddressIntegration() {
	_, token := suite.createTestUserAndToken()

	addressReq := CreateAddressRequest{
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

	jsonData, _ := json.Marshal(addressReq)
	req, _ := http.NewRequest("POST", "/api/users/addresses", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), addressReq.Type, data["type"])
	assert.Equal(suite.T(), addressReq.FirstName, data["firstName"])
	assert.Equal(suite.T(), addressReq.Address1, data["address1"])
}

func (suite *UserIntegrationTestSuite) TestGetAddressesIntegration() {
	user, token := suite.createTestUserAndToken()

	// Create test addresses
	address1 := &models.Address{
		UserID:     user.ID,
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
	address2 := &models.Address{
		UserID:     user.ID,
		Type:       "billing",
		FirstName:  "John",
		LastName:   "Doe",
		Address1:   "456 Oak Ave",
		City:       "Los Angeles",
		State:      "CA",
		PostalCode: "90210",
		Country:    "US",
		IsDefault:  false,
	}
	suite.db.Create(address1)
	suite.db.Create(address2)

	req, _ := http.NewRequest("GET", "/api/users/addresses", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)
}

func (suite *UserIntegrationTestSuite) TestCompleteAddressWorkflow() {
	_, token := suite.createTestUserAndToken()

	// 1. Create address
	addressReq := CreateAddressRequest{
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

	jsonData, _ := json.Marshal(addressReq)
	req, _ := http.NewRequest("POST", "/api/users/addresses", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResponse)
	assert.NoError(suite.T(), err)

	addressData := createResponse["data"].(map[string]interface{})
	addressID := addressData["id"].(string)

	// 2. Get specific address
	req, _ = http.NewRequest("GET", "/api/users/addresses/"+addressID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// 3. Update address
	updateReq := UpdateAddressRequest{
		FirstName:  "Jane",
		LastName:   "Smith",
		Address1:   "456 Oak Ave",
		City:       "Los Angeles",
		State:      "CA",
		PostalCode: "90210",
		Country:    "US",
		IsDefault:  true,
	}

	jsonData, _ = json.Marshal(updateReq)
	req, _ = http.NewRequest("PUT", "/api/users/addresses/"+addressID, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var updateResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &updateResponse)
	assert.NoError(suite.T(), err)

	updatedData := updateResponse["data"].(map[string]interface{})
	assert.Equal(suite.T(), updateReq.FirstName, updatedData["firstName"])
	assert.Equal(suite.T(), updateReq.LastName, updatedData["lastName"])

	// 4. Delete address
	req, _ = http.NewRequest("DELETE", "/api/users/addresses/"+addressID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// 5. Verify address is deleted
	req, _ = http.NewRequest("GET", "/api/users/addresses/"+addressID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func TestUserIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserIntegrationTestSuite))
}
