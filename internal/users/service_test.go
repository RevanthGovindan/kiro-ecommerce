package users

import (
	"testing"

	"ecommerce-website/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserServiceTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *Service
}

func (suite *UserServiceTestSuite) SetupSuite() {
	suite.db = setupTestDB(suite.T())
	suite.service = NewService(suite.db)
}

func (suite *UserServiceTestSuite) TearDownSuite() {
	// SQLite in-memory database will be automatically cleaned up
}

func (suite *UserServiceTestSuite) SetupTest() {
	// Clean up before each test
	suite.db.Exec("DELETE FROM addresses")
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserServiceTestSuite) createTestUser() *models.User {
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

func (suite *UserServiceTestSuite) TestGetProfile() {
	user := suite.createTestUser()

	// Test successful profile retrieval
	profile, err := suite.service.GetProfile(user.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), profile)
	assert.Equal(suite.T(), user.Email, profile.Email)
	assert.Equal(suite.T(), user.FirstName, profile.FirstName)
	assert.Equal(suite.T(), user.LastName, profile.LastName)
	assert.Empty(suite.T(), profile.Password) // Password should be removed

	// Test user not found
	_, err = suite.service.GetProfile("nonexistent-id")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrUserNotFound, err)
}

func (suite *UserServiceTestSuite) TestUpdateProfile() {
	user := suite.createTestUser()

	req := UpdateProfileRequest{
		FirstName: "Jane",
		LastName:  "Smith",
		Phone:     stringPtr("9876543210"),
	}

	// Test successful profile update
	updatedUser, err := suite.service.UpdateProfile(user.ID, req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedUser)
	assert.Equal(suite.T(), req.FirstName, updatedUser.FirstName)
	assert.Equal(suite.T(), req.LastName, updatedUser.LastName)
	assert.Equal(suite.T(), req.Phone, updatedUser.Phone)
	assert.Empty(suite.T(), updatedUser.Password) // Password should be removed

	// Test user not found
	_, err = suite.service.UpdateProfile("nonexistent-id", req)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrUserNotFound, err)
}

func (suite *UserServiceTestSuite) TestGetUserOrders() {
	user := suite.createTestUser()

	// Test with no orders
	orders, err := suite.service.GetUserOrders(user.ID)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), orders)

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

	// Test with orders
	orders, err = suite.service.GetUserOrders(user.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), orders, 2)
}

func (suite *UserServiceTestSuite) TestCreateAddress() {
	user := suite.createTestUser()

	req := CreateAddressRequest{
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

	// Test successful address creation
	address, err := suite.service.CreateAddress(user.ID, req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), address)
	assert.Equal(suite.T(), user.ID, address.UserID)
	assert.Equal(suite.T(), req.Type, address.Type)
	assert.Equal(suite.T(), req.FirstName, address.FirstName)
	assert.Equal(suite.T(), req.LastName, address.LastName)
	assert.Equal(suite.T(), req.Address1, address.Address1)
	assert.Equal(suite.T(), req.IsDefault, address.IsDefault)

	// Test user not found
	_, err = suite.service.CreateAddress("nonexistent-id", req)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrUserNotFound, err)
}

func (suite *UserServiceTestSuite) TestCreateAddressDefaultHandling() {
	user := suite.createTestUser()

	// Create first default shipping address
	req1 := CreateAddressRequest{
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
	address1, err := suite.service.CreateAddress(user.ID, req1)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), address1.IsDefault)

	// Create second default shipping address
	req2 := CreateAddressRequest{
		Type:       "shipping",
		FirstName:  "Jane",
		LastName:   "Smith",
		Address1:   "456 Oak Ave",
		City:       "Los Angeles",
		State:      "CA",
		PostalCode: "90210",
		Country:    "US",
		IsDefault:  true,
	}
	address2, err := suite.service.CreateAddress(user.ID, req2)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), address2.IsDefault)

	// Check that first address is no longer default
	var updatedAddress1 models.Address
	suite.db.First(&updatedAddress1, "id = ?", address1.ID)
	assert.False(suite.T(), updatedAddress1.IsDefault)
}

func (suite *UserServiceTestSuite) TestGetAddresses() {
	user := suite.createTestUser()

	// Test with no addresses
	addresses, err := suite.service.GetAddresses(user.ID)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), addresses)

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

	// Test with addresses
	addresses, err = suite.service.GetAddresses(user.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), addresses, 2)
	// Default address should be first
	assert.True(suite.T(), addresses[0].IsDefault)
}

func (suite *UserServiceTestSuite) TestGetAddress() {
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

	// Test successful address retrieval
	retrievedAddress, err := suite.service.GetAddress(user.ID, address.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedAddress)
	assert.Equal(suite.T(), address.ID, retrievedAddress.ID)
	assert.Equal(suite.T(), address.UserID, retrievedAddress.UserID)

	// Test address not found
	_, err = suite.service.GetAddress(user.ID, "nonexistent-id")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrAddressNotFound, err)

	// Test unauthorized access (different user)
	otherUser := suite.createTestUser()
	otherUser.Email = "other@example.com"
	suite.db.Save(otherUser)

	_, err = suite.service.GetAddress(otherUser.ID, address.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrAddressNotFound, err)
}

func (suite *UserServiceTestSuite) TestUpdateAddress() {
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

	req := UpdateAddressRequest{
		FirstName:  "Jane",
		LastName:   "Smith",
		Address1:   "456 Oak Ave",
		City:       "Los Angeles",
		State:      "CA",
		PostalCode: "90210",
		Country:    "US",
		IsDefault:  true,
	}

	// Test successful address update
	updatedAddress, err := suite.service.UpdateAddress(user.ID, address.ID, req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedAddress)
	assert.Equal(suite.T(), req.FirstName, updatedAddress.FirstName)
	assert.Equal(suite.T(), req.LastName, updatedAddress.LastName)
	assert.Equal(suite.T(), req.Address1, updatedAddress.Address1)
	assert.Equal(suite.T(), req.IsDefault, updatedAddress.IsDefault)

	// Test address not found
	_, err = suite.service.UpdateAddress(user.ID, "nonexistent-id", req)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrAddressNotFound, err)
}

func (suite *UserServiceTestSuite) TestDeleteAddress() {
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

	// Test successful address deletion
	err := suite.service.DeleteAddress(user.ID, address.ID)
	assert.NoError(suite.T(), err)

	// Verify address is deleted
	var deletedAddress models.Address
	err = suite.db.First(&deletedAddress, "id = ?", address.ID).Error
	assert.Error(suite.T(), err) // Should not find the address

	// Test address not found
	err = suite.service.DeleteAddress(user.ID, "nonexistent-id")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrAddressNotFound, err)
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
