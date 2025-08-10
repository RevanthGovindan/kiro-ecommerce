'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { Input } from '../ui/Input';
import { Button } from '../ui/Button';

interface HeaderProps {
  onSearch?: (query: string) => void;
}

export const Header: React.FC<HeaderProps> = ({ onSearch }) => {
  const [searchQuery, setSearchQuery] = useState('');

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (onSearch && searchQuery.trim()) {
      onSearch(searchQuery.trim());
    }
  };

  return (
    <header className="bg-white shadow-sm border-b">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-4">
          {/* Logo */}
          <Link href="/" className="text-2xl font-bold text-gray-900 hover:text-blue-600">
            Ecommerce Store
          </Link>

          {/* Search Bar */}
          <div className="flex-1 max-w-lg mx-8">
            <form onSubmit={handleSearch} className="relative">
              <Input
                type="text"
                placeholder="Search products..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pr-12"
              />
              <Button
                type="submit"
                size="sm"
                className="absolute right-1 top-1 bottom-1 px-3"
              >
                Search
              </Button>
            </form>
          </div>

          {/* Navigation */}
          <nav className="flex items-center space-x-6">
            <Link href="/products" className="text-gray-600 hover:text-gray-900 font-medium">
              Products
            </Link>
            <Link href="/cart" className="text-gray-600 hover:text-gray-900 font-medium">
              Cart
            </Link>
            <Link href="/account" className="text-gray-600 hover:text-gray-900 font-medium">
              Account
            </Link>
          </nav>
        </div>
      </div>
    </header>
  );
};