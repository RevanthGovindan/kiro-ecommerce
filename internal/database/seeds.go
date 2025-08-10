package database

import (
	"log"

	"ecommerce-website/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// SeedData populates the database with initial data for development and testing
func SeedData() error {
	log.Println("Starting database seeding...")
	
	// Seed categories
	if err := seedCategories(); err != nil {
		return err
	}
	
	// Seed users
	if err := seedUsers(); err != nil {
		return err
	}
	
	// Seed products
	if err := seedProducts(); err != nil {
		return err
	}
	
	log.Println("Database seeding completed successfully")
	return nil
}

func seedCategories() error {
	categories := []models.Category{
		{
			Name:        "Electronics",
			Slug:        "electronics",
			Description: stringPtr("Electronic devices and accessories"),
			IsActive:    true,
			SortOrder:   1,
		},
		{
			Name:        "Clothing",
			Slug:        "clothing",
			Description: stringPtr("Fashion and apparel"),
			IsActive:    true,
			SortOrder:   2,
		},
		{
			Name:        "Books",
			Slug:        "books",
			Description: stringPtr("Books and literature"),
			IsActive:    true,
			SortOrder:   3,
		},
		{
			Name:        "Home & Garden",
			Slug:        "home-garden",
			Description: stringPtr("Home improvement and garden supplies"),
			IsActive:    true,
			SortOrder:   4,
		},
	}
	
	for _, category := range categories {
		var existingCategory models.Category
		if err := DB.Where("slug = ?", category.Slug).First(&existingCategory).Error; err != nil {
			// Category doesn't exist, create it
			if err := DB.Create(&category).Error; err != nil {
				return err
			}
			log.Printf("Created category: %s", category.Name)
		}
	}
	
	// Create subcategories
	var electronicsCategory models.Category
	if err := DB.Where("slug = ?", "electronics").First(&electronicsCategory).Error; err == nil {
		subcategories := []models.Category{
			{
				Name:        "Smartphones",
				Slug:        "smartphones",
				Description: stringPtr("Mobile phones and accessories"),
				ParentID:    &electronicsCategory.ID,
				IsActive:    true,
				SortOrder:   1,
			},
			{
				Name:        "Laptops",
				Slug:        "laptops",
				Description: stringPtr("Portable computers"),
				ParentID:    &electronicsCategory.ID,
				IsActive:    true,
				SortOrder:   2,
			},
		}
		
		for _, subcategory := range subcategories {
			var existingSubcategory models.Category
			if err := DB.Where("slug = ?", subcategory.Slug).First(&existingSubcategory).Error; err != nil {
				if err := DB.Create(&subcategory).Error; err != nil {
					return err
				}
				log.Printf("Created subcategory: %s", subcategory.Name)
			}
		}
	}
	
	return nil
}

func seedUsers() error {
	// Hash password for test users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	users := []models.User{
		{
			Email:     "admin@example.com",
			Password:  string(hashedPassword),
			FirstName: "Admin",
			LastName:  "User",
			Role:      "admin",
			IsActive:  true,
		},
		{
			Email:     "customer@example.com",
			Password:  string(hashedPassword),
			FirstName: "John",
			LastName:  "Doe",
			Phone:     stringPtr("+1234567890"),
			Role:      "customer",
			IsActive:  true,
		},
		{
			Email:     "jane@example.com",
			Password:  string(hashedPassword),
			FirstName: "Jane",
			LastName:  "Smith",
			Phone:     stringPtr("+1987654321"),
			Role:      "customer",
			IsActive:  true,
		},
	}
	
	for _, user := range users {
		var existingUser models.User
		if err := DB.Where("email = ?", user.Email).First(&existingUser).Error; err != nil {
			// User doesn't exist, create it
			if err := DB.Create(&user).Error; err != nil {
				return err
			}
			log.Printf("Created user: %s", user.Email)
			
			// Create sample address for customers
			if user.Role == "customer" {
				address := models.Address{
					UserID:     user.ID,
					Type:       "shipping",
					FirstName:  user.FirstName,
					LastName:   user.LastName,
					Address1:   "123 Main St",
					City:       "Anytown",
					State:      "CA",
					PostalCode: "12345",
					Country:    "US",
					Phone:      user.Phone,
					IsDefault:  true,
				}
				if err := DB.Create(&address).Error; err != nil {
					log.Printf("Warning: failed to create address for user %s: %v", user.Email, err)
				}
			}
		}
	}
	
	return nil
}

func seedProducts() error {
	// Get categories for products
	var electronicsCategory, smartphonesCategory, laptopsCategory, clothingCategory models.Category
	
	DB.Where("slug = ?", "electronics").First(&electronicsCategory)
	DB.Where("slug = ?", "smartphones").First(&smartphonesCategory)
	DB.Where("slug = ?", "laptops").First(&laptopsCategory)
	DB.Where("slug = ?", "clothing").First(&clothingCategory)
	
	products := []models.Product{
		{
			Name:        "iPhone 15 Pro",
			Description: "Latest iPhone with advanced camera system and A17 Pro chip",
			Price:       999.99,
			CompareAtPrice: float64Ptr(1099.99),
			SKU:         "IPHONE15PRO-128",
			Inventory:   50,
			CategoryID:  smartphonesCategory.ID,
			Images:      models.StringArray{"iphone15pro-1.jpg", "iphone15pro-2.jpg"},
			Specifications: models.JSONB{
				"storage":     "128GB",
				"color":       "Natural Titanium",
				"display":     "6.1-inch Super Retina XDR",
				"camera":      "48MP Main camera",
				"connectivity": "5G",
			},
			SEOTitle:       stringPtr("iPhone 15 Pro - Latest Apple Smartphone"),
			SEODescription: stringPtr("Get the latest iPhone 15 Pro with advanced features and premium design"),
			IsActive:       true,
		},
		{
			Name:        "MacBook Pro 14-inch",
			Description: "Powerful laptop with M3 chip for professional work",
			Price:       1999.99,
			SKU:         "MBP14-M3-512",
			Inventory:   25,
			CategoryID:  laptopsCategory.ID,
			Images:      models.StringArray{"macbook-pro-14-1.jpg", "macbook-pro-14-2.jpg"},
			Specifications: models.JSONB{
				"processor": "Apple M3",
				"memory":    "16GB",
				"storage":   "512GB SSD",
				"display":   "14.2-inch Liquid Retina XDR",
				"ports":     "3x Thunderbolt 4, HDMI, SD card slot",
			},
			SEOTitle:       stringPtr("MacBook Pro 14-inch with M3 Chip"),
			SEODescription: stringPtr("Professional laptop with incredible performance and battery life"),
			IsActive:       true,
		},
		{
			Name:        "Samsung Galaxy S24",
			Description: "Android flagship with AI-powered features",
			Price:       799.99,
			SKU:         "GALAXY-S24-256",
			Inventory:   40,
			CategoryID:  smartphonesCategory.ID,
			Images:      models.StringArray{"galaxy-s24-1.jpg", "galaxy-s24-2.jpg"},
			Specifications: models.JSONB{
				"storage":     "256GB",
				"color":       "Phantom Black",
				"display":     "6.2-inch Dynamic AMOLED 2X",
				"camera":      "50MP Triple camera",
				"connectivity": "5G",
			},
			SEOTitle:       stringPtr("Samsung Galaxy S24 - AI-Powered Android Phone"),
			SEODescription: stringPtr("Experience the future with Galaxy S24's AI features and premium design"),
			IsActive:       true,
		},
		{
			Name:        "Classic Cotton T-Shirt",
			Description: "Comfortable 100% cotton t-shirt in various colors",
			Price:       24.99,
			SKU:         "COTTON-TEE-M-BLUE",
			Inventory:   100,
			CategoryID:  clothingCategory.ID,
			Images:      models.StringArray{"cotton-tee-blue-1.jpg", "cotton-tee-blue-2.jpg"},
			Specifications: models.JSONB{
				"material": "100% Cotton",
				"size":     "Medium",
				"color":    "Blue",
				"fit":      "Regular",
				"care":     "Machine washable",
			},
			SEOTitle:       stringPtr("Classic Cotton T-Shirt - Comfortable Everyday Wear"),
			SEODescription: stringPtr("Soft and comfortable cotton t-shirt perfect for casual wear"),
			IsActive:       true,
		},
		{
			Name:        "Wireless Bluetooth Headphones",
			Description: "High-quality wireless headphones with noise cancellation",
			Price:       149.99,
			CompareAtPrice: float64Ptr(199.99),
			SKU:         "BT-HEADPHONES-NC",
			Inventory:   75,
			CategoryID:  electronicsCategory.ID,
			Images:      models.StringArray{"bt-headphones-1.jpg", "bt-headphones-2.jpg"},
			Specifications: models.JSONB{
				"connectivity":      "Bluetooth 5.0",
				"battery_life":      "30 hours",
				"noise_cancellation": true,
				"weight":           "250g",
				"color":            "Black",
			},
			SEOTitle:       stringPtr("Wireless Bluetooth Headphones with Noise Cancellation"),
			SEODescription: stringPtr("Premium wireless headphones with superior sound quality and long battery life"),
			IsActive:       true,
		},
	}
	
	for _, product := range products {
		var existingProduct models.Product
		if err := DB.Where("sku = ?", product.SKU).First(&existingProduct).Error; err != nil {
			// Product doesn't exist, create it
			if err := DB.Create(&product).Error; err != nil {
				return err
			}
			log.Printf("Created product: %s", product.Name)
		}
	}
	
	return nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}