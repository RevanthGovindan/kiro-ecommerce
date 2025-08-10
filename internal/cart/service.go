package cart

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ecommerce-website/internal/database"
	"ecommerce-website/internal/models"

	"github.com/redis/go-redis/v9"
)

// ServiceInterface defines the interface for cart service
type ServiceInterface interface {
	GetCart(ctx context.Context, sessionID string) (*models.Cart, error)
	GetCartWithProducts(ctx context.Context, sessionID string) (*models.Cart, error)
	ClearCart(ctx context.Context, sessionID string) error
	AddItem(ctx context.Context, sessionID string, productID string, quantity int) (*models.Cart, error)
	UpdateItem(ctx context.Context, sessionID string, productID string, quantity int) (*models.Cart, error)
	RemoveItem(ctx context.Context, sessionID string, productID string) (*models.Cart, error)
	SaveCart(ctx context.Context, cart *models.Cart) error
}

const (
	cartKeyPrefix = "cart:"
	cartTTL       = 24 * time.Hour // Cart expires after 24 hours
)

type Service struct {
	redisClient *redis.Client
}

// NewService creates a new cart service
func NewService() *Service {
	return &Service{
		redisClient: database.GetRedisClient(),
	}
}

// GetCart retrieves a cart from Redis by session ID
func (s *Service) GetCart(ctx context.Context, sessionID string) (*models.Cart, error) {
	key := cartKeyPrefix + sessionID

	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Cart doesn't exist, return empty cart
			return &models.Cart{
				SessionID: sessionID,
				Items:     []models.CartItem{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get cart from Redis: %w", err)
	}

	var cart models.Cart
	if err := json.Unmarshal([]byte(data), &cart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart data: %w", err)
	}

	return &cart, nil
}

// SaveCart saves a cart to Redis
func (s *Service) SaveCart(ctx context.Context, cart *models.Cart) error {
	key := cartKeyPrefix + cart.SessionID

	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart data: %w", err)
	}

	if err := s.redisClient.Set(ctx, key, data, cartTTL).Err(); err != nil {
		return fmt.Errorf("failed to save cart to Redis: %w", err)
	}

	return nil
}

// AddItem adds an item to the cart
func (s *Service) AddItem(ctx context.Context, sessionID string, productID string, quantity int) (*models.Cart, error) {
	// Get current cart
	cart, err := s.GetCart(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Get product details from database
	product, err := s.getProduct(productID)
	if err != nil {
		return nil, err
	}

	// Check if product is active and has sufficient inventory
	if !product.IsActive {
		return nil, fmt.Errorf("product is not available")
	}

	if product.Inventory < quantity {
		return nil, fmt.Errorf("insufficient inventory: only %d items available", product.Inventory)
	}

	// Check if item already exists in cart
	existingItem := cart.FindItem(productID)
	if existingItem != nil {
		// Check if total quantity would exceed inventory
		totalQuantity := existingItem.Quantity + quantity
		if product.Inventory < totalQuantity {
			return nil, fmt.Errorf("insufficient inventory: only %d items available", product.Inventory)
		}
		existingItem.Quantity = totalQuantity
	} else {
		// Add new item to cart
		cartItem := models.CartItem{
			ProductID: productID,
			Quantity:  quantity,
			Price:     product.Price,
			Product:   *product,
		}
		cart.Items = append(cart.Items, cartItem)
	}

	// Recalculate totals
	cart.CalculateTotals()

	// Save cart
	if err := s.SaveCart(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

// UpdateItem updates the quantity of an item in the cart
func (s *Service) UpdateItem(ctx context.Context, sessionID string, productID string, quantity int) (*models.Cart, error) {
	// Get current cart
	cart, err := s.GetCart(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Find the item in cart
	item := cart.FindItem(productID)
	if item == nil {
		return nil, fmt.Errorf("item not found in cart")
	}

	// If quantity is 0, remove the item
	if quantity == 0 {
		cart.RemoveItem(productID)
	} else {
		// Get product details to check inventory
		product, err := s.getProduct(productID)
		if err != nil {
			return nil, err
		}

		// Check inventory
		if product.Inventory < quantity {
			return nil, fmt.Errorf("insufficient inventory: only %d items available", product.Inventory)
		}

		// Update quantity
		item.Quantity = quantity
	}

	// Recalculate totals
	cart.CalculateTotals()

	// Save cart
	if err := s.SaveCart(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

// RemoveItem removes an item from the cart
func (s *Service) RemoveItem(ctx context.Context, sessionID string, productID string) (*models.Cart, error) {
	// Get current cart
	cart, err := s.GetCart(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Remove the item
	if !cart.RemoveItem(productID) {
		return nil, fmt.Errorf("item not found in cart")
	}

	// Recalculate totals
	cart.CalculateTotals()

	// Save cart
	if err := s.SaveCart(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

// ClearCart removes all items from the cart
func (s *Service) ClearCart(ctx context.Context, sessionID string) error {
	key := cartKeyPrefix + sessionID
	return s.redisClient.Del(ctx, key).Err()
}

// GetCartWithProducts retrieves a cart and populates product details
func (s *Service) GetCartWithProducts(ctx context.Context, sessionID string) (*models.Cart, error) {
	cart, err := s.GetCart(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Populate product details for each item
	for i := range cart.Items {
		product, err := s.getProduct(cart.Items[i].ProductID)
		if err != nil {
			// If product is not found, we might want to remove it from cart
			// For now, we'll just skip it
			continue
		}
		cart.Items[i].Product = *product
		// Update price in case it changed
		cart.Items[i].Price = product.Price
	}

	// Recalculate totals in case prices changed
	cart.CalculateTotals()

	return cart, nil
}

// getProduct retrieves a product from the database
func (s *Service) getProduct(productID string) (*models.Product, error) {
	var product models.Product
	if err := database.GetDB().Where("id = ?", productID).First(&product).Error; err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}
	return &product, nil
}
