'use client';

import React, { useEffect, useState } from 'react';
import { Product, apiClient } from '@/lib/api';
import { ProductGrid } from '../products/ProductGrid';

interface FeaturedProductsProps {
  onAddToCart?: (productId: string) => void;
}

export const FeaturedProducts: React.FC<FeaturedProductsProps> = ({ onAddToCart }) => {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchFeaturedProducts = async () => {
      try {
        setLoading(true);
        setError(null);
        
        // Fetch first 8 products sorted by creation date (newest first)
        const response = await apiClient.getProducts({
          page: 1,
          pageSize: 8,
          sortBy: 'created_at',
          sortOrder: 'desc',
          inStock: true,
        });
        
        setProducts(response.products);
      } catch (err) {
        console.error('Error fetching featured products:', err);
        setError('Failed to load featured products');
      } finally {
        setLoading(false);
      }
    };

    fetchFeaturedProducts();
  }, []);

  if (error) {
    return (
      <div className="text-center py-8">
        <p className="text-red-600 mb-4">{error}</p>
        <button 
          onClick={() => window.location.reload()} 
          className="text-blue-600 hover:text-blue-800 underline"
        >
          Try again
        </button>
      </div>
    );
  }

  return (
    <section className="py-12">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-8">
          <h2 className="text-3xl font-bold text-gray-900 mb-4">Featured Products</h2>
          <p className="text-lg text-gray-600">Discover our latest and most popular items</p>
        </div>
        
        <ProductGrid 
          products={products} 
          onAddToCart={onAddToCart}
          loading={loading}
        />
      </div>
    </section>
  );
};