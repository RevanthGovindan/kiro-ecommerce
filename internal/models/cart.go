package models

import (
	"time"
)

// CartItem represents an item in the shopping cart
type CartItem struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Total     float64 `json:"total"`
	Product   Product `json:"product,omitempty"`
}

// Cart represents a shopping cart
type Cart struct {
	SessionID string     `json:"sessionId"`
	UserID    *string    `json:"userId,omitempty"`
	Items     []CartItem `json:"items"`
	Subtotal  float64    `json:"subtotal"`
	Tax       float64    `json:"tax"`
	Total     float64    `json:"total"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// AddItemRequest represents the request to add an item to cart
type AddItemRequest struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// UpdateItemRequest represents the request to update an item in cart
type UpdateItemRequest struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"min=0"`
}

// RemoveItemRequest represents the request to remove an item from cart
type RemoveItemRequest struct {
	ProductID string `json:"productId" binding:"required"`
}

// CalculateTotals calculates and updates cart totals
func (c *Cart) CalculateTotals() {
	c.Subtotal = 0
	for i := range c.Items {
		c.Items[i].Total = c.Items[i].Price * float64(c.Items[i].Quantity)
		c.Subtotal += c.Items[i].Total
	}
	
	// For now, tax is 0 - this could be calculated based on location
	c.Tax = 0
	c.Total = c.Subtotal + c.Tax
	c.UpdatedAt = time.Now()
}

// FindItem finds an item in the cart by product ID
func (c *Cart) FindItem(productID string) *CartItem {
	for i := range c.Items {
		if c.Items[i].ProductID == productID {
			return &c.Items[i]
		}
	}
	return nil
}

// RemoveItem removes an item from the cart by product ID
func (c *Cart) RemoveItem(productID string) bool {
	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return true
		}
	}
	return false
}

// IsEmpty returns true if the cart has no items
func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}

// GetItemCount returns the total number of items in the cart
func (c *Cart) GetItemCount() int {
	count := 0
	for _, item := range c.Items {
		count += item.Quantity
	}
	return count
}