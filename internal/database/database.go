package database

import (
	"fmt"
	"log"

	"ecommerce-website/internal/config"
	"ecommerce-website/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize sets up the database connection and runs migrations
func Initialize(cfg *config.Config) error {
	var err error
	
	// Configure GORM with custom logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	
	// Connect to PostgreSQL
	DB, err = gorm.Open(postgres.Open(cfg.DatabaseURL), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Get underlying sql.DB to configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	
	log.Println("Database connection established successfully")
	
	// Run migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	log.Println("Database migrations completed successfully")
	return nil
}

// runMigrations runs all database migrations
func runMigrations() error {
	// Enable UUID extension
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}
	
	// Auto-migrate all models
	err := DB.AutoMigrate(
		&models.User{},
		&models.Address{},
		&models.Category{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}
	
	// Create additional indexes for better performance
	if err := createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	
	return nil
}

// createIndexes creates additional database indexes for performance
func createIndexes() error {
	indexes := []string{
		// User indexes
		"CREATE INDEX IF NOT EXISTS idx_users_email_active ON users(email, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)",
		
		// Address indexes
		"CREATE INDEX IF NOT EXISTS idx_addresses_user_type ON addresses(user_id, type)",
		"CREATE INDEX IF NOT EXISTS idx_addresses_default ON addresses(user_id, is_default)",
		
		// Category indexes
		"CREATE INDEX IF NOT EXISTS idx_categories_parent_active ON categories(parent_id, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_categories_sort_order ON categories(sort_order)",
		
		// Product indexes
		"CREATE INDEX IF NOT EXISTS idx_products_category_active ON products(category_id, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_products_price_active ON products(price, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_products_inventory_active ON products(inventory, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_products_name_trgm ON products USING gin(name gin_trgm_ops)",
		
		// Order indexes
		"CREATE INDEX IF NOT EXISTS idx_orders_user_status ON orders(user_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_orders_total ON orders(total)",
		
		// OrderItem indexes
		"CREATE INDEX IF NOT EXISTS idx_order_items_order_product ON order_items(order_id, product_id)",
	}
	
	for _, index := range indexes {
		if err := DB.Exec(index).Error; err != nil {
			// Log warning but don't fail if index creation fails (might already exist)
			log.Printf("Warning: failed to create index: %s - %v", index, err)
		}
	}
	
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// InitializeTest sets up an in-memory SQLite database for testing
func InitializeTest(cfg *config.Config) error {
	var err error
	
	// Configure GORM with silent logger for tests
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	
	// Connect to in-memory SQLite
	DB, err = gorm.Open(sqlite.Open(":memory:"), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}
	
	// Run migrations for test database
	if err := runTestMigrations(); err != nil {
		return fmt.Errorf("failed to run test migrations: %w", err)
	}
	
	return nil
}

// runTestMigrations runs migrations for test database (SQLite)
func runTestMigrations() error {
	// Auto-migrate all models
	err := DB.AutoMigrate(
		&models.User{},
		&models.Address{},
		&models.Category{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate test database: %w", err)
	}
	
	return nil
}