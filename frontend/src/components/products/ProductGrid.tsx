'use client';

import React from 'react';
import { Product } from '@/lib/api';
import { ProductCard } from './ProductCard';

interface ProductGridProps {
  products: Product[];
  onAddToCart?: (productId: string) => void;
  loading?: boolean;
  viewMode?: 'grid' | 'list';
}

export const ProductGrid: React.FC<ProductGridProps> = ({ 
  products, 
  onAddToCart, 
  loading = false,
  viewMode = 'grid'
}) => {
  if (loading) {
    const gridClasses = viewMode === 'grid' 
      ? "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6"
      : "space-y-4";
    
    return (
      <div className={gridClasses}>
        {Array.from({ length: 8 }).map((_, index) => (
          <div key={index} className={`bg-white rounded-lg shadow-md overflow-hidden animate-pulse ${
            viewMode === 'list' ? 'flex' : ''
          }`} data-testid="loading-skeleton">
            <div className={`bg-gray-200 ${viewMode === 'list' ? 'w-48 h-32' : 'h-48'}`}></div>
            <div className="p-4 flex-1">
              <div className="h-4 bg-gray-200 rounded mb-2"></div>
              <div className="h-3 bg-gray-200 rounded mb-3"></div>
              <div className="flex justify-between items-center">
                <div className="h-6 bg-gray-200 rounded w-20"></div>
                <div className="h-8 bg-gray-200 rounded w-24"></div>
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (products.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-500 text-lg mb-2">No products found</div>
        <p className="text-gray-400">Try adjusting your search or filters</p>
      </div>
    );
  }

  const containerClasses = viewMode === 'grid' 
    ? "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6"
    : "space-y-4";

  return (
    <div className={containerClasses} role="grid">
      {products.map((product) => (
        <ProductCard
          key={product.id}
          product={product}
          onAddToCart={onAddToCart}
          viewMode={viewMode}
        />
      ))}
    </div>
  );
};