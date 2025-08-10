package products

import (
	"net/http"
	"strconv"

	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetProducts handles GET /api/products
func (h *Handler) GetProducts(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	pagination := PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	// Parse filters
	filters := ProductFilters{}

	if categoryID := c.Query("category_id"); categoryID != "" {
		filters.CategoryID = &categoryID
	}

	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filters.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filters.MaxPrice = &maxPrice
		}
	}

	if inStockStr := c.Query("in_stock"); inStockStr != "" {
		if inStock, err := strconv.ParseBool(inStockStr); err == nil {
			filters.InStock = &inStock
		}
	}

	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	// Parse sorting
	sort := ProductSort{
		Field: c.DefaultQuery("sort_by", "created_at"),
		Order: c.DefaultQuery("sort_order", "desc"),
	}

	// Get products
	response, err := h.service.GetProducts(filters, sort, pagination)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_PRODUCTS_ERROR", "Failed to fetch products", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Products retrieved successfully", response)
}

// GetProductByID handles GET /api/products/:id
func (h *Handler) GetProductByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Product ID is required", nil)
		return
	}

	product, err := h.service.GetProductByID(id)
	if err != nil {
		if err.Error() == "product not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_PRODUCT_ERROR", "Failed to fetch product", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product retrieved successfully", product)
}

// SearchProducts handles GET /api/products/search
func (h *Handler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_SEARCH_QUERY", "Search query is required", nil)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	pagination := PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	response, err := h.service.SearchProducts(query, pagination)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "SEARCH_PRODUCTS_ERROR", "Failed to search products", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Search completed successfully", response)
}

// GetCategories handles GET /api/categories
func (h *Handler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_CATEGORIES_ERROR", "Failed to fetch categories", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Categories retrieved successfully", gin.H{
		"categories": categories,
		"total":      len(categories),
	})
}

// GetCategoryByID handles GET /api/categories/:id
func (h *Handler) GetCategoryByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_CATEGORY_ID", "Category ID is required", nil)
		return
	}

	category, err := h.service.GetCategoryByID(id)
	if err != nil {
		if err.Error() == "category not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_CATEGORY_ERROR", "Failed to fetch category", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Category retrieved successfully", category)
}

// Admin Product Management Handlers

// CreateProduct handles POST /api/admin/products
func (h *Handler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request data", err.Error())
		return
	}

	// Validate required fields
	if req.Name == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_NAME", "Product name is required", nil)
		return
	}
	if req.SKU == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_SKU", "Product SKU is required", nil)
		return
	}
	if req.Price <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PRICE", "Product price must be greater than 0", nil)
		return
	}
	if req.CategoryID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_CATEGORY", "Category ID is required", nil)
		return
	}

	product, err := h.service.CreateProduct(req)
	if err != nil {
		if err.Error() == "category not found" {
			utils.ErrorResponse(c, http.StatusBadRequest, "CATEGORY_NOT_FOUND", "Category not found", nil)
			return
		}
		if err.Error() == "sku already exists" {
			utils.ErrorResponse(c, http.StatusConflict, "SKU_EXISTS", "Product with this SKU already exists", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "CREATE_PRODUCT_ERROR", "Failed to create product", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Product created successfully", product)
}

// UpdateProduct handles PUT /api/admin/products/:id
func (h *Handler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Product ID is required", nil)
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request data", err.Error())
		return
	}

	// Validate price if provided
	if req.Price != nil && *req.Price <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PRICE", "Product price must be greater than 0", nil)
		return
	}

	product, err := h.service.UpdateProduct(id, req)
	if err != nil {
		if err.Error() == "product not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", nil)
			return
		}
		if err.Error() == "category not found" {
			utils.ErrorResponse(c, http.StatusBadRequest, "CATEGORY_NOT_FOUND", "Category not found", nil)
			return
		}
		if err.Error() == "sku already exists" {
			utils.ErrorResponse(c, http.StatusConflict, "SKU_EXISTS", "Product with this SKU already exists", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_PRODUCT_ERROR", "Failed to update product", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product updated successfully", product)
}

// DeleteProduct handles DELETE /api/admin/products/:id (soft delete)
func (h *Handler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Product ID is required", nil)
		return
	}

	err := h.service.DeleteProduct(id)
	if err != nil {
		if err.Error() == "product not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_PRODUCT_ERROR", "Failed to delete product", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product deleted successfully", nil)
}

// UpdateInventory handles PUT /api/admin/products/:id/inventory
func (h *Handler) UpdateInventory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Product ID is required", nil)
		return
	}

	var req UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request data", err.Error())
		return
	}

	if req.Inventory < 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_INVENTORY", "Inventory cannot be negative", nil)
		return
	}

	product, err := h.service.UpdateInventory(id, req.Inventory)
	if err != nil {
		if err.Error() == "product not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_INVENTORY_ERROR", "Failed to update inventory", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Inventory updated successfully", product)
}

// GetAllProductsAdmin handles GET /api/admin/products (includes inactive products)
func (h *Handler) GetAllProductsAdmin(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	pagination := PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	// Parse filters (similar to regular GetProducts but includes inactive)
	filters := AdminProductFilters{}

	if categoryID := c.Query("category_id"); categoryID != "" {
		filters.CategoryID = &categoryID
	}

	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filters.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filters.MaxPrice = &maxPrice
		}
	}

	if inStockStr := c.Query("in_stock"); inStockStr != "" {
		if inStock, err := strconv.ParseBool(inStockStr); err == nil {
			filters.InStock = &inStock
		}
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filters.IsActive = &isActive
		}
	}

	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	// Parse sorting
	sort := ProductSort{
		Field: c.DefaultQuery("sort_by", "created_at"),
		Order: c.DefaultQuery("sort_order", "desc"),
	}

	// Get products
	response, err := h.service.GetAllProductsAdmin(filters, sort, pagination)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_PRODUCTS_ERROR", "Failed to fetch products", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Products retrieved successfully", response)
}
