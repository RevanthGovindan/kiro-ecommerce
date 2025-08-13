//go:build integration
// +build integration

package integration

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ecommerce-website/internal/models"
)

// setupTestDB creates a test database and returns a cleanup function
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	// Use in-memory SQLite for tests
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Cart{},
		&models.CartItem{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

// setupTestDBWithPostgres creates a test database using PostgreSQL (for more realistic testing)
func setupTestDBWithPostgres(t *testing.T) (*gorm.DB, func()) {
	// Skip if not running with PostgreSQL
	if os.Getenv("USE_POSTGRES_TESTS") != "true" {
		t.Skip("Skipping PostgreSQL integration tests. Set USE_POSTGRES_TESTS=true to run.")
	}

	// Get database URL from environment
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://testuser:testpass@localhost:5432/ecommerce_test?sslmode=disable"
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create a unique schema for this test
	testSchema := fmt.Sprintf("test_%d", time.Now().UnixNano())
	db.Exec(fmt.Sprintf("CREATE SCHEMA %s", testSchema))
	db.Exec(fmt.Sprintf("SET search_path TO %s", testSchema))

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Cart{},
		&models.CartItem{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		// Drop the test schema
		db.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE", testSchema))

		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Setup
	log.Println("Setting up integration tests...")

	// Run tests
	code := m.Run()

	// Teardown
	log.Println("Tearing down integration tests...")

	os.Exit(code)
}

// Helper functions for creating test data

func createTestUser(db *gorm.DB, email string) *models.User {
	user := &models.User{
		Email:     email,
		Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
	}
	db.Create(user)
	return user
}

func createTestCategory(db *gorm.DB, name, slug string) *models.Category {
	category := &models.Category{
		Name:     name,
		Slug:     slug,
		IsActive: true,
	}
	db.Create(category)
	return category
}

func createTestProduct(db *gorm.DB, name, sku string, price float64, categoryID string) *models.Product {
	product := &models.Product{
		Name:        name,
		Description: fmt.Sprintf("Description for %s", name),
		Price:       price,
		SKU:         sku,
		Inventory:   10,
		CategoryID:  categoryID,
		IsActive:    true,
	}
	db.Create(product)
	return product
}

func createTestOrder(db *gorm.DB, userID string, total float64) *models.Order {
	order := &models.Order{
		UserID:   userID,
		Status:   "pending",
		Subtotal: total,
		Total:    total,
	}
	db.Create(order)
	return order
}
