package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"ecommerce-website/internal/auth"
	"ecommerce-website/internal/cart"
	"ecommerce-website/internal/config"
	"ecommerce-website/internal/database"
	"ecommerce-website/internal/orders"
	"ecommerce-website/internal/products"
	"ecommerce-website/internal/users"
	"ecommerce-website/pkg/utils"

	"github.com/gin-contrib/cors"
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

	// Configure CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://192.168.1.5:8080", "http://127.0.0.1:3000", "http://0.0.0.0:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With", "Accept", "Accept-Encoding", "Accept-Language", "Connection", "Host"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize authentication service
	authService := auth.NewService(database.GetDB(), cfg)
	authHandler := auth.NewHandler(authService)

	// Initialize product service
	productService := products.NewService(database.GetDB())
	productHandler := products.NewHandler(productService)

	// Initialize user service
	userService := users.NewService(database.GetDB())
	userHandler := users.NewHandler(userService)

	// Initialize orders service
	ordersService := orders.NewService(database.GetDB())
	ordersHandler := orders.NewHandler(ordersService)

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

	// Setup orders routes
	orders.SetupRoutes(r, ordersHandler, authService)

	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
