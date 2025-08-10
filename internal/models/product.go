package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONB type for PostgreSQL JSONB fields
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, j)
}

// StringArray type for PostgreSQL text[] fields
type StringArray []string

// Value implements the driver.Valuer interface
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "{}", nil
	}
	
	// Format as PostgreSQL array literal
	result := "{"
	for i, str := range s {
		if i > 0 {
			result += ","
		}
		// Escape quotes and wrap in quotes
		escaped := strings.ReplaceAll(str, `"`, `\"`)
		result += `"` + escaped + `"`
	}
	result += "}"
	return result, nil
}

// Scan implements the sql.Scanner interface
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return s.scanBytes(v)
	case string:
		return s.scanBytes([]byte(v))
	default:
		return errors.New("cannot scan into StringArray")
	}
}

func (s *StringArray) scanBytes(src []byte) error {
	str := string(src)
	if str == "{}" {
		*s = StringArray{}
		return nil
	}
	
	// Remove braces and split by comma
	str = strings.Trim(str, "{}")
	if str == "" {
		*s = StringArray{}
		return nil
	}
	
	// Simple parsing - this could be improved for more complex cases
	parts := strings.Split(str, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		// Remove quotes and unescape
		part = strings.Trim(part, `"`)
		part = strings.ReplaceAll(part, `\"`, `"`)
		result[i] = part
	}
	
	*s = StringArray(result)
	return nil
}

type Product struct {
	ID             string      `json:"id" gorm:"primaryKey"`
	Name           string      `json:"name" gorm:"not null;index"`
	Description    string      `json:"description"`
	Price          float64     `json:"price" gorm:"not null;index"`
	CompareAtPrice *float64    `json:"compareAtPrice,omitempty"`
	SKU            string      `json:"sku" gorm:"uniqueIndex;not null"`
	Inventory      int         `json:"inventory" gorm:"default:0;index"`
	IsActive       bool        `json:"isActive" gorm:"default:true;index"`
	CategoryID     string      `json:"categoryId" gorm:"not null;index"`
	Images         StringArray `json:"images" gorm:"type:text[]"`
	Specifications JSONB       `json:"specifications" gorm:"type:jsonb"`
	SEOTitle       *string     `json:"seoTitle,omitempty"`
	SEODescription *string     `json:"seoDescription,omitempty"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
	Category       Category    `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	OrderItems     []OrderItem `json:"orderItems,omitempty" gorm:"foreignKey:ProductID"`
}

// BeforeCreate hook to generate UUID
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}