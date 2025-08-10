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