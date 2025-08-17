package products

import "ecommerce-website/internal/models"

// CreateProductRequest represents the request body for creating a product
type CreateProductRequest struct {
	Name           string                 `json:"name" binding:"required"`
	Description    string                 `json:"description"`
	Price          float64                `json:"price" binding:"required"`
	CompareAtPrice *float64               `json:"compareAtPrice,omitempty"`
	SKU            string                 `json:"sku" binding:"required"`
	Inventory      int                    `json:"inventory"`
	CategoryID     string                 `json:"categoryId" binding:"required"`
	Images         []string               `json:"images"`
	Specifications map[string]interface{} `json:"specifications"`
	SEOTitle       *string                `json:"seoTitle,omitempty"`
	SEODescription *string                `json:"seoDescription,omitempty"`
	IsActive       *bool                  `json:"isActive,omitempty"`
}

// UpdateProductRequest represents the request body for updating a product
type UpdateProductRequest struct {
	Name           *string                `json:"name,omitempty"`
	Description    *string                `json:"description,omitempty"`
	Price          *float64               `json:"price,omitempty"`
	CompareAtPrice *float64               `json:"compareAtPrice,omitempty"`
	SKU            *string                `json:"sku,omitempty"`
	Inventory      *int                   `json:"inventory,omitempty"`
	CategoryID     *string                `json:"categoryId,omitempty"`
	Images         []string               `json:"images,omitempty"`
	Specifications map[string]interface{} `json:"specifications,omitempty"`
	SEOTitle       *string                `json:"seoTitle,omitempty"`
	SEODescription *string                `json:"seoDescription,omitempty"`
	IsActive       *bool                  `json:"isActive,omitempty"`
}

// UpdateInventoryRequest represents the request body for updating inventory
type UpdateInventoryRequest struct {
	Inventory int `json:"inventory" binding:"required"`
}

// AdminProductFilters represents filters for admin product queries (includes inactive products)
type AdminProductFilters struct {
	CategoryID *string
	MinPrice   *float64
	MaxPrice   *float64
	InStock    *bool
	IsActive   *bool
	Search     *string
}

// AdminProductListResponse represents paginated admin product response
type AdminProductListResponse struct {
	Products    []models.Product `json:"products"`
	Total       int64            `json:"total"`
	Page        int              `json:"page"`
	PageSize    int              `json:"pageSize"`
	TotalPages  int              `json:"totalPages"`
	HasNext     bool             `json:"hasNext"`
	HasPrevious bool             `json:"hasPrevious"`
}

// Advanced Search Types

// AdvancedSearchFilters represents filters for advanced product search
type AdvancedSearchFilters struct {
	CategoryID *string
	MinPrice   *float64
	MaxPrice   *float64
	InStock    *bool
	Search     *string
}

// AdvancedSearchSort represents sorting options for advanced search
type AdvancedSearchSort struct {
	Field string // name, price, created_at, popularity
	Order string // asc, desc
}

// AdvancedSearchResponse represents advanced search response with facets
type AdvancedSearchResponse struct {
	Products    []models.Product `json:"products"`
	Total       int64            `json:"total"`
	Page        int              `json:"page"`
	PageSize    int              `json:"pageSize"`
	TotalPages  int              `json:"totalPages"`
	Suggestions []string         `json:"suggestions,omitempty"`
	Facets      *SearchFacets    `json:"facets,omitempty"`
}

// SearchFacets represents search facets for filtering
type SearchFacets struct {
	Categories  []CategoryFacet   `json:"categories"`
	PriceRanges []PriceRangeFacet `json:"priceRanges"`
}

// CategoryFacet represents a category facet
type CategoryFacet struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// PriceRangeFacet represents a price range facet
type PriceRangeFacet struct {
	Range string  `json:"range"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Count int64   `json:"count"`
}

// CreateCategoryRequest represents the request body for creating a category
type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Description *string `json:"description,omitempty"`
	ParentID    *string `json:"parentId,omitempty"`
	IsActive    *bool   `json:"isActive,omitempty"`
	SortOrder   *int    `json:"sortOrder,omitempty"`
}
