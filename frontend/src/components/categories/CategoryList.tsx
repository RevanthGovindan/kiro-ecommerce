'use client';

import React from 'react';
import Link from 'next/link';
import { Category } from '@/lib/api';

interface CategoryListProps {
  categories: Category[];
  loading?: boolean;
}

export const CategoryList: React.FC<CategoryListProps> = ({ categories, loading = false }) => {
  if (loading) {
    return (
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        {Array.from({ length: 6 }).map((_, index) => (
          <div key={index} className="animate-pulse">
            <div className="h-20 bg-gray-200 rounded-lg mb-2"></div>
            <div className="h-4 bg-gray-200 rounded"></div>
          </div>
        ))}
      </div>
    );
  }

  if (categories.length === 0) {
    return (
      <div className="text-center py-8">
        <p className="text-gray-500">No categories available</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
      {categories.map((category) => (
        <Link
          key={category.id}
          href={`/products?category=${category.id}`}
          className="group text-center"
        >
          <div className="h-20 bg-gradient-to-br from-blue-100 to-blue-200 rounded-lg mb-2 flex items-center justify-center group-hover:from-blue-200 group-hover:to-blue-300 transition-colors">
            <span className="text-2xl">📦</span>
          </div>
          <h3 className="text-sm font-medium text-gray-900 group-hover:text-blue-600 transition-colors">
            {category.name}
          </h3>
        </Link>
      ))}
    </div>
  );
};