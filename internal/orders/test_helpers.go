package orders

import (
	"context"
	"testing"
	"time"

	"ecommerce-website/internal/models"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// MockCartService is a mock implementation of the cart service
type MockCartService struct {
	mock.Mock
}

func (m *MockCartService) GetCart(ctx context.Context, sessionID string) (*models.Cart, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (m *MockCartService) GetCartWithProducts(ctx context.Context, sessionID string) (*models.Cart, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (m *MockCartService) ClearCart(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockCartService) AddItem(ctx context.Context, sessionID string, productID string, quantity int) (*models.Cart, error) {
	args := m.Called(ctx, sessionID, productID, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (m *MockCartService) UpdateItem(ctx context.Context, sessionID string, productID string, quantity int) (*models.Cart, error) {
	args := m.Called(ctx, sessionID, productID, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (m *MockCartService) RemoveItem(ctx context.Context, sessionID string, productID string) (*models.Cart, error) {
	args := m.Called(ctx, sessionID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (m *MockCartService) SaveCart(ctx context.Context, cart *models.Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

// MockEmailService is a mock implementation of the email service
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendOrderStatusUpdate(order *models.Order, oldStatus, newStatus string) error {
	args := m.Called(order, oldStatus, newStatus)
	return args.Error(0)
}

// TestHelpers provides utility functions for testing orders
type TestHelpers struct {
	db *gorm.DB
}

// NewTestHelpers creates a new test helpers instance
func NewTestHelpers(db *gorm.DB) *TestHelpers {
	return &TestHelpers{db: db}
}

// CreateTestUser creates a test user in the database
func (h *TestHelpers) CreateTestUser(t *testing.T, email string) *models.User {
	user := &models.User{
		Email:     email,
		Password:  "hashedpassword",
		FirstName: "Test",
		LastName:  "User",
		Role:      "customer",
		IsActive:  true,
	}
	err := h.db.Create(user).Error
	require.NoError(t, err)
	return user
}

// CreateTestAdmin creates a test admin user in the database
func (h *TestHelpers) CreateTestAdmin(t *testing.T, email string) *models.User {
	user := &models.User{
		Email:     email,
		Password:  "hashedpassword",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
		IsActive:  true,
	}
	err := h.db.Create(user).Error
	require.NoError(t, err)
	return user
}

// CreateTestCategory creates a test category in the database
func (h *TestHelpers) CreateTestCategory(t *testing.T, name string) *models.Category {
	category := &models.Category{
		Name:     name,
		Slug:     name + "-slug",
		IsActive: true,
	}
	err := h.db.Create(category).Error
	require.NoError(t, err)
	return category
}

// CreateTestProduct creates a test product in the database
func (h *TestHelpers) CreateTestProduct(t *testing.T, categoryID string, inventory int, price float64) *models.Product {
	product := &models.Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       price,
		SKU:         "TEST-SKU-" + time.Now().Format("20060102150405"),
		Inventory:   inventory,
		IsActive:    true,
		CategoryID:  categoryID,
	}
	err := h.db.Create(product).Error
	require.NoError(t, err)
	return product
}

// CreateTestOrder creates a test order in the database
func (h *TestHelpers) CreateTestOrder(t *testing.T, userID string, status string) *models.Order {
	order := &models.Order{
		UserID:   userID,
		Status:   status,
		Subtotal: 99.99,
		Tax:      0,
		Shipping: 0,
		Total:    99.99,
		ShippingAddress: models.OrderAddress{
			FirstName:  "John",
			LastName:   "Doe",
			Address1:   "123 Main St",
			City:       "Anytown",
			State:      "CA",
			PostalCode: "12345",
			Country:    "US",
		},
		BillingAddress: models.OrderAddress{
			FirstName:  "John",
			LastName:   "Doe",
			Address1:   "123 Main St",
			City:       "Anytown",
			State:      "CA",
			PostalCode: "12345",
			Country:    "US",
		},
		PaymentIntentID: "pi_test123",
	}
	err := h.db.Create(order).Error
	require.NoError(t, err)
	return order
}

// CreateTestOrderItem creates a test order item in the database
func (h *TestHelpers) CreateTestOrderItem(t *testing.T, orderID, productID string, quantity int, price float64) *models.OrderItem {
	orderItem := &models.OrderItem{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
		Total:     price * float64(quantity),
	}
	err := h.db.Create(orderItem).Error
	require.NoError(t, err)
	return orderItem
}

// CreateTestOrderWithItems creates a complete test order with items
func (h *TestHelpers) CreateTestOrderWithItems(t *testing.T, userID string, products []struct {
	ProductID string
	Quantity  int
	Price     float64
}) *models.Order {
	// Calculate totals
	var subtotal float64
	for _, p := range products {
		subtotal += p.Price * float64(p.Quantity)
	}

	// Create order
	order := &models.Order{
		UserID:   userID,
		Status:   "pending",
		Subtotal: subtotal,
		Tax:      0,
		Shipping: 0,
		Total:    subtotal,
		ShippingAddress: models.OrderAddress{
			FirstName:  "John",
			LastName:   "Doe",
			Address1:   "123 Main St",
			City:       "Anytown",
			State:      "CA",
			PostalCode: "12345",
			Country:    "US",
		},
		BillingAddress: models.OrderAddress{
			FirstName:  "John",
			LastName:   "Doe",
			Address1:   "123 Main St",
			City:       "Anytown",
			State:      "CA",
			PostalCode: "12345",
			Country:    "US",
		},
		PaymentIntentID: "pi_test123",
	}
	err := h.db.Create(order).Error
	require.NoError(t, err)

	// Create order items
	for _, p := range products {
		orderItem := &models.OrderItem{
			OrderID:   order.ID,
			ProductID: p.ProductID,
			Quantity:  p.Quantity,
			Price:     p.Price,
			Total:     p.Price * float64(p.Quantity),
		}
		err := h.db.Create(orderItem).Error
		require.NoError(t, err)
	}

	return order
}

// GetValidOrderAddress returns a valid order address for testing
func (h *TestHelpers) GetValidOrderAddress() models.OrderAddress {
	return models.OrderAddress{
		FirstName:  "John",
		LastName:   "Doe",
		Address1:   "123 Main St",
		City:       "Anytown",
		State:      "CA",
		PostalCode: "12345",
		Country:    "US",
		Phone:      stringPtr("555-1234"),
	}
}

// GetValidCreateOrderRequest returns a valid create order request for testing
func (h *TestHelpers) GetValidCreateOrderRequest(sessionID string) *CreateOrderRequest {
	return &CreateOrderRequest{
		SessionID:       sessionID,
		ShippingAddress: h.GetValidOrderAddress(),
		BillingAddress:  h.GetValidOrderAddress(),
		PaymentIntentID: "pi_test123",
		Notes:           stringPtr("Test order notes"),
	}
}

// CleanupTestData removes all test data from the database
func (h *TestHelpers) CleanupTestData(t *testing.T) {
	// Delete in reverse order of dependencies
	err := h.db.Exec("DELETE FROM order_items").Error
	require.NoError(t, err)

	err = h.db.Exec("DELETE FROM orders").Error
	require.NoError(t, err)

	err = h.db.Exec("DELETE FROM products").Error
	require.NoError(t, err)

	err = h.db.Exec("DELETE FROM categories").Error
	require.NoError(t, err)

	err = h.db.Exec("DELETE FROM users").Error
	require.NoError(t, err)
}

// AssertOrderEquals asserts that two orders are equal
func (h *TestHelpers) AssertOrderEquals(t *testing.T, expected, actual *models.Order) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.UserID, actual.UserID)
	require.Equal(t, expected.Status, actual.Status)
	require.Equal(t, expected.Subtotal, actual.Subtotal)
	require.Equal(t, expected.Tax, actual.Tax)
	require.Equal(t, expected.Shipping, actual.Shipping)
	require.Equal(t, expected.Total, actual.Total)
	require.Equal(t, expected.PaymentIntentID, actual.PaymentIntentID)
}

// AssertOrderItemEquals asserts that two order items are equal
func (h *TestHelpers) AssertOrderItemEquals(t *testing.T, expected, actual *models.OrderItem) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.OrderID, actual.OrderID)
	require.Equal(t, expected.ProductID, actual.ProductID)
	require.Equal(t, expected.Quantity, actual.Quantity)
	require.Equal(t, expected.Price, actual.Price)
	require.Equal(t, expected.Total, actual.Total)
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
