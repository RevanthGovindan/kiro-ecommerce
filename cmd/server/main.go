package main

import (
	"net/http"
	"os"
	"time"

	"ecommerce-website/internal/auth"
	"ecommerce-website/internal/cart"
	"ecommerce-website/internal/config"
	"ecommerce-website/internal/database"
	"ecommerce-website/internal/errors"
	"ecommerce-website/internal/logger"
	"ecommerce-website/internal/middleware"
	"ecommerce-website/internal/monitoring"
	"ecommerce-website/internal/orders"
	"ecommerce-website/internal/payments"
	"ecommerce-website/internal/products"
	"ecommerce-website/internal/users"
	imageutils "ecommerce-website/internal/utils"
	"ecommerce-website/pkg/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize structured logging
	logLevel := logger.INFO
	if cfg.Environment == "development" {
		logLevel = logger.DEBUG
	}
	logger.Initialize(logger.Config{
		Level:       logLevel,
		ServiceName: "ecommerce-api",
	})

	log := logger.GetLogger()
	log.Info("Starting ecommerce API server", map[string]interface{}{
		"environment": cfg.Environment,
		"port":        cfg.Port,
	})

	// Initialize monitoring
	monitoring.Initialize()

	// Initialize database
	if err := database.Initialize(cfg); err != nil {
		log.Fatal("Failed to initialize database", err)
	}
	defer database.Close()

	// Initialize Redis
	if err := database.InitializeRedis(cfg); err != nil {
		log.Fatal("Failed to initialize Redis", err)
	}
	defer database.CloseRedis()

	// Seed database if SEED_DATA environment variable is set
	if os.Getenv("SEED_DATA") == "true" {
		if err := database.SeedData(); err != nil {
			log.Warn("Failed to seed database", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	r := gin.New()

	// Add comprehensive middleware stack
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.ErrorHandlingMiddleware())

	// Initialize image optimization config
	imageutils.DefaultImageConfig.CDNBaseURL = cfg.CDNBaseURL

	// Security middleware
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.InputSanitizationMiddleware())
	r.Use(middleware.ValidateContentLength(cfg.MaxRequestSize))

	// Configure CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://192.168.1.5:8080", "http://127.0.0.1:3000", "http://0.0.0.0:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With", "Accept", "Accept-Encoding", "Accept-Language", "Connection", "Host"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "X-Cache"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// General rate limiting
	r.Use(middleware.RateLimitMiddleware(middleware.GeneralRateLimit))

	// Cache invalidation middleware
	r.Use(middleware.CacheInvalidationMiddleware())

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

	// Initialize payments service
	paymentsService := payments.NewService(database.GetDB(), cfg.RazorpayKeyID, cfg.RazorpaySecret)
	paymentsHandler := payments.NewHandler(paymentsService)

	// Initialize error handling service
	errorHandler := errors.NewHandler()

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

	// Setup authentication routes with stricter rate limiting
	authGroup := r.Group("/api/auth")
	authGroup.Use(middleware.RateLimitMiddleware(middleware.AuthRateLimit))
	auth.SetupRoutes(r, authHandler, authService)

	// Setup product routes with caching
	productGroup := r.Group("/api/products")
	productGroup.Use(middleware.CacheMiddleware(middleware.ProductCatalogCache))
	products.SetupRoutes(r, productHandler, authService)

	// Setup category routes with caching
	categoryGroup := r.Group("/api/categories")
	categoryGroup.Use(middleware.CacheMiddleware(middleware.CategoryCache))

	// Setup user routes
	users.SetupRoutes(r, userHandler, authService)

	// Setup cart routes
	cart.RegisterRoutes(api)

	// Setup orders routes
	orders.SetupRoutes(r, ordersHandler, authService)

	// Setup payments routes with stricter rate limiting
	paymentGroup := r.Group("/api/payments")
	paymentGroup.Use(middleware.RateLimitMiddleware(middleware.PaymentRateLimit))
	payments.SetupRoutes(r, paymentsHandler, authService)

	// Setup error handling and monitoring routes
	errors.SetupRoutes(r, errorHandler, authService)

	// Serve static files in development mode
	if cfg.Environment == "development" {
		log.Info("Development mode: serving static files from api/ directory", map[string]interface{}{
			"static_path": "/api-docs",
			"files_dir":   "./api",
		})
		r.Static("/api-docs", "./api")
		r.GET("/", func(c *gin.Context) {
			log.Debug("Root path accessed, redirecting to API documentation")
			c.Redirect(http.StatusMovedPermanently, "/api-docs/")
		})

		// Add a specific route for the API documentation
		r.GET("/docs", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/api-docs/")
		})
	}

	// Setup admin routes with admin rate limiting
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(authService.AuthMiddleware())
	adminGroup.Use(authService.AdminMiddleware())
	adminGroup.Use(middleware.RateLimitMiddleware(middleware.AdminRateLimit))

	log.Info("Server starting", map[string]interface{}{
		"port":        cfg.Port,
		"environment": cfg.Environment,
	})

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
