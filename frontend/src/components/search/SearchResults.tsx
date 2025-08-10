'use client';

import React, { useEffect, useState } from 'react';
import { Product, apiClient } from '@/lib/api';
import { ProductGrid } from '../products/ProductGrid';

interface SearchResultsProps {
  query: string;
  onAddToCart?: (productId: string) => void;
}

export const SearchResults: React.FC<SearchResultsProps> = ({ query, onAddToCart }) => {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    const searchProducts = async () => {
      if (!query.trim()) {
        setProducts([]);
        setTotal(0);
        return;
      }

      try {
        setLoading(true);
        setError(null);
        
        const response = await apiClient.searchProducts(query, 1, 20);
        setProducts(response.products);
        setTotal(response.total);
      } catch (err) {
        console.error('Error searching products:', err);
        setError('Failed to search products');
      } finally {
        setLoading(false);
      }
    };

    const debounceTimer = setTimeout(searchProducts, 300);
    return () => clearTimeout(debounceTimer);
  }, [query]);

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
      </div>
      
      <ProductGrid 
        products={products} 
        onAddToCart={onAddToCart}
        loading={loading}
      />
    </div>
  );
};