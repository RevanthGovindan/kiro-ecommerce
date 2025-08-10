package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"not null"`
	Slug        string     `json:"slug" gorm:"uniqueIndex;not null"`
	Description *string    `json:"description,omitempty"`
	ParentID    *string    `json:"parentId,omitempty"`
	IsActive    bool       `json:"isActive" gorm:"default:true"`
	SortOrder   int        `json:"sortOrder" gorm:"default:0"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
	Products    []Product  `json:"products,omitempty" gorm:"foreignKey:CategoryID"`
	Children    []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Parent      *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
}

// BeforeCreate hook to generate UUID
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}