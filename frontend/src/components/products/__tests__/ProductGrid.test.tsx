import React from 'react';
import { render, screen } from '@testing-library/react';
import { ProductGrid } from '../ProductGrid';
import { Product } from '@/lib/api';

// Mock Next.js components
jest.mock('next/link', () => {
  return ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  );
});

jest.mock('next/image', () => {
  return ({ src, alt, ...props }: any) => (
    <img src={src} alt={alt} {...props} />
  );
});

const mockProducts: Product[] = [
  {
    id: '1',
    name: 'Product 1',
    description: 'Description 1',
    price: 99.99,
    sku: 'PROD-001',
    inventory: 10,
    isActive: true,
    categoryId: 'cat-1',
    images: ['/product1.jpg'],
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    name: 'Product 2',
    description: 'Description 2',
    price: 149.99,
    sku: 'PROD-002',
    inventory: 5,
    isActive: true,
    categoryId: 'cat-1',
    images: ['/product2.jpg'],
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

describe('ProductGrid', () => {
  it('renders all products', () => {
    render(<ProductGrid products={mockProducts} />);
    
    expect(screen.getByText('Product 1')).toBeInTheDocument();
    expect(screen.getByText('Product 2')).toBeInTheDocument();
  });

  it('shows loading skeleton when loading is true', () => {
    render(<ProductGrid products={[]} loading={true} />);
    
    const skeletonElements = screen.getAllByTestId('loading-skeleton');
    expect(skeletonElements.length).toBeGreaterThan(0);
  });

  it('shows empty state when no products are provided', () => {
    render(<ProductGrid products={[]} />);
    
    expect(screen.getByText('No products found')).toBeInTheDocument();
    expect(screen.getByText('Try adjusting your search or filters')).toBeInTheDocument();
  });

  it('renders products in a grid layout', () => {
    render(<ProductGrid products={mockProducts} />);
    
    const gridContainer = screen.getByRole('grid', { hidden: true });
    expect(gridContainer).toHaveClass('grid');
  });
});