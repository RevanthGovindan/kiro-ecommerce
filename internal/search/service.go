package search

import (
	"fmt"
	"log"

	"ecommerce-website/internal/models"

	"gorm.io/gorm"
)

type Service struct {
	db             *gorm.DB
	elasticsearch  *ElasticsearchService
	fallbackSearch bool
}

func NewService(db *gorm.DB) *Service {
	es, err := NewElasticsearchService()
	if err != nil {
		log.Printf("Warning: Elasticsearch not available, falling back to database search: %v", err)
		return &Service{
			db:             db,
			elasticsearch:  nil,
			fallbackSearch: true,
		}
	}

	return &Service{
		db:             db,
		elasticsearch:  es,
		fallbackSearch: false,
	}
}

// SearchProducts performs advanced product search
func (s *Service) SearchProducts(filters SearchFilters, sort SearchSort, page, pageSize int, includeFacets bool) (*SearchResponse, error) {
	// Use Elasticsearch if available
	if !s.fallbackSearch && s.elasticsearch != nil {
		return s.elasticsearch.SearchProducts(filters, sort, page, pageSize, includeFacets)
	}

	// Fallback to database search
	return s.fallbackDatabaseSearch(filters, sort, page, pageSize)
}

// GetSuggestions returns search suggestions
func (s *Service) GetSuggestions(query string, size int) ([]string, error) {
	if !s.fallbackSearch && s.elasticsearch != nil {
		return s.elasticsearch.GetSuggestions(query, size)
	}

	// Fallback to database-based suggestions
	return s.fallbackDatabaseSuggestions(query, size)
}

// IndexProduct indexes a product in Elasticsearch
func (s *Service) IndexProduct(product *models.Product) error {
	if s.fallbackSearch || s.elasticsearch == nil {
		return nil // No-op if Elasticsearch is not available
	}

	return s.elasticsearch.IndexProduct(product)
}

// DeleteProduct removes a product from Elasticsearch index
func (s *Service) DeleteProduct(productID string) error {
	if s.fallbackSearch || s.elasticsearch == nil {
		return nil // No-op if Elasticsearch is not available
	}

	return s.elasticsearch.DeleteProduct(productID)
}

// ReindexAllProducts reindexes all products from database to Elasticsearch
func (s *Service) ReindexAllProducts() error {
	if s.fallbackSearch || s.elasticsearch == nil {
		return fmt.Errorf("Elasticsearch not available")
	}

	var products []models.Product
	if err := s.db.Preload("Category").Where("is_active = ?", true).Find(&products).Error; err != nil {
		return fmt.Errorf("failed to fetch products for reindexing: %w", err)
	}

	for _, product := range products {
		if err := s.elasticsearch.IndexProduct(&product); err != nil {
			log.Printf("Failed to index product %s: %v", product.ID, err)
		}
	}

	log.Printf("Reindexed %d products", len(products))
	return nil
}

// fallbackDatabaseSearch performs search using database queries
func (s *Service) fallbackDatabaseSearch(filters SearchFilters, sort SearchSort, page, pageSize int) (*SearchResponse, error) {
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
		searchTerm := "%" + *filters.Search + "%"
		query = query.Where("LOWER(name) ILIKE ? OR LOWER(description) ILIKE ?", searchTerm, searchTerm)
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
			if sort.Order == "desc" {
				order = "DESC"
			}
			orderClause = fmt.Sprintf("%s %s", sort.Field, order)
		}
	}
	query = query.Order(orderClause)

	// Apply pagination
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	// Calculate pagination info
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &SearchResponse{
		Products:   products,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// fallbackDatabaseSuggestions provides basic suggestions using database
func (s *Service) fallbackDatabaseSuggestions(query string, size int) ([]string, error) {
	if size <= 0 {
		size = 5
	}

	var products []models.Product
	searchTerm := query + "%"

	if err := s.db.Select("name").
		Where("is_active = ? AND LOWER(name) LIKE ?", true, searchTerm).
		Limit(size).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch suggestions: %w", err)
	}

	var suggestions []string
	for _, product := range products {
		suggestions = append(suggestions, product.Name)
	}

	return suggestions, nil
}
