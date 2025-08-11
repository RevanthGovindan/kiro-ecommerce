'use client';

import React, { useEffect, useState } from 'react';
import { Product, apiClient, AdvancedSearchResponse } from '@/lib/api';
import { ProductGrid } from '../products/ProductGrid';

interface SearchResultsProps {
  query: string;
  onAddToCart?: (productId: string) => void;
  useAdvancedSearch?: boolean;
}

export const SearchResults: React.FC<SearchResultsProps> = ({ 
  query, 
  onAddToCart,
  useAdvancedSearch = false 
}) => {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);
  const [suggestions, setSuggestions] = useState<string[]>([]);

  useEffect(() => {
    const searchProducts = async () => {
      if (!query.trim()) {
        setProducts([]);
        setTotal(0);
        setSuggestions([]);
        return;
      }

      try {
        setLoading(true);
        setError(null);
        
        if (useAdvancedSearch) {
          const response = await apiClient.advancedSearchProducts({
            query,
            page: 1,
            pageSize: 20,
            includeFacets: false,
          });
          setProducts(response.products);
          setTotal(response.total);
          setSuggestions(response.suggestions || []);
        } else {
          const response = await apiClient.searchProducts(query, 1, 20);
          setProducts(response.products);
          setTotal(response.total);
        }
      } catch (err) {
        console.error('Error searching products:', err);
        setError('Failed to search products');
      } finally {
        setLoading(false);
      }
    };

    const debounceTimer = setTimeout(searchProducts, 300);
    return () => clearTimeout(debounceTimer);
  }, [query, useAdvancedSearch]);

  if (!query.trim()) {
    return null;
  }

  if (error) {
    return (
      <div className="text-center py-8">
        <p className="text-red-600">{error}</p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Search Results for &quot;{query}&quot;
        </h2>
        <p className="text-gray-600">
          {loading ? 'Searching...' : `Found ${total} products`}
        </p>
        
        {/* Show suggestions if available and no results */}
        {suggestions.length > 0 && products.length === 0 && !loading && (
          <div className="mt-4">
            <p className="text-sm text-gray-600 mb-2">Did you mean:</p>
            <div className="flex flex-wrap gap-2">
              {suggestions.map((suggestion, index) => (
                <button
                  key={index}
                  className="px-3 py-1 text-sm bg-gray-100 text-gray-700 rounded-full hover:bg-gray-200"
                  onClick={() => window.location.href = `/products?search=${encodeURIComponent(suggestion)}`}
                >
                  {suggestion}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
      
      <ProductGrid 
        products={products} 
        onAddToCart={onAddToCart}
        loading={loading}
      />
    </div>
  );
};