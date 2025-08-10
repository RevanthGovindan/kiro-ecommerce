package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	UserID          string    `json:"userId" gorm:"not null;index"`
	Status          string    `json:"status" gorm:"type:varchar(20);default:'pending';index"`
	Subtotal        float64   `json:"subtotal" gorm:"not null"`
	Tax             float64   `json:"tax" gorm:"default:0"`
	Shipping        float64   `json:"shipping" gorm:"default:0"`
	Total           float64   `json:"total" gorm:"not null;index"`
	ShippingAddress OrderAddress `json:"shippingAddress" gorm:"embedded;embeddedPrefix:shipping_"`
	BillingAddress  OrderAddress `json:"billingAddress" gorm:"embedded;embeddedPrefix:billing_"`
	PaymentIntentID string    `json:"paymentIntentId"`
	Notes           *string   `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	User            User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Items           []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
}

// BeforeCreate hook to generate UUID
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

type OrderAddress struct {
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	Company    *string `json:"company,omitempty"`
	Address1   string  `json:"address1"`
	Address2   *string `json:"address2,omitempty"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postalCode"`
	Country    string  `json:"country"`
	Phone      *string `json:"phone,omitempty"`
}

type OrderItem struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	OrderID   string    `json:"orderId" gorm:"not null;index"`
	ProductID string    `json:"productId" gorm:"not null;index"`
	Quantity  int       `json:"quantity" gorm:"not null"`
	Price     float64   `json:"price" gorm:"not null"` // Price at time of order
	Total     float64   `json:"total" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Order     Order     `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Product   Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

// BeforeCreate hook to generate UUID and calculate total
func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	if oi.ID == "" {
		oi.ID = uuid.New().String()
	}
	oi.Total = oi.Price * float64(oi.Quantity)
	return nil
}

// BeforeUpdate hook to recalculate total
func (oi *OrderItem) BeforeUpdate(tx *gorm.DB) error {
	oi.Total = oi.Price * float64(oi.Quantity)
	return nil
}