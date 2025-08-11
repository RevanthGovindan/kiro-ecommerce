'use client';

import React from 'react';
import { HeroSection } from '@/components/home/HeroSection';
import { CategoriesSection } from '@/components/home/CategoriesSection';
import { FeaturedProducts } from '@/components/home/FeaturedProducts';

export default function Home() {
  const handleAddToCart = (productId: string) => {
    // TODO: Implement add to cart functionality
    console.log('Add to cart:', productId);
    alert('Add to cart functionality will be implemented in a future task');
  };

  return (
    <div className="min-h-screen bg-white">
      <HeroSection />
      <CategoriesSection />
      <FeaturedProducts onAddToCart={handleAddToCart} />
    </div>
  );
}