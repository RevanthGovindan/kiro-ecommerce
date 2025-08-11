'use client';

import React from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { Product } from '@/lib/api';
import { Button } from '../ui/Button';

interface ProductCardProps {
  product: Product;
  onAddToCart?: (productId: string) => void;
  viewMode?: 'grid' | 'list';
}

export const ProductCard: React.FC<ProductCardProps> = ({ product, onAddToCart, viewMode = 'grid' }) => {
  const handleAddToCart = () => {
    if (onAddToCart) {
      onAddToCart(product.id);
    }
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price);
  };

  const primaryImage = (product.images && product.images.length > 0 && product.images[0]) 
    ? product.images[0] 
    : '/placeholder-product.jpg';
  
  // Validate the image URL (works in both server and client environments)
  const isValidUrl = (url: string) => {
    if (!url || typeof url !== 'string') return false;
    
    try {
      // Handle relative URLs
      if (url.startsWith('/')) return true;
      
      // Handle absolute URLs
      new URL(url);
      return true;
    } catch {
      return false;
    }
  };
  
  const safeImageUrl = isValidUrl(primaryImage) ? primaryImage : '/placeholder-product.jpg';

  if (viewMode === 'list') {
    return (
      <div className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow flex">
        <Link href={`/products/${product.id}`}>
          <div className="relative w-48 h-32 bg-gray-200 flex-shrink-0">
            <Image
              src={safeImageUrl}
              alt={product.name}
              fill
              className="object-cover"
              sizes="192px"
            />
            {product.inventory === 0 && (
              <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center">
                <span className="text-white font-semibold text-sm">Out of Stock</span>
              </div>
            )}
          </div>
        </Link>
        
        <div className="p-4 flex-1 flex flex-col justify-between">
          <div>
            <Link href={`/products/${product.id}`}>
              <h3 className="text-lg font-semibold text-gray-900 mb-2 hover:text-blue-600 line-clamp-1">
                {product.name}
              </h3>
            </Link>
            
            <p className="text-gray-600 text-sm mb-3 line-clamp-2">
              {product.description}
            </p>
          </div>
          
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-2">
              <span className="text-xl font-bold text-gray-900">
                {formatPrice(product.price)}
              </span>
              {product.compareAtPrice && product.compareAtPrice > product.price && (
                <span className="text-sm text-gray-500 line-through">
                  {formatPrice(product.compareAtPrice)}
                </span>
              )}
            </div>
            
            <div className="flex flex-col items-end space-y-1">
              <Button
                size="sm"
                onClick={handleAddToCart}
                disabled={product.inventory === 0}
                className="min-w-[100px]"
              >
                {product.inventory === 0 ? 'Out of Stock' : 'Add to Cart'}
              </Button>
              
              {product.inventory > 0 && product.inventory <= 5 && (
                <p className="text-xs text-orange-600">
                  Only {product.inventory} left
                </p>
              )}
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div data-testid="product-card" className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow">
      <Link href={`/products/${product.id}`}>
        <div className="relative h-48 bg-gray-200">
          <Image
            src={safeImageUrl}
            alt={product.name}
            fill
            className="object-cover"
            sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
          />
          {product.inventory === 0 && (
            <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center">
              <span className="text-white font-semibold">Out of Stock</span>
            </div>
          )}
        </div>
      </Link>
      
      <div className="p-4">
        <Link href={`/products/${product.id}`}>
          <h3 className="text-lg font-semibold text-gray-900 mb-2 hover:text-blue-600 line-clamp-2">
            {product.name}
          </h3>
        </Link>
        
        <p className="text-gray-600 text-sm mb-3 line-clamp-2">
          {product.description}
        </p>
        
        <div className="flex justify-between items-center">
          <div className="flex items-center space-x-2">
            <span className="text-xl font-bold text-gray-900">
              {formatPrice(product.price)}
            </span>
            {product.compareAtPrice && product.compareAtPrice > product.price && (
              <span className="text-sm text-gray-500 line-through">
                {formatPrice(product.compareAtPrice)}
              </span>
            )}
          </div>
          
          <Button
            size="sm"
            onClick={handleAddToCart}
            disabled={product.inventory === 0}
            className="min-w-[100px]"
          >
            {product.inventory === 0 ? 'Out of Stock' : 'Add to Cart'}
          </Button>
        </div>
        
        {product.inventory > 0 && product.inventory <= 5 && (
          <p className="text-sm text-orange-600 mt-2">
            Only {product.inventory} left in stock
          </p>
        )}
      </div>
    </div>
  );
};