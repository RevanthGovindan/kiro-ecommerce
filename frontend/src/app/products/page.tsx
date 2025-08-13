'use client';

import React, { useEffect, useState, useCallback, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { Product, Category, apiClient } from '@/lib/api';

import { ProductGrid } from '@/components/products/ProductGrid';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

function ProductsPageContent() {
  const searchParams = useSearchParams();
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  
  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  
  // Filter state
  const [filters, setFilters] = useState({
    search: searchParams?.get('search') || '',
    categoryId: searchParams?.get('category') || '',
    minPrice: searchParams?.get('minPrice') || '',
    maxPrice: searchParams?.get('maxPrice') || '',
    sortBy: (searchParams?.get('sortBy') as 'name' | 'price' | 'created_at') || 'created_at',
    sortOrder: (searchParams?.get('sortOrder') as 'asc' | 'desc') || 'desc',
    inStock: searchParams?.get('inStock') === 'true',
  });

  const fetchProducts = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      
      const params = {
        page: currentPage,
        pageSize: 20,
        ...filters,
        minPrice: filters.minPrice ? parseFloat(filters.minPrice) : undefined,
        maxPrice: filters.maxPrice ? parseFloat(filters.maxPrice) : undefined,
      };

      const response = await apiClient.getProducts(params);
      setProducts(response.products);
      setTotalPages(response.totalPages);
      setTotal(response.total);
    } catch (err) {
      console.error('Error fetching products:', err);
      setError('Failed to load products');
    } finally {
      setLoading(false);
    }
  }, [currentPage, filters]);

  const fetchCategories = useCallback(async () => {
    try {
      const response = await apiClient.getCategories();
      setCategories(response.categories.filter(cat => cat.isActive));
    } catch (err) {
      console.error('Error fetching categories:', err);
    }
  }, []);

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  useEffect(() => {
    fetchCategories();
  }, [fetchCategories]);

  const handleSearch = (query: string) => {
    setFilters(prev => ({ ...prev, search: query }));
    setCurrentPage(1);
  };

  const handleFilterChange = (key: string, value: string | boolean) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setCurrentPage(1);
  };

  const handleAddToCart = (productId: string) => {
    console.log('Add to cart:', productId);
    alert('Add to cart functionality will be implemented in a future task');
  };

  const clearFilters = () => {
    setFilters({
      search: '',
      categoryId: '',
      minPrice: '',
      maxPrice: '',
      sortBy: 'created_at',
      sortOrder: 'desc',
      inStock: false,
    });
    setCurrentPage(1);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-4">Products</h1>
          
          {/* Filters */}
          <div className="bg-white p-6 rounded-lg shadow-sm mb-6">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
              {/* Category Filter */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Category
                </label>
                <select
                  value={filters.categoryId}
                  onChange={(e) => handleFilterChange('categoryId', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">All Categories</option>
                  {categories.map((category) => (
                    <option key={category.id} value={category.id}>
                      {category.name}
                    </option>
                  ))}
                </select>
              </div>

              {/* Price Range */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Min Price
                </label>
                <Input
                  type="number"
                  placeholder="0"
                  value={filters.minPrice}
                  onChange={(e) => handleFilterChange('minPrice', e.target.value)}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Max Price
                </label>
                <Input
                  type="number"
                  placeholder="1000"
                  value={filters.maxPrice}
                  onChange={(e) => handleFilterChange('maxPrice', e.target.value)}
                />
              </div>

              {/* Sort */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Sort By
                </label>
                <select
                  value={`${filters.sortBy}-${filters.sortOrder}`}
                  onChange={(e) => {
                    const [sortBy, sortOrder] = e.target.value.split('-');
                    handleFilterChange('sortBy', sortBy);
                    handleFilterChange('sortOrder', sortOrder);
                  }}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="created_at-desc">Newest First</option>
                  <option value="created_at-asc">Oldest First</option>
                  <option value="name-asc">Name A-Z</option>
                  <option value="name-desc">Name Z-A</option>
                  <option value="price-asc">Price Low to High</option>
                  <option value="price-desc">Price High to Low</option>
                </select>
              </div>
            </div>

            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={filters.inStock}
                    onChange={(e) => handleFilterChange('inStock', e.target.checked)}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700">In Stock Only</span>
                </label>
              </div>

              <div className="flex items-center space-x-4">
                <Button variant="outline" size="sm" onClick={clearFilters}>
                  Clear Filters
                </Button>
                
                {/* View Mode Toggle */}
                <div className="flex border border-gray-300 rounded-md">
                  <button
                    onClick={() => setViewMode('grid')}
                    className={`px-3 py-1 text-sm ${
                      viewMode === 'grid' 
                        ? 'bg-blue-600 text-white' 
                        : 'bg-white text-gray-700 hover:bg-gray-50'
                    }`}
                  >
                    Grid
                  </button>
                  <button
                    onClick={() => setViewMode('list')}
                    className={`px-3 py-1 text-sm ${
                      viewMode === 'list' 
                        ? 'bg-blue-600 text-white' 
                        : 'bg-white text-gray-700 hover:bg-gray-50'
                    }`}
                  >
                    List
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* Results Summary */}
          <div className="flex justify-between items-center mb-6">
            <p className="text-gray-600">
              Showing {products.length} of {total} products
              {filters.search && ` for "${filters.search}"`}
            </p>
          </div>
        </div>

        {/* Error State */}
        {error && (
          <div className="text-center py-8">
            <p className="text-red-600 mb-4">{error}</p>
            <Button onClick={() => fetchProducts()}>Try Again</Button>
          </div>
        )}

        {/* Products Grid */}
        {!error && (
          <ProductGrid 
            products={products} 
            onAddToCart={handleAddToCart}
            loading={loading}
            viewMode={viewMode}
          />
        )}

        {/* Pagination */}
        {!loading && !error && totalPages > 1 && (
          <div className="flex justify-center items-center space-x-2 mt-8">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
              disabled={currentPage === 1}
            >
              Previous
            </Button>
            
            {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
              const page = i + 1;
              return (
                <Button
                  key={page}
                  variant={currentPage === page ? 'primary' : 'outline'}
                  size="sm"
                  onClick={() => setCurrentPage(page)}
                >
                  {page}
                </Button>
              );
            })}
            
            <Button
              variant="outline"
              size="sm"
              onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
              disabled={currentPage === totalPages}
            >
              Next
            </Button>
          </div>
        )}
      </main>
    </div>
  );
}

export default function ProductsPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <ProductsPageContent />
    </Suspense>
  );
}