import React from 'react';
import { render, screen } from '@testing-library/react';
import { ProductGrid } from '../ProductGrid';
import { Product } from '@/lib/api';

// Mock Next.js components
jest.mock('next/link', () => {
  const MockLink = ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  );
  MockLink.displayName = 'MockLink';
  return MockLink;
});

jest.mock('next/image', () => {
  const MockImage = ({ src, alt, fill, ...props }: { src: string; alt: string; fill?: boolean; [key: string]: unknown }) => (
    // eslint-disable-next-line @next/next/no-img-element
    <img src={src} alt={alt} data-fill={fill} {...props} />
  );
  MockImage.displayName = 'MockImage';
  return MockImage;
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
  it('renders all products in grid view', () => {
    render(<ProductGrid products={mockProducts} viewMode="grid" />);
    
    expect(screen.getByText('Product 1')).toBeInTheDocument();
    expect(screen.getByText('Product 2')).toBeInTheDocument();
  });

  it('renders all products in list view', () => {
    render(<ProductGrid products={mockProducts} viewMode="list" />);
    
    expect(screen.getByText('Product 1')).toBeInTheDocument();
    expect(screen.getByText('Product 2')).toBeInTheDocument();
  });

  it('shows loading skeleton when loading is true in grid view', () => {
    render(<ProductGrid products={[]} loading={true} viewMode="grid" />);
    
    const skeletonElements = screen.getAllByTestId('loading-skeleton');
    expect(skeletonElements.length).toBeGreaterThan(0);
  });

  it('shows loading skeleton when loading is true in list view', () => {
    render(<ProductGrid products={[]} loading={true} viewMode="list" />);
    
    const skeletonElements = screen.getAllByTestId('loading-skeleton');
    expect(skeletonElements.length).toBeGreaterThan(0);
  });

  it('shows empty state when no products are provided', () => {
    render(<ProductGrid products={[]} />);
    
    expect(screen.getByText('No products found')).toBeInTheDocument();
    expect(screen.getByText('Try adjusting your search or filters')).toBeInTheDocument();
  });

  it('renders products in a grid layout when viewMode is grid', () => {
    render(<ProductGrid products={mockProducts} viewMode="grid" />);
    
    const gridContainer = screen.getByRole('grid', { hidden: true });
    expect(gridContainer).toHaveClass('grid');
  });

  it('renders products in a list layout when viewMode is list', () => {
    render(<ProductGrid products={mockProducts} viewMode="list" />);
    
    const listContainer = screen.getByRole('grid', { hidden: true });
    expect(listContainer).toHaveClass('space-y-4');
  });
});