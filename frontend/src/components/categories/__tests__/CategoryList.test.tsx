import React from 'react';
import { render, screen } from '@testing-library/react';
import { CategoryList } from '../CategoryList';
import { Category } from '@/lib/api';

// Mock Next.js Link component
jest.mock('next/link', () => {
  const MockLink = ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  );
  MockLink.displayName = 'MockLink';
  return MockLink;
});

const mockCategories: Category[] = [
  {
    id: '1',
    name: 'Electronics',
    slug: 'electronics',
    isActive: true,
    sortOrder: 1,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    name: 'Clothing',
    slug: 'clothing',
    isActive: true,
    sortOrder: 2,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

describe('CategoryList', () => {
  it('renders all categories', () => {
    render(<CategoryList categories={mockCategories} />);
    
    expect(screen.getByText('Electronics')).toBeInTheDocument();
    expect(screen.getByText('Clothing')).toBeInTheDocument();
  });

  it('shows loading skeleton when loading is true', () => {
    render(<CategoryList categories={[]} loading={true} />);
    
    const skeletonElements = document.querySelectorAll('.animate-pulse');
    expect(skeletonElements.length).toBeGreaterThan(0);
  });

  it('shows empty state when no categories are provided', () => {
    render(<CategoryList categories={[]} />);
    
    expect(screen.getByText('No categories available')).toBeInTheDocument();
  });

  it('creates correct links for categories', () => {
    render(<CategoryList categories={mockCategories} />);
    
    const electronicsLink = screen.getByRole('link', { name: /electronics/i });
    expect(electronicsLink).toHaveAttribute('href', '/products?category=1');
    
    const clothingLink = screen.getByRole('link', { name: /clothing/i });
    expect(clothingLink).toHaveAttribute('href', '/products?category=2');
  });

  it('renders categories in a grid layout', () => {
    render(<CategoryList categories={mockCategories} />);
    
    const gridContainer = document.querySelector('.grid');
    expect(gridContainer).toBeInTheDocument();
    expect(gridContainer).toHaveClass('grid-cols-2', 'md:grid-cols-4', 'lg:grid-cols-6');
  });
});