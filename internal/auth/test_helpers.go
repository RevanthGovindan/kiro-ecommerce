package auth

import (
	"testing"
	"time"

	"ecommerce-website/internal/config"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUser is a simplified user model for testing with SQLite
type TestUser struct {
	ID                     string     `json:"id" gorm:"primaryKey"`
	Email                  string     `json:"email" gorm:"uniqueIndex;not null"`
	Password               string     `json:"-" gorm:"not null"`
	FirstName              string     `json:"firstName" gorm:"not null"`
	LastName               string     `json:"lastName" gorm:"not null"`
	Phone                  *string    `json:"phone,omitempty"`
	Role                   string     `json:"role" gorm:"default:'customer'"`
	IsActive               bool       `json:"isActive" gorm:"default:true"`
	EmailVerified          bool       `json:"emailVerified" gorm:"default:false"`
	EmailVerificationToken *string    `json:"-" gorm:"type:varchar(255)"`
	PasswordResetToken     *string    `json:"-" gorm:"type:varchar(255)"`
	PasswordResetExpiry    *time.Time `json:"-"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
}

func (TestUser) TableName() string {
	return "users"
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	return db
}

func setupTestService(t *testing.T) (*Service, *gorm.DB) {
	db := setupTestDB(t)
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}
	service := NewService(db, cfg)
	return service, db
}

func setupTestHandler(t *testing.T) (*Handler, *gorm.DB) {
	service, db := setupTestService(t)
	handler := NewHandler(service)
	return handler, db
}