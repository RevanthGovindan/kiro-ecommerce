package users

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-website/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserHandlerTestSuite struct {
	suite.Suite
	db      *gorm.DB
	handler *Handler
	router  *gin.Engine
}

func (suite *UserHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	suite.db = setupTestDB(suite.T())
	service := NewService(suite.db)
	suite.handler = NewHandler(service)

	suite.router = gin.New()
}

func (suite *UserHandlerTestSuite) TearDownSuite() {
	// SQLite in-memory database will be automatically cleaned up
}

func (suite *UserHandlerTestSuite) SetupTest() {
	// Clean up before each test
	suite.db.Exec("DELETE FROM addresses")
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserHandlerTestSuite) createTestUser() *models.User {
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
	return user
}

func (suite *UserHandlerTestSuite) TestGetProfile() {
	user := suite.createTestUser()

	// Test successful profile retrieval
	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", user.ID)

	suite.handler.GetProfile(c)

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

func (suite *UserHandlerTestSuite) TestGetProfileUnauthorized() {
	req, _ := http.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set userID to simulate unauthorized access

	suite.handler.GetProfile(c)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
}

func (suite *UserHandlerTestSuite) TestUpdateProfile() {
	user := suite.createTestUser()

	updateReq := UpdateProfileRequest{
		FirstName: "Jane",
		LastName:  "Smith",
		Phone:     stringPtr("9876543210"),
	}

	jsonData, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", user.ID)

	suite.handler.UpdateProfile(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), updateReq.FirstName, data["firstName"])
	assert.Equal(suite.T(), updateReq.LastName, data["lastName"])
}

func (suite *UserHandlerTestSuite) TestUpdateProfileValidationError() {
	user := suite.createTestUser()

	// Invalid request (missing required fields)
	invalidReq := map[string]interface{}{
		"firstName": "", // Empty required field
	}

	jsonData, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", user.ID)

	suite.handler.UpdateProfile(c)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
}

func (suite *UserHandlerTestSuite) TestGetOrders() {
	user := suite.createTestUser()

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

	req, _ := http.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", user.ID)

	suite.handler.GetOrders(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)
}

func (suite *UserHandlerTestSuite) TestCreateAddress() {
	user := suite.createTestUser()

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
	req, _ := http.NewRequest("POST", "/addresses", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", user.ID)

	suite.handler.CreateAddress(c)

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

func (suite *UserHandlerTestSuite) TestCreateAddressValidationError() {
	user := suite.createTestUser()

	// Invalid request (missing required fields)
	invalidReq := map[string]interface{}{
		"type": "invalid_type", // Invalid type
	}

	jsonData, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest("POST", "/addresses", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", user.ID)

	suite.handler.CreateAddress(c)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
}

func (suite *UserHandlerTestSuite) TestGetAddresses() {
	user := suite.createTestUser()

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

	req, _ := http.NewRequest("GET", "/addresses", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", user.ID)

	suite.handler.GetAddresses(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)
}

func (suite *UserHandlerTestSuite) TestGetAddress() {
	user := suite.createTestUser()
	address := &models.Address{
		UserID:     user.ID,
		Type:       "shipping",
		FirstName:  "John",
		LastName:   "Doe",
		Address1:   "123 Main St",
		City:       "New York",
		State:      "NY",
		PostalCode: "10001",
		Country:    "US",
	}
	suite.db.Create(address)

	req, _ := http.NewRequest("GET", "/addresses/"+address.ID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: address.ID}}
	c.Set("user_id", user.ID)

	suite.handler.GetAddress(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), address.ID, data["id"])
	assert.Equal(suite.T(), address.FirstName, data["firstName"])
}

func (suite *UserHandlerTestSuite) TestUpdateAddress() {
	user := suite.createTestUser()
	address := &models.Address{
		UserID:     user.ID,
		Type:       "shipping",
		FirstName:  "John",
		LastName:   "Doe",
		Address1:   "123 Main St",
		City:       "New York",
		State:      "NY",
		PostalCode: "10001",
		Country:    "US",
		IsDefault:  false,
	}
	suite.db.Create(address)

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

	jsonData, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/addresses/"+address.ID, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: address.ID}}
	c.Set("user_id", user.ID)

	suite.handler.UpdateAddress(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), updateReq.FirstName, data["firstName"])
	assert.Equal(suite.T(), updateReq.LastName, data["lastName"])
}

func (suite *UserHandlerTestSuite) TestDeleteAddress() {
	user := suite.createTestUser()
	address := &models.Address{
		UserID:     user.ID,
		Type:       "shipping",
		FirstName:  "John",
		LastName:   "Doe",
		Address1:   "123 Main St",
		City:       "New York",
		State:      "NY",
		PostalCode: "10001",
		Country:    "US",
	}
	suite.db.Create(address)

	req, _ := http.NewRequest("DELETE", "/addresses/"+address.ID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: address.ID}}
	c.Set("user_id", user.ID)

	suite.handler.DeleteAddress(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	// Verify address is deleted
	var deletedAddress models.Address
	err = suite.db.First(&deletedAddress, "id = ?", address.ID).Error
	assert.Error(suite.T(), err) // Should not find the address
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}
