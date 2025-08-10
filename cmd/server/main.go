package main

import (
	"log"
	"net/http"
	"os"

	"ecommerce-website/internal/auth"
	"ecommerce-website/internal/cart"
	"ecommerce-website/internal/config"
	"ecommerce-website/internal/database"
	"ecommerce-website/internal/products"
	"ecommerce-website/internal/users"
	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := database.Initialize(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// Initialize Redis
	if err := database.InitializeRedis(cfg); err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}
	defer database.CloseRedis()

	// Seed database if SEED_DATA environment variable is set
	if os.Getenv("SEED_DATA") == "true" {
		if err := database.SeedData(); err != nil {
			log.Printf("Warning: Failed to seed database: %v", err)
		}
	}

	r := gin.Default()

	// Initialize authentication service
	authService := auth.NewService(database.GetDB(), cfg)
	authHandler := auth.NewHandler(authService)

	// Initialize product service
	productService := products.NewService(database.GetDB())
	productHandler := products.NewHandler(productService)

	// Initialize user service
	userService := users.NewService(database.GetDB())
	userHandler := users.NewHandler(userService)

	// Setup routes
	r.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, "Ecommerce API is running", gin.H{
			"status":   "ok",
			"database": "connected",
			"redis":    "connected",
		})
	})

	// API routes group
	api := r.Group("/api")

	// Setup authentication routes
	auth.SetupRoutes(r, authHandler, authService)

	// Setup product routes
	products.SetupRoutes(r, productHandler)

	// Setup user routes
	users.SetupRoutes(r, userHandler, authService)

	// Setup cart routes
	cart.RegisterRoutes(api)

	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
