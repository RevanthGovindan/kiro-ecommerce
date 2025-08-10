package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payment struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	OrderID           string    `json:"orderId" gorm:"not null;index"`
	RazorpayOrderID   string    `json:"razorpayOrderId" gorm:"unique;not null"`
	RazorpayPaymentID *string   `json:"razorpayPaymentId,omitempty" gorm:"unique"`
	RazorpaySignature *string   `json:"razorpaySignature,omitempty"`
	Amount            int64     `json:"amount" gorm:"not null"` // Amount in paise
	Currency          string    `json:"currency" gorm:"default:'INR'"`
	Status            string    `json:"status" gorm:"type:varchar(20);default:'created';index"`
	Method            *string   `json:"method,omitempty"`
	Description       *string   `json:"description,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	Order             Order     `json:"order,omitempty" gorm:"foreignKey:OrderID"`
}

// BeforeCreate hook to generate UUID
func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// Payment status constants
const (
	PaymentStatusCreated   = "created"
	PaymentStatusPaid      = "paid"
	PaymentStatusFailed    = "failed"
	PaymentStatusCancelled = "cancelled"
)
