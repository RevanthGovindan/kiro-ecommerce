package products

import (
	"fmt"
	"math"
	"strings"

	"ecommerce-website/internal/models"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ProductFilters represents filters for product queries
type ProductFilters struct {
	CategoryID *string
	MinPrice   *float64
	MaxPrice   *float64
	InStock    *bool
	Search     *string
}

// ProductSort represents sorting options
type ProductSort struct {
	Field string // name, price, created_at
	Order string // asc, desc
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
}

// ProductListResponse represents paginated product response
type ProductListResponse struct {
	Products    []models.Product `json:"products"`
	Total       int64            `json:"total"`
	Page        int              `json:"page"`
	PageSize    int              `json:"pageSize"`
	TotalPages  int              `json:"totalPages"`
	HasNext     bool             `json:"hasNext"`
	HasPrevious bool             `json:"hasPrevious"`
}

// GetProducts retrieves products with filtering, sorting, and pagination
func (s *Service) GetProducts(filters ProductFilters, sort ProductSort, pagination PaginationParams) (*ProductListResponse, error) {
	var products []models.Product
	var total int64

	// Build base query
	query := s.db.Model(&models.Product{}).
		Preload("Category").
		Where("is_active = ?", true)

	// Apply filters
	if filters.CategoryID != nil {
		query = query.Where("category_id = ?", *filters.CategoryID)
	}

	if filters.MinPrice != nil {
		query = query.Where("price >= ?", *filters.MinPrice)
	}

	if filters.MaxPrice != nil {
		query = query.Where("price <= ?", *filters.MaxPrice)
	}

	if filters.InStock != nil && *filters.InStock {
		query = query.Where("inventory > 0")
	}

	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + strings.ToLower(*filters.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	// Apply sorting
	orderClause := "created_at DESC" // default
	if sort.Field != "" {
		validFields := map[string]bool{
			"name":       true,
			"price":      true,
			"created_at": true,
		}

		if validFields[sort.Field] {
			order := "ASC"
			if strings.ToUpper(sort.Order) == "DESC" {
				order = "DESC"
			}
			orderClause = fmt.Sprintf("%s %s", sort.Field, order)
		}
	}
	query = query.Order(orderClause)

	// Apply pagination
	if pagination.PageSize <= 0 {
		pagination.PageSize = 20 // default page size
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Offset(offset).Limit(pagination.PageSize).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	// Calculate pagination info
	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))
	hasNext := pagination.Page < totalPages
	hasPrevious := pagination.Page > 1

	return &ProductListResponse{
		Products:    products,
		Total:       total,
		Page:        pagination.Page,
		PageSize:    pagination.PageSize,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}, nil
}

// GetProductByID retrieves a single product by ID
func (s *Service) GetProductByID(id string) (*models.Product, error) {
	var product models.Product

	if err := s.db.Preload("Category").
		Where("id = ? AND is_active = ?", id, true).
		First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}

	return &product, nil
}

// SearchProducts performs text search on products
func (s *Service) SearchProducts(query string, pagination PaginationParams) (*ProductListResponse, error) {
	filters := ProductFilters{
		Search: &query,
	}

	return s.GetProducts(filters, ProductSort{}, pagination)
}

// GetCategories retrieves all active categories
func (s *Service) GetCategories() ([]models.Category, error) {
	var categories []models.Category

	if err := s.db.Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	return categories, nil
}

// GetCategoryByID retrieves a single category by ID
func (s *Service) GetCategoryByID(id string) (*models.Category, error) {
	var category models.Category

	if err := s.db.Where("id = ? AND is_active = ?", id, true).
		First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	return &category, nil
}

// Admin Product Management Methods

// CreateProduct creates a new product
func (s *Service) CreateProduct(req CreateProductRequest) (*models.Product, error) {
	// Check if category exists
	var category models.Category
	if err := s.db.Where("id = ? AND is_active = ?", req.CategoryID, true).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to verify category: %w", err)
	}

	// Check if SKU already exists
	var existingProduct models.Product
	if err := s.db.Where("sku = ?", req.SKU).First(&existingProduct).Error; err == nil {
		return nil, fmt.Errorf("sku already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check SKU uniqueness: %w", err)
	}

	// Set default values
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Create product
	product := models.Product{
		Name:           req.Name,
		Description:    req.Description,
		Price:          req.Price,
		CompareAtPrice: req.CompareAtPrice,
		SKU:            req.SKU,
		Inventory:      req.Inventory,
		CategoryID:     req.CategoryID,
		Images:         models.StringArray(req.Images),
		Specifications: models.JSONB(req.Specifications),
		SEOTitle:       req.SEOTitle,
		SEODescription: req.SEODescription,
		IsActive:       isActive,
	}

	if err := s.db.Create(&product).Error; err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Load the category relationship
	if err := s.db.Preload("Category").First(&product, "id = ?", product.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load created product: %w", err)
	}

	return &product, nil
}

// UpdateProduct updates an existing product
func (s *Service) UpdateProduct(id string, req UpdateProductRequest) (*models.Product, error) {
	// Find the product (including soft deleted ones for admin)
	var product models.Product
	if err := s.db.Unscoped().Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	// Check if category exists (if being updated)
	if req.CategoryID != nil {
		var category models.Category
		if err := s.db.Where("id = ? AND is_active = ?", *req.CategoryID, true).First(&category).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("category not found")
			}
			return nil, fmt.Errorf("failed to verify category: %w", err)
		}
	}

	// Check SKU uniqueness (if being updated)
	if req.SKU != nil && *req.SKU != product.SKU {
		var existingProduct models.Product
		if err := s.db.Where("sku = ? AND id != ?", *req.SKU, id).First(&existingProduct).Error; err == nil {
			return nil, fmt.Errorf("sku already exists")
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check SKU uniqueness: %w", err)
		}
	}

	// Update fields
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.CompareAtPrice != nil {
		updates["compare_at_price"] = *req.CompareAtPrice
	}
	if req.SKU != nil {
		updates["sku"] = *req.SKU
	}
	if req.Inventory != nil {
		updates["inventory"] = *req.Inventory
	}
	if req.CategoryID != nil {
		updates["category_id"] = *req.CategoryID
	}
	if req.Images != nil {
		updates["images"] = models.StringArray(req.Images)
	}
	if req.Specifications != nil {
		updates["specifications"] = models.JSONB(req.Specifications)
	}
	if req.SEOTitle != nil {
		updates["seo_title"] = *req.SEOTitle
	}
	if req.SEODescription != nil {
		updates["seo_description"] = *req.SEODescription
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	// Perform update
	if err := s.db.Model(&product).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Load updated product with category
	if err := s.db.Preload("Category").First(&product, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to load updated product: %w", err)
	}

	return &product, nil
}

// DeleteProduct soft deletes a product
func (s *Service) DeleteProduct(id string) error {
	// Find the product
	var product models.Product
	if err := s.db.Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("product not found")
		}
		return fmt.Errorf("failed to find product: %w", err)
	}

	// Soft delete the product
	if err := s.db.Delete(&product).Error; err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// UpdateInventory updates the inventory level of a product
func (s *Service) UpdateInventory(id string, inventory int) (*models.Product, error) {
	// Find the product
	var product models.Product
	if err := s.db.Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	// Update inventory
	if err := s.db.Model(&product).Update("inventory", inventory).Error; err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Load updated product with category
	if err := s.db.Preload("Category").First(&product, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to load updated product: %w", err)
	}

	return &product, nil
}

// GetAllProductsAdmin retrieves all products including inactive ones for admin
func (s *Service) GetAllProductsAdmin(filters AdminProductFilters, sort ProductSort, pagination PaginationParams) (*AdminProductListResponse, error) {
	var products []models.Product
	var total int64

	// Build base query (include soft deleted products)
	query := s.db.Unscoped().Model(&models.Product{}).Preload("Category")

	// Apply filters
	if filters.CategoryID != nil {
		query = query.Where("category_id = ?", *filters.CategoryID)
	}

	if filters.MinPrice != nil {
		query = query.Where("price >= ?", *filters.MinPrice)
	}

	if filters.MaxPrice != nil {
		query = query.Where("price <= ?", *filters.MaxPrice)
	}

	if filters.InStock != nil && *filters.InStock {
		query = query.Where("inventory > 0")
	}

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}

	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + strings.ToLower(*filters.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(sku) LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	// Apply sorting
	orderClause := "created_at DESC" // default
	if sort.Field != "" {
		validFields := map[string]bool{
			"name":       true,
			"price":      true,
			"created_at": true,
			"inventory":  true,
			"sku":        true,
		}

		if validFields[sort.Field] {
			order := "ASC"
			if strings.ToUpper(sort.Order) == "DESC" {
				order = "DESC"
			}
			orderClause = fmt.Sprintf("%s %s", sort.Field, order)
		}
	}
	query = query.Order(orderClause)

	// Apply pagination
	if pagination.PageSize <= 0 {
		pagination.PageSize = 20 // default page size
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Offset(offset).Limit(pagination.PageSize).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	// Calculate pagination info
	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))
	hasNext := pagination.Page < totalPages
	hasPrevious := pagination.Page > 1

	return &AdminProductListResponse{
		Products:    products,
		Total:       total,
		Page:        pagination.Page,
		PageSize:    pagination.PageSize,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}, nil
}
