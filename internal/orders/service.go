package orders

import (
	"context"
	"fmt"
	"strings"

	"ecommerce-website/internal/cart"
	"ecommerce-website/internal/email"
	"ecommerce-website/internal/models"

	"gorm.io/gorm"
)

// ServiceInterface defines the interface for orders service
type ServiceInterface interface {
	CreateOrder(ctx context.Context, userID string, req *CreateOrderRequest) (*models.Order, error)
	GetOrder(orderID string, userID string) (*models.Order, error)
	GetUserOrders(userID string, page, limit int) ([]models.Order, int64, error)
	GetAllOrders(page, limit int, status, userID string) ([]models.Order, int64, error)
	UpdateOrderStatus(orderID string, status string) (*models.Order, error)
	GetAllCustomers(page, limit int, search string) ([]models.User, int64, error)
}

type Service struct {
	db           *gorm.DB
	cartService  cart.ServiceInterface
	emailService email.ServiceInterface
}

// NewService creates a new orders service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:           db,
		cartService:  cart.NewService(),
		emailService: email.NewService(),
	}
}

// NewServiceWithCartService creates a new orders service with a provided cart service
func NewServiceWithCartService(db *gorm.DB, cartService cart.ServiceInterface) *Service {
	return &Service{
		db:           db,
		cartService:  cartService,
		emailService: email.NewService(),
	}
}

// NewServiceWithDependencies creates a new orders service with all dependencies
func NewServiceWithDependencies(db *gorm.DB, cartService cart.ServiceInterface, emailService email.ServiceInterface) *Service {
	return &Service{
		db:           db,
		cartService:  cartService,
		emailService: emailService,
	}
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	SessionID       string              `json:"sessionId" binding:"required"`
	ShippingAddress models.OrderAddress `json:"shippingAddress" binding:"required"`
	BillingAddress  models.OrderAddress `json:"billingAddress" binding:"required"`
	PaymentIntentID string              `json:"paymentIntentId" binding:"required"`
	Notes           *string             `json:"notes,omitempty"`
}

// CreateOrder creates a new order from cart items
func (s *Service) CreateOrder(ctx context.Context, userID string, req *CreateOrderRequest) (*models.Order, error) {
	// Get cart with products
	cart, err := s.cartService.GetCartWithProducts(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	// Validate cart is not empty
	if cart.IsEmpty() {
		return nil, fmt.Errorf("cart is empty")
	}

	// Start database transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validate inventory and calculate totals
	var orderItems []models.OrderItem
	var subtotal float64

	for _, cartItem := range cart.Items {
		// Get current product to check inventory
		var product models.Product
		if err := tx.Where("id = ?", cartItem.ProductID).First(&product).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("product not found: %s", cartItem.ProductID)
		}

		// Check if product is active
		if !product.IsActive {
			tx.Rollback()
			return nil, fmt.Errorf("product %s is no longer available", product.Name)
		}

		// Check inventory
		if product.Inventory < cartItem.Quantity {
			tx.Rollback()
			return nil, fmt.Errorf("insufficient inventory for product %s: only %d available", product.Name, product.Inventory)
		}

		// Update inventory
		if err := tx.Model(&product).Update("inventory", product.Inventory-cartItem.Quantity).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update inventory for product %s: %w", product.Name, err)
		}

		// Create order item
		orderItem := models.OrderItem{
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     product.Price, // Use current price from database
			Total:     product.Price * float64(cartItem.Quantity),
		}
		orderItems = append(orderItems, orderItem)
		subtotal += orderItem.Total
	}

	// Calculate tax and shipping (for now, these are 0)
	tax := 0.0
	shipping := 0.0
	total := subtotal + tax + shipping

	// Create order
	order := models.Order{
		UserID:          userID,
		Status:          "pending",
		Subtotal:        subtotal,
		Tax:             tax,
		Shipping:        shipping,
		Total:           total,
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
		PaymentIntentID: req.PaymentIntentID,
		Notes:           req.Notes,
	}

	// Save order
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Save order items
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := tx.Create(&orderItems[i]).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Clear cart after successful order creation
	if err := s.cartService.ClearCart(ctx, req.SessionID); err != nil {
		// Log error but don't fail the order creation
		fmt.Printf("Warning: failed to clear cart after order creation: %v\n", err)
	}

	// Load order with items and user for response
	if err := s.db.Preload("Items.Product").Preload("User").Where("id = ?", order.ID).First(&order).Error; err != nil {
		return nil, fmt.Errorf("failed to load created order: %w", err)
	}

	return &order, nil
}

// GetOrder retrieves an order by ID
func (s *Service) GetOrder(orderID string, userID string) (*models.Order, error) {
	var order models.Order

	query := s.db.Preload("Items.Product").Preload("User")

	// If not admin, filter by user ID
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Where("id = ?", orderID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

// GetUserOrders retrieves all orders for a user with pagination
func (s *Service) GetUserOrders(userID string, page, limit int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	// Count total orders
	if err := s.db.Model(&models.Order{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get orders with pagination
	if err := s.db.Preload("Items.Product").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, total, nil
}

// UpdateOrderStatus updates the status of an order (admin only)
func (s *Service) UpdateOrderStatus(orderID string, status string) (*models.Order, error) {
	// Validate status
	validStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
		"shipped":    true,
		"delivered":  true,
		"cancelled":  true,
		"refunded":   true,
	}

	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid order status: %s", status)
	}

	// Get current order to check existing status
	var currentOrder models.Order
	if err := s.db.Preload("Items.Product").Preload("User").Where("id = ?", orderID).First(&currentOrder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get current order: %w", err)
	}

	oldStatus := currentOrder.Status

	// Update order status
	result := s.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", status)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update order status: %w", result.Error)
	}

	// Get updated order
	var order models.Order
	if err := s.db.Preload("Items.Product").Preload("User").Where("id = ?", orderID).First(&order).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated order: %w", err)
	}

	// Send email notification if status changed
	if oldStatus != status {
		if err := s.emailService.SendOrderStatusUpdate(&order, oldStatus, status); err != nil {
			// Log error but don't fail the status update
			fmt.Printf("Warning: failed to send order status update email: %v\n", err)
		}
	}

	return &order, nil
}

// GetAllOrders retrieves all orders with pagination (admin only)
func (s *Service) GetAllOrders(page, limit int, status, userID string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := s.db.Model(&models.Order{})

	// Filter by status if provided
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by user ID if provided
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Count total orders
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get orders with pagination
	query = s.db.Preload("Items.Product").Preload("User")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, total, nil
}

// GetAllCustomers retrieves all customers with pagination and search (admin only)
func (s *Service) GetAllCustomers(page, limit int, search string) ([]models.User, int64, error) {
	var customers []models.User
	var total int64

	query := s.db.Model(&models.User{}).Where("role = ? AND is_active = ?", "customer", true)

	// Add search functionality
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	// Count total customers
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get customers with pagination, including order count
	if err := query.Select("users.*, COUNT(orders.id) as order_count").
		Joins("LEFT JOIN orders ON users.id = orders.user_id").
		Group("users.id").
		Order("users.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&customers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get customers: %w", err)
	}

	// Remove password from response
	for i := range customers {
		customers[i].Password = ""
	}

	return customers, total, nil
}
