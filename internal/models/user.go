package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                   string     `json:"id" gorm:"primaryKey"`
	Email                string     `json:"email" gorm:"uniqueIndex;not null"`
	Password             string     `json:"-" gorm:"not null"` // hashed
	FirstName            string     `json:"firstName" gorm:"not null"`
	LastName             string     `json:"lastName" gorm:"not null"`
	Phone                *string    `json:"phone,omitempty"`
	Role                 string     `json:"role" gorm:"type:varchar(20);default:'customer'"`
	IsActive             bool       `json:"isActive" gorm:"default:true"`
	EmailVerified        bool       `json:"emailVerified" gorm:"default:false"`
	EmailVerificationToken *string  `json:"-" gorm:"type:varchar(255)"`
	PasswordResetToken   *string    `json:"-" gorm:"type:varchar(255)"`
	PasswordResetExpiry  *time.Time `json:"-"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	Addresses            []Address  `json:"addresses,omitempty" gorm:"foreignKey:UserID"`
	Orders               []Order    `json:"orders,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to generate UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

type Address struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	UserID     string    `json:"userId" gorm:"not null"`
	Type       string    `json:"type" gorm:"type:varchar(20);not null"` // shipping, billing
	FirstName  string    `json:"firstName" gorm:"not null"`
	LastName   string    `json:"lastName" gorm:"not null"`
	Company    *string   `json:"company,omitempty"`
	Address1   string    `json:"address1" gorm:"not null"`
	Address2   *string   `json:"address2,omitempty"`
	City       string    `json:"city" gorm:"not null"`
	State      string    `json:"state" gorm:"not null"`
	PostalCode string    `json:"postalCode" gorm:"not null"`
	Country    string    `json:"country" gorm:"not null"`
	Phone      *string   `json:"phone,omitempty"`
	IsDefault  bool      `json:"isDefault" gorm:"default:false"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	User       User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to generate UUID
func (a *Address) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}