package main

import (
	"flag"
	"log"

	"ecommerce-website/internal/config"
	"ecommerce-website/internal/database"
)

func main() {
	var (
		seed = flag.Bool("seed", false, "Run database seeding after migration")
		drop = flag.Bool("drop", false, "Drop all tables before migration (DANGEROUS)")
	)
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	log.Println("Starting database migration...")

	// Initialize database connection
	if err := database.Initialize(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// Drop tables if requested (for development only)
	if *drop {
		log.Println("WARNING: Dropping all tables...")
		if err := dropAllTables(); err != nil {
			log.Fatal("Failed to drop tables:", err)
		}
		log.Println("All tables dropped successfully")

		// Re-run migrations after dropping
		if err := database.Initialize(cfg); err != nil {
			log.Fatal("Failed to re-initialize database after drop:", err)
		}
	}

	// Run seeding if requested
	if *seed {
		log.Println("Running database seeding...")
		if err := database.SeedData(); err != nil {
			log.Fatal("Failed to seed database:", err)
		}
		log.Println("Database seeding completed successfully")
	}

	log.Println("Migration completed successfully!")
}

func dropAllTables() error {
	// Get database instance
	db := database.DB

	// First, drop all foreign key constraints
	log.Println("Dropping foreign key constraints...")
	constraints := []string{
		"ALTER TABLE addresses DROP CONSTRAINT IF EXISTS fk_users_addresses CASCADE",
		"ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_users_orders CASCADE",
		"ALTER TABLE order_items DROP CONSTRAINT IF EXISTS fk_orders_order_items CASCADE",
		"ALTER TABLE order_items DROP CONSTRAINT IF EXISTS fk_products_order_items CASCADE",
		"ALTER TABLE products DROP CONSTRAINT IF EXISTS fk_categories_products CASCADE",
		"ALTER TABLE categories DROP CONSTRAINT IF EXISTS fk_categories_children CASCADE",
	}

	for _, constraint := range constraints {
		if err := db.Exec(constraint).Error; err != nil {
			log.Printf("Warning: failed to drop constraint: %v", err)
		}
	}

	// Drop tables in reverse order to handle foreign key constraints
	tables := []string{
		"order_items",
		"orders",
		"products",
		"categories",
		"addresses",
		"users",
	}

	for _, table := range tables {
		if err := db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE").Error; err != nil {
			return err
		}
		log.Printf("Dropped table: %s", table)
	}

	return nil
}
