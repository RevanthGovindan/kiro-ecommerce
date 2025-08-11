'use client';

import React, { useContext } from 'react';
import { useSearchParams } from 'next/navigation';
import { AdvancedSearch } from '@/components/search/AdvancedSearch';
import { CartContext } from '@/contexts/CartContext';

export default function SearchPage() {
  const searchParams = useSearchParams();
  const { addToCart } = useContext(CartContext);
  
  const initialQuery = searchParams.get('q') || '';
  const initialCategoryId = searchParams.get('category') || '';

  const handleAddToCart = async (productId: string) => {
    try {
      await addToCart(productId, 1);
    } catch (error) {
      console.error('Error adding to cart:', error);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <AdvancedSearch
        onAddToCart={handleAddToCart}
        initialQuery={initialQuery}
        initialCategoryId={initialCategoryId}
      />
    </div>
  );
}