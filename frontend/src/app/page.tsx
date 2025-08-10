'use client';

import React from 'react';
import { Header } from '@/components/layout/Header';
import { HeroSection } from '@/components/home/HeroSection';
import { CategoriesSection } from '@/components/home/CategoriesSection';
import { FeaturedProducts } from '@/components/home/FeaturedProducts';

export default function Home() {
  const handleSearch = (query: string) => {
    // Redirect to products page with search query
    window.location.href = `/products?search=${encodeURIComponent(query)}`;
  };

  const handleAddToCart = (productId: string) => {
    // TODO: Implement add to cart functionality
    console.log('Add to cart:', productId);
    alert('Add to cart functionality will be implemented in a future task');
  };

  return (
    <div className="min-h-screen bg-white">
      <Header onSearch={handleSearch} />
      
      <main>
        <HeroSection />
        <CategoriesSection />
        <FeaturedProducts onAddToCart={handleAddToCart} />
      </main>
    </div>
  );
}