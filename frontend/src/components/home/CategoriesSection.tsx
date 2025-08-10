'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { Category, apiClient } from '@/lib/api';
import { CategoryList } from '../categories/CategoryList';

export const CategoriesSection: React.FC = () => {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        setLoading(true);
        setError(null);
        
        const response = await apiClient.getCategories();
        // Show only top-level categories (no parent) and limit to 6
        const topLevelCategories = response.categories
          .filter(cat => !cat.parentId && cat.isActive)
          .slice(0, 6);
        
        setCategories(topLevelCategories);
      } catch (err) {
        console.error('Error fetching categories:', err);
        setError('Failed to load categories');
      } finally {
        setLoading(false);
      }
    };

    fetchCategories();
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
    <section className="py-12 bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-8">
          <h2 className="text-3xl font-bold text-gray-900 mb-4">Shop by Category</h2>
          <p className="text-lg text-gray-600">Find exactly what you're looking for</p>
        </div>
        
        <CategoryList categories={categories} loading={loading} />
        
        {!loading && categories.length > 0 && (
          <div className="text-center mt-8">
            <Link 
              href="/products" 
              className="text-blue-600 hover:text-blue-800 font-medium underline"
            >
              View all categories &rarr;
            </Link>
          </div>
        )}
      </div>
    </section>
  );
};