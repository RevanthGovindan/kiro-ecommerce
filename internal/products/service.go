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