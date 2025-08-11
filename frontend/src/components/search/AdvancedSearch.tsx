'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { Product, Category, AdvancedSearchResponse, SearchFacets, apiClient } from '@/lib/api';
import { ProductGrid } from '../products/ProductGrid';
import { SearchSuggestions } from './SearchSuggestions';

interface AdvancedSearchProps {
  onAddToCart?: (productId: string) => void;
  initialQuery?: string;
  initialCategoryId?: string;
}

interface SearchFilters {
  query: string;
  categoryId: string;
  minPrice: string;
  maxPrice: string;
  inStock: boolean;
  sortBy: 'name' | 'price' | 'created_at' | 'popularity';
  sortOrder: 'asc' | 'desc';
}

export const AdvancedSearch: React.FC<AdvancedSearchProps> = ({
  onAddToCart,
  initialQuery = '',
  initialCategoryId = '',
}) => {
  const [filters, setFilters] = useState<SearchFilters>({
    query: initialQuery,
    categoryId: initialCategoryId,
    minPrice: '',
    maxPrice: '',
    inStock: false,
    sortBy: 'created_at',
    sortOrder: 'desc',
  });

  const [searchResults, setSearchResults] = useState<AdvancedSearchResponse | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [suggestions, setSuggestions] = useState<string[]>([]);

  const pageSize = 20;

  // Load categories on component mount
  useEffect(() => {
    const loadCategories = async () => {
      try {
        const response = await apiClient.getCategories();
        setCategories(response.categories);
      } catch (err) {
        console.error('Error loading categories:', err);
      }
    };

    loadCategories();
  }, []);

  // Debounced search function
  const debouncedSearch = useCallback(
    debounce(async (searchFilters: SearchFilters, page: number) => {
      try {
        setLoading(true);
        setError(null);

        const params = {
          query: searchFilters.query || undefined,
          categoryId: searchFilters.categoryId || undefined,
          minPrice: searchFilters.minPrice ? parseFloat(searchFilters.minPrice) : undefined,
          maxPrice: searchFilters.maxPrice ? parseFloat(searchFilters.maxPrice) : undefined,
          inStock: searchFilters.inStock || undefined,
          sortBy: searchFilters.sortBy,
          sortOrder: searchFilters.sortOrder,
          page,
          pageSize,
          includeFacets: true,
        };

        const response = await apiClient.advancedSearchProducts(params);
        setSearchResults(response);
      } catch (err) {
        console.error('Error performing advanced search:', err);
        setError('Failed to search products');
      } finally {
        setLoading(false);
      }
    }, 300),
    []
  );

  // Debounced suggestions function
  const debouncedSuggestions = useCallback(
    debounce(async (query: string) => {
      if (query.length >= 2) {
        try {
          const response = await apiClient.getSearchSuggestions(query, 5);
          setSuggestions(response.suggestions);
        } catch (err) {
          console.error('Error getting suggestions:', err);
          setSuggestions([]);
        }
      } else {
        setSuggestions([]);
      }
    }, 200),
    []
  );

  // Perform search when filters change
  useEffect(() => {
    debouncedSearch(filters, currentPage);
  }, [filters, currentPage, debouncedSearch]);

  // Get suggestions when query changes
  useEffect(() => {
    if (filters.query && showSuggestions) {
      debouncedSuggestions(filters.query);
    } else {
      setSuggestions([]);
    }
  }, [filters.query, showSuggestions, debouncedSuggestions]);

  const handleFilterChange = (key: keyof SearchFilters, value: string | boolean) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setCurrentPage(1); // Reset to first page when filters change
  };

  const handleQueryChange = (value: string) => {
    handleFilterChange('query', value);
  };

  const handleSuggestionSelect = (suggestion: string) => {
    setFilters(prev => ({ ...prev, query: suggestion }));
    setShowSuggestions(false);
    setSuggestions([]);
  };

  const handleClearFilters = () => {
    setFilters({
      query: '',
      categoryId: '',
      minPrice: '',
      maxPrice: '',
      inStock: false,
      sortBy: 'created_at',
      sortOrder: 'desc',
    });
    setCurrentPage(1);
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleFacetClick = (type: 'category' | 'priceRange', value: string) => {
    if (type === 'category') {
      handleFilterChange('categoryId', filters.categoryId === value ? '' : value);
    } else if (type === 'priceRange') {
      const [min, max] = value.split('-');
      if (max === '+') {
        handleFilterChange('minPrice', min);
        handleFilterChange('maxPrice', '');
      } else {
        handleFilterChange('minPrice', min);
        handleFilterChange('maxPrice', max);
      }
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Search Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">Product Search</h1>
        
        {/* Search Input */}
        <div className="relative mb-6">
          <input
            type="text"
            placeholder="Search products..."
            value={filters.query}
            onChange={(e) => handleQueryChange(e.target.value)}
            onFocus={() => setShowSuggestions(true)}
            onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
            className="w-full px-4 py-3 pl-12 pr-4 text-lg border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg className="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          
          {/* Search Suggestions */}
          {showSuggestions && suggestions.length > 0 && (
            <SearchSuggestions
              suggestions={suggestions}
              onSelect={handleSuggestionSelect}
            />
          )}
        </div>
      </div>

      <div className="flex flex-col lg:flex-row gap-8">
        {/* Filters Sidebar */}
        <div className="lg:w-1/4">
          <div className="bg-white rounded-lg shadow-md p-6 sticky top-4">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Filters</h2>
              <button
                onClick={handleClearFilters}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                Clear All
              </button>
            </div>

            {/* Category Filter */}
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Category
              </label>
              <select
                value={filters.categoryId}
                onChange={(e) => handleFilterChange('categoryId', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">All Categories</option>
                {categories.map((category) => (
                  <option key={category.id} value={category.id}>
                    {category.name}
                  </option>
                ))}
              </select>
            </div>

            {/* Price Range Filter */}
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Price Range
              </label>
              <div className="flex gap-2">
                <input
                  type="number"
                  placeholder="Min"
                  value={filters.minPrice}
                  onChange={(e) => handleFilterChange('minPrice', e.target.value)}
                  className="w-1/2 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
                <input
                  type="number"
                  placeholder="Max"
                  value={filters.maxPrice}
                  onChange={(e) => handleFilterChange('maxPrice', e.target.value)}
                  className="w-1/2 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>

            {/* In Stock Filter */}
            <div className="mb-6">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={filters.inStock}
                  onChange={(e) => handleFilterChange('inStock', e.target.checked)}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <span className="ml-2 text-sm text-gray-700">In Stock Only</span>
              </label>
            </div>

            {/* Sort Options */}
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Sort By
              </label>
              <select
                value={`${filters.sortBy}-${filters.sortOrder}`}
                onChange={(e) => {
                  const [sortBy, sortOrder] = e.target.value.split('-') as [typeof filters.sortBy, typeof filters.sortOrder];
                  handleFilterChange('sortBy', sortBy);
                  handleFilterChange('sortOrder', sortOrder);
                }}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="created_at-desc">Newest First</option>
                <option value="created_at-asc">Oldest First</option>
                <option value="name-asc">Name A-Z</option>
                <option value="name-desc">Name Z-A</option>
                <option value="price-asc">Price Low to High</option>
                <option value="price-desc">Price High to Low</option>
                <option value="popularity-desc">Most Popular</option>
              </select>
            </div>

            {/* Facets */}
            {searchResults?.facets && (
              <div>
                {/* Category Facets */}
                {searchResults.facets.categories.length > 0 && (
                  <div className="mb-6">
                    <h3 className="text-sm font-medium text-gray-700 mb-2">Categories</h3>
                    <div className="space-y-2">
                      {searchResults.facets.categories.map((facet) => (
                        <button
                          key={facet.id}
                          onClick={() => handleFacetClick('category', facet.id)}
                          className={`flex items-center justify-between w-full text-left px-2 py-1 rounded text-sm ${
                            filters.categoryId === facet.id
                              ? 'bg-blue-100 text-blue-800'
                              : 'text-gray-600 hover:bg-gray-100'
                          }`}
                        >
                          <span>{facet.name}</span>
                          <span className="text-xs text-gray-500">({facet.count})</span>
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                {/* Price Range Facets */}
                {searchResults.facets.priceRanges.length > 0 && (
                  <div className="mb-6">
                    <h3 className="text-sm font-medium text-gray-700 mb-2">Price Ranges</h3>
                    <div className="space-y-2">
                      {searchResults.facets.priceRanges.map((facet) => (
                        <button
                          key={facet.range}
                          onClick={() => handleFacetClick('priceRange', facet.range)}
                          className="flex items-center justify-between w-full text-left px-2 py-1 rounded text-sm text-gray-600 hover:bg-gray-100"
                        >
                          <span>${facet.range}</span>
                          <span className="text-xs text-gray-500">({facet.count})</span>
                        </button>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Results */}
        <div className="lg:w-3/4">
          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-6">
              <p className="text-red-600">{error}</p>
            </div>
          )}

          {searchResults && (
            <div className="mb-6">
              <div className="flex items-center justify-between">
                <p className="text-gray-600">
                  {loading ? 'Searching...' : `Found ${searchResults.total} products`}
                </p>
                {searchResults.total > 0 && (
                  <p className="text-sm text-gray-500">
                    Page {currentPage} of {searchResults.totalPages}
                  </p>
                )}
              </div>
            </div>
          )}

          <ProductGrid
            products={searchResults?.products || []}
            onAddToCart={onAddToCart}
            loading={loading}
          />

          {/* Pagination */}
          {searchResults && searchResults.totalPages > 1 && (
            <div className="mt-8 flex justify-center">
              <nav className="flex items-center space-x-2">
                <button
                  onClick={() => handlePageChange(currentPage - 1)}
                  disabled={currentPage === 1}
                  className="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Previous
                </button>
                
                {Array.from({ length: Math.min(5, searchResults.totalPages) }, (_, i) => {
                  const page = i + 1;
                  return (
                    <button
                      key={page}
                      onClick={() => handlePageChange(page)}
                      className={`px-3 py-2 text-sm font-medium rounded-md ${
                        currentPage === page
                          ? 'bg-blue-600 text-white'
                          : 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50'
                      }`}
                    >
                      {page}
                    </button>
                  );
                })}
                
                <button
                  onClick={() => handlePageChange(currentPage + 1)}
                  disabled={currentPage === searchResults.totalPages}
                  className="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </nav>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

// Debounce utility function
function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout;
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}