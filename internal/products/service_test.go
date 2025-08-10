package products

import (
	"fmt"
	"testing"

	"ecommerce-website/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ProductServiceTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *Service
}

func (suite *ProductServiceTestSuite) SetupSuite() {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.Category{}, &models.Product{}, &models.User{}, &models.Order{}, &models.OrderItem{})
	suite.Require().NoError(err)

	suite.db = db
	suite.service = NewService(db)
}

func (suite *ProductServiceTestSuite) SetupTest() {
	// Clean up tables before each test
	suite.db.Exec("DELETE FROM order_items")
	suite.db.Exec("DELETE FROM orders")
	suite.db.Exec("DELETE FROM products")
	suite.db.Exec("DELETE FROM categories")
	suite.db.Exec("DELETE FROM users")
}

func (suite *ProductServiceTestSuite) createTestCategory() *models.Category {
	category := &models.Category{
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
	}
	suite.db.Create(category)
	return category
}

func (suite *ProductServiceTestSuite) createTestProduct(categoryID string, name string, price float64) *models.Product {
	product := &models.Product{
		Name:       name,
		Price:      price,
		SKU:        "SKU-" + name,
		Inventory:  10,
		IsActive:   true,
		CategoryID: categoryID,
		Images:     models.StringArray{"image1.jpg", "image2.jpg"},
	}
	suite.db.Create(product)
	return product
}

func (suite *ProductServiceTestSuite) TestGetProducts_Success() {
	// Setup test data
	category := suite.createTestCategory()
	suite.createTestProduct(category.ID, "Laptop", 999.99)
	suite.createTestProduct(category.ID, "Phone", 599.99)

	// Test getting products
	filters := ProductFilters{}
	sort := ProductSort{}
	pagination := PaginationParams{Page: 1, PageSize: 10}

	result, err := suite.service.GetProducts(filters, sort, pagination)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(2), result.Total)
	assert.Len(suite.T(), result.Products, 2)
	assert.Equal(suite.T(), 1, result.TotalPages)
	assert.False(suite.T(), result.HasNext)
	assert.False(suite.T(), result.HasPrevious)
}

func (suite *ProductServiceTestSuite) TestGetProducts_WithCategoryFilter() {
	// Setup test data
	category1 := suite.createTestCategory()
	category2 := &models.Category{
		Name:     "Books",
		Slug:     "books",
		IsActive: true,
	}
	suite.db.Create(category2)

	suite.createTestProduct(category1.ID, "Laptop", 999.99)
	suite.createTestProduct(category2.ID, "Book", 19.99)

	// Test filtering by category
	filters := ProductFilters{CategoryID: &category1.ID}
	sort := ProductSort{}
	pagination := PaginationParams{Page: 1, PageSize: 10}

	result, err := suite.service.GetProducts(filters, sort, pagination)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Products, 1)
	assert.Equal(suite.T(), "Laptop", result.Products[0].Name)
}

func (suite *ProductServiceTestSuite) TestGetProducts_WithPriceFilter() {
	// Setup test data
	category := suite.createTestCategory()
	suite.createTestProduct(category.ID, "Expensive", 1000.00)
	suite.createTestProduct(category.ID, "Cheap", 50.00)
	suite.createTestProduct(category.ID, "Medium", 500.00)

	// Test price range filter
	minPrice := 100.0
	maxPrice := 800.0
	filters := ProductFilters{MinPrice: &minPrice, MaxPrice: &maxPrice}
	sort := ProductSort{}
	pagination := PaginationParams{Page: 1, PageSize: 10}

	result, err := suite.service.GetProducts(filters, sort, pagination)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Products, 1)
	assert.Equal(suite.T(), "Medium", result.Products[0].Name)
}

func (suite *ProductServiceTestSuite) TestGetProducts_WithSearch() {
	// Setup test data
	category := suite.createTestCategory()
	laptop := suite.createTestProduct(category.ID, "Gaming Laptop", 999.99)
	laptop.Description = "High performance gaming laptop"
	suite.db.Save(laptop)

	suite.createTestProduct(category.ID, "Phone", 599.99)

	// Test search functionality
	search := "gaming"
	filters := ProductFilters{Search: &search}
	sort := ProductSort{}
	pagination := PaginationParams{Page: 1, PageSize: 10}

	result, err := suite.service.GetProducts(filters, sort, pagination)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Products, 1)
	assert.Equal(suite.T(), "Gaming Laptop", result.Products[0].Name)
}

func (suite *ProductServiceTestSuite) TestGetProducts_WithSorting() {
	// Setup test data
	category := suite.createTestCategory()
	suite.createTestProduct(category.ID, "B-Product", 200.00)
	suite.createTestProduct(category.ID, "A-Product", 100.00)
	suite.createTestProduct(category.ID, "C-Product", 300.00)

	// Test sorting by price ascending
	filters := ProductFilters{}
	sort := ProductSort{Field: "price", Order: "asc"}
	pagination := PaginationParams{Page: 1, PageSize: 10}

	result, err := suite.service.GetProducts(filters, sort, pagination)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(3), result.Total)
	assert.Len(suite.T(), result.Products, 3)
	assert.Equal(suite.T(), "A-Product", result.Products[0].Name)
	assert.Equal(suite.T(), "B-Product", result.Products[1].Name)
	assert.Equal(suite.T(), "C-Product", result.Products[2].Name)
}

func (suite *ProductServiceTestSuite) TestGetProducts_WithPagination() {
	// Setup test data
	category := suite.createTestCategory()
	for i := 1; i <= 5; i++ {
		suite.createTestProduct(category.ID, fmt.Sprintf("Product%d", i), float64(i*100))
	}

	// Test pagination - page 1
	filters := ProductFilters{}
	sort := ProductSort{Field: "name", Order: "asc"}
	pagination := PaginationParams{Page: 1, PageSize: 2}

	result, err := suite.service.GetProducts(filters, sort, pagination)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(5), result.Total)
	assert.Len(suite.T(), result.Products, 2)
	assert.Equal(suite.T(), 3, result.TotalPages)
	assert.True(suite.T(), result.HasNext)
	assert.False(suite.T(), result.HasPrevious)

	// Test pagination - page 2
	pagination.Page = 2
	result, err = suite.service.GetProducts(filters, sort, pagination)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result.Products, 2)
	assert.True(suite.T(), result.HasNext)
	assert.True(suite.T(), result.HasPrevious)
}

func (suite *ProductServiceTestSuite) TestGetProductByID_Success() {
	// Setup test data
	category := suite.createTestCategory()
	product := suite.createTestProduct(category.ID, "Test Product", 99.99)

	// Test getting product by ID
	result, err := suite.service.GetProductByID(product.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), product.ID, result.ID)
	assert.Equal(suite.T(), "Test Product", result.Name)
	assert.Equal(suite.T(), category.Name, result.Category.Name)
}

func (suite *ProductServiceTestSuite) TestGetProductByID_NotFound() {
	// Test getting non-existent product
	result, err := suite.service.GetProductByID("non-existent")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "product not found")
}

func (suite *ProductServiceTestSuite) TestGetProductByID_InactiveProduct() {
	// Setup test data
	category := suite.createTestCategory()
	product := suite.createTestProduct(category.ID, "Inactive Product", 99.99)
	product.IsActive = false
	suite.db.Save(product)

	// Test getting inactive product
	result, err := suite.service.GetProductByID(product.ID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "product not found")
}

func (suite *ProductServiceTestSuite) TestSearchProducts() {
	// Setup test data
	category := suite.createTestCategory()
	laptop := suite.createTestProduct(category.ID, "Gaming Laptop", 999.99)
	laptop.Description = "High performance gaming laptop"
	suite.db.Save(laptop)

	suite.createTestProduct(category.ID, "Phone", 599.99)

	// Test search
	pagination := PaginationParams{Page: 1, PageSize: 10}
	result, err := suite.service.SearchProducts("laptop", pagination)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Products, 1)
	assert.Equal(suite.T(), "Gaming Laptop", result.Products[0].Name)
}

func (suite *ProductServiceTestSuite) TestGetCategories_Success() {
	// Setup test data
	category1 := suite.createTestCategory()
	category2 := &models.Category{
		Name:      "Books",
		Slug:      "books",
		IsActive:  true,
		SortOrder: 1,
	}
	suite.db.Create(category2)

	// Test getting categories
	result, err := suite.service.GetCategories()

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	// Should be sorted by sort_order, then name
	assert.Equal(suite.T(), category1.Name, result[0].Name)
	assert.Equal(suite.T(), category2.Name, result[1].Name)
}

func (suite *ProductServiceTestSuite) TestGetCategoryByID_Success() {
	// Setup test data
	category := suite.createTestCategory()

	// Test getting category by ID
	result, err := suite.service.GetCategoryByID(category.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), category.ID, result.ID)
	assert.Equal(suite.T(), "Electronics", result.Name)
}

func (suite *ProductServiceTestSuite) TestGetCategoryByID_NotFound() {
	// Test getting non-existent category
	result, err := suite.service.GetCategoryByID("non-existent")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "category not found")
}

// Admin Product Management Service Tests

func (suite *ProductServiceTestSuite) TestCreateProduct_Success() {
	// Setup test data
	category := suite.createTestCategory()

	req := CreateProductRequest{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		SKU:         "TEST-SKU-001",
		Inventory:   50,
		CategoryID:  category.ID,
		Images:      []string{"image1.jpg", "image2.jpg"},
		Specifications: map[string]interface{}{
			"color": "red",
			"size":  "large",
		},
	}

	// Test creating product
	result, err := suite.service.CreateProduct(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Test Product", result.Name)
	assert.Equal(suite.T(), "TEST-SKU-001", result.SKU)
	assert.Equal(suite.T(), 99.99, result.Price)
	assert.Equal(suite.T(), 50, result.Inventory)
	assert.Equal(suite.T(), category.ID, result.CategoryID)
	assert.Equal(suite.T(), category.Name, result.Category.Name)
	assert.True(suite.T(), result.IsActive)
}

func (suite *ProductServiceTestSuite) TestCreateProduct_CategoryNotFound() {
	req := CreateProductRequest{
		Name:       "Test Product",
		Price:      99.99,
		SKU:        "TEST-SKU-001",
		CategoryID: "non-existent",
	}

	result, err := suite.service.CreateProduct(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "category not found")
}

func (suite *ProductServiceTestSuite) TestCreateProduct_DuplicateSKU() {
	category := suite.createTestCategory()
	suite.createTestProduct(category.ID, "Existing Product", 99.99)

	req := CreateProductRequest{
		Name:       "New Product",
		Price:      199.99,
		SKU:        "SKU-Existing Product", // Same SKU as existing product
		CategoryID: category.ID,
	}

	result, err := suite.service.CreateProduct(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "sku already exists")
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_Success() {
	category := suite.createTestCategory()
	product := suite.createTestProduct(category.ID, "Original Product", 99.99)

	updatePrice := 149.99
	updateName := "Updated Product"
	req := UpdateProductRequest{
		Name:  &updateName,
		Price: &updatePrice,
	}

	result, err := suite.service.UpdateProduct(product.ID, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Updated Product", result.Name)
	assert.Equal(suite.T(), 149.99, result.Price)
	assert.Equal(suite.T(), product.SKU, result.SKU) // SKU unchanged
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_NotFound() {
	updateName := "Updated Product"
	req := UpdateProductRequest{
		Name: &updateName,
	}

	result, err := suite.service.UpdateProduct("non-existent", req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "product not found")
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_DuplicateSKU() {
	category := suite.createTestCategory()
	product1 := suite.createTestProduct(category.ID, "Product1", 99.99)
	product2 := suite.createTestProduct(category.ID, "Product2", 199.99)

	// Try to update product2 with product1's SKU
	newSKU := product1.SKU
	req := UpdateProductRequest{
		SKU: &newSKU,
	}

	result, err := suite.service.UpdateProduct(product2.ID, req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "sku already exists")
}

func (suite *ProductServiceTestSuite) TestDeleteProduct_Success() {
	category := suite.createTestCategory()
	product := suite.createTestProduct(category.ID, "Product to Delete", 99.99)

	err := suite.service.DeleteProduct(product.ID)

	assert.NoError(suite.T(), err)

	// Verify product is soft deleted
	var deletedProduct models.Product
	err = suite.db.Unscoped().Where("id = ?", product.ID).First(&deletedProduct).Error
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), deletedProduct.DeletedAt)

	// Verify product is not found in regular queries
	_, err = suite.service.GetProductByID(product.ID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "product not found")
}

func (suite *ProductServiceTestSuite) TestDeleteProduct_NotFound() {
	err := suite.service.DeleteProduct("non-existent")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "product not found")
}

func (suite *ProductServiceTestSuite) TestUpdateInventory_Success() {
	category := suite.createTestCategory()
	product := suite.createTestProduct(category.ID, "Product", 99.99)

	result, err := suite.service.UpdateInventory(product.ID, 100)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 100, result.Inventory)
}

func (suite *ProductServiceTestSuite) TestUpdateInventory_NotFound() {
	result, err := suite.service.UpdateInventory("non-existent", 100)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "product not found")
}

func (suite *ProductServiceTestSuite) TestGetAllProductsAdmin_Success() {
	category := suite.createTestCategory()

	// Create active product
	suite.createTestProduct(category.ID, "Active Product", 99.99)

	// Create inactive product
	inactiveProduct := suite.createTestProduct(category.ID, "Inactive Product", 199.99)
	inactiveProduct.IsActive = false
	suite.db.Save(inactiveProduct)

	// Create soft deleted product
	deletedProduct := suite.createTestProduct(category.ID, "Deleted Product", 299.99)
	suite.db.Delete(deletedProduct)

	filters := AdminProductFilters{}
	sort := ProductSort{}
	pagination := PaginationParams{Page: 1, PageSize: 10}

	result, err := suite.service.GetAllProductsAdmin(filters, sort, pagination)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(3), result.Total) // Includes inactive and deleted
	assert.Len(suite.T(), result.Products, 3)
}

func (suite *ProductServiceTestSuite) TestGetAllProductsAdmin_WithActiveFilter() {
	category := suite.createTestCategory()

	// Create active product
	suite.createTestProduct(category.ID, "Active Product", 99.99)

	// Create inactive product
	inactiveProduct := suite.createTestProduct(category.ID, "Inactive Product", 199.99)
	inactiveProduct.IsActive = false
	suite.db.Save(inactiveProduct)

	isActive := true
	filters := AdminProductFilters{IsActive: &isActive}
	sort := ProductSort{}
	pagination := PaginationParams{Page: 1, PageSize: 10}

	result, err := suite.service.GetAllProductsAdmin(filters, sort, pagination)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Products, 1)
	assert.Equal(suite.T(), "Active Product", result.Products[0].Name)
}

func (suite *ProductServiceTestSuite) TestGetAllProductsAdmin_WithSearch() {
	category := suite.createTestCategory()

	// Create products with different names and SKUs
	product1 := suite.createTestProduct(category.ID, "Gaming Laptop", 999.99)
	product1.SKU = "GAMING-001"
	suite.db.Save(product1)

	suite.createTestProduct(category.ID, "Phone", 599.99)

	search := "gaming"
	filters := AdminProductFilters{Search: &search}
	sort := ProductSort{}
	pagination := PaginationParams{Page: 1, PageSize: 10}

	result, err := suite.service.GetAllProductsAdmin(filters, sort, pagination)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Products, 1)
	assert.Equal(suite.T(), "Gaming Laptop", result.Products[0].Name)
}

func TestProductServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProductServiceTestSuite))
}
