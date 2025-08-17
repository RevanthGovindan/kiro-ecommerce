package products

import (
	"fmt"
	"math"
	"strings"

	"ecommerce-website/internal/models"
	"ecommerce-website/internal/search"

	"gorm.io/gorm"
)

type Service struct {
	db            *gorm.DB
	searchService *search.Service
}

func NewService(db *gorm.DB) *Service {
	searchService := search.NewService(db)
	return &Service{
		db:            db,
		searchService: searchService,
	}
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

// AdvancedSearchProducts performs advanced search with Elasticsearch integration
func (s *Service) AdvancedSearchProducts(filters AdvancedSearchFilters, sort AdvancedSearchSort, page, pageSize int, includeFacets bool) (*AdvancedSearchResponse, error) {
	// Convert filters to search.SearchFilters
	searchFilters := search.SearchFilters{
		CategoryID: filters.CategoryID,
		MinPrice:   filters.MinPrice,
		MaxPrice:   filters.MaxPrice,
		InStock:    filters.InStock,
		Search:     filters.Search,
	}

	// Convert sort to search.SearchSort
	searchSort := search.SearchSort{
		Field: sort.Field,
		Order: sort.Order,
	}

	// Use search service for advanced search
	searchResponse, err := s.searchService.SearchProducts(searchFilters, searchSort, page, pageSize, includeFacets)
	if err != nil {
		return nil, fmt.Errorf("failed to perform advanced search: %w", err)
	}

	// Convert search.SearchResponse to AdvancedSearchResponse
	advancedResponse := &AdvancedSearchResponse{
		Products:   searchResponse.Products,
		Total:      searchResponse.Total,
		Page:       searchResponse.Page,
		PageSize:   searchResponse.PageSize,
		TotalPages: searchResponse.TotalPages,
	}

	// Add suggestions if available
	if searchResponse.Suggestions != nil {
		advancedResponse.Suggestions = searchResponse.Suggestions
	}

	// Add facets if available
	if searchResponse.Facets != nil {
		advancedResponse.Facets = &SearchFacets{
			Categories:  make([]CategoryFacet, len(searchResponse.Facets.Categories)),
			PriceRanges: make([]PriceRangeFacet, len(searchResponse.Facets.PriceRanges)),
		}

		// Convert category facets
		for i, cf := range searchResponse.Facets.Categories {
			advancedResponse.Facets.Categories[i] = CategoryFacet{
				ID:    cf.ID,
				Name:  cf.Name,
				Count: cf.Count,
			}
		}

		// Convert price range facets
		for i, prf := range searchResponse.Facets.PriceRanges {
			advancedResponse.Facets.PriceRanges[i] = PriceRangeFacet{
				Range: prf.Range,
				Min:   prf.Min,
				Max:   prf.Max,
				Count: prf.Count,
			}
		}
	}

	return advancedResponse, nil
}

// GetSearchSuggestions returns search suggestions
func (s *Service) GetSearchSuggestions(query string, size int) ([]string, error) {
	return s.searchService.GetSuggestions(query, size)
}

// buildFacets builds facets for advanced search
func (s *Service) buildFacets(filters ProductFilters) (*SearchFacets, error) {
	facets := &SearchFacets{}

	// Build category facets
	categoryFacets, err := s.buildCategoryFacets(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to build category facets: %w", err)
	}
	facets.Categories = categoryFacets

	// Build price range facets
	priceRangeFacets, err := s.buildPriceRangeFacets(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to build price range facets: %w", err)
	}
	facets.PriceRanges = priceRangeFacets

	return facets, nil
}

// buildCategoryFacets builds category facets
func (s *Service) buildCategoryFacets(filters ProductFilters) ([]CategoryFacet, error) {
	type CategoryCount struct {
		CategoryID   string `json:"category_id"`
		CategoryName string `json:"category_name"`
		Count        int64  `json:"count"`
	}

	var categoryCounts []CategoryCount

	query := s.db.Table("products").
		Select("products.category_id, categories.name as category_name, COUNT(*) as count").
		Joins("LEFT JOIN categories ON products.category_id = categories.id").
		Where("products.is_active = ? AND categories.is_active = ?", true, true).
		Group("products.category_id, categories.name")

	// Apply filters (excluding category filter for facets)
	if filters.MinPrice != nil {
		query = query.Where("products.price >= ?", *filters.MinPrice)
	}

	if filters.MaxPrice != nil {
		query = query.Where("products.price <= ?", *filters.MaxPrice)
	}

	if filters.InStock != nil && *filters.InStock {
		query = query.Where("products.inventory > 0")
	}

	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + *filters.Search + "%"
		query = query.Where("LOWER(products.name) LIKE LOWER(?) OR LOWER(products.description) LIKE LOWER(?)", searchTerm, searchTerm)
	}

	if err := query.Find(&categoryCounts).Error; err != nil {
		return nil, err
	}

	var facets []CategoryFacet
	for _, cc := range categoryCounts {
		facets = append(facets, CategoryFacet{
			ID:    cc.CategoryID,
			Name:  cc.CategoryName,
			Count: cc.Count,
		})
	}

	return facets, nil
}

// buildPriceRangeFacets builds price range facets
func (s *Service) buildPriceRangeFacets(filters ProductFilters) ([]PriceRangeFacet, error) {
	priceRanges := []struct {
		Range string
		Min   float64
		Max   *float64
	}{
		{"0-25", 0, &[]float64{25}[0]},
		{"25-50", 25, &[]float64{50}[0]},
		{"50-100", 50, &[]float64{100}[0]},
		{"100-250", 100, &[]float64{250}[0]},
		{"250+", 250, nil},
	}

	var facets []PriceRangeFacet

	for _, pr := range priceRanges {
		query := s.db.Model(&models.Product{}).
			Where("is_active = ?", true)

		// Apply filters (excluding price filter for facets)
		if filters.CategoryID != nil {
			query = query.Where("category_id = ?", *filters.CategoryID)
		}

		if filters.InStock != nil && *filters.InStock {
			query = query.Where("inventory > 0")
		}

		if filters.Search != nil && *filters.Search != "" {
			searchTerm := "%" + *filters.Search + "%"
			query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", searchTerm, searchTerm)
		}

		// Apply price range
		query = query.Where("price >= ?", pr.Min)
		if pr.Max != nil {
			query = query.Where("price <= ?", *pr.Max)
		}

		var count int64
		if err := query.Count(&count).Error; err != nil {
			return nil, err
		}

		maxValue := float64(999999)
		if pr.Max != nil {
			maxValue = *pr.Max
		}

		facets = append(facets, PriceRangeFacet{
			Range: pr.Range,
			Min:   pr.Min,
			Max:   maxValue,
			Count: count,
		})
	}

	return facets, nil
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

// CreateCategory creates a new category
func (s *Service) CreateCategory(req CreateCategoryRequest) (*models.Category, error) {
	// Check if slug already exists
	var existingCategory models.Category
	if err := s.db.Where("slug = ?", req.Slug).First(&existingCategory).Error; err == nil {
		return nil, fmt.Errorf("category with slug already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check slug uniqueness: %w", err)
	}

	// Check if parent category exists (if provided)
	if req.ParentID != nil {
		var parentCategory models.Category
		if err := s.db.Where("id = ? AND is_active = ?", *req.ParentID, true).First(&parentCategory).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("parent category not found")
			}
			return nil, fmt.Errorf("failed to verify parent category: %w", err)
		}
	}

	// Set default values
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	// Create category
	category := models.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ParentID:    req.ParentID,
		IsActive:    isActive,
		SortOrder:   sortOrder,
	}

	if err := s.db.Create(&category).Error; err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	// Load the parent relationship if exists
	if req.ParentID != nil {
		if err := s.db.Preload("Parent").First(&category, "id = ?", category.ID).Error; err != nil {
			return nil, fmt.Errorf("failed to load created category: %w", err)
		}
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

	// Index the product in search service
	if err := s.searchService.IndexProduct(&product); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to index product in search: %v\n", err)
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

	// Re-index the product in search service
	if err := s.searchService.IndexProduct(&product); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to re-index product in search: %v\n", err)
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

	// Remove the product from search index
	if err := s.searchService.DeleteProduct(id); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to delete product from search index: %v\n", err)
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

	// Re-index the product in search service
	if err := s.searchService.IndexProduct(&product); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to re-index product in search: %v\n", err)
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
