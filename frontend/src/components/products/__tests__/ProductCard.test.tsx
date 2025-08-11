import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { ProductCard } from '../ProductCard';
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

const mockProduct: Product = {
  id: '1',
  name: 'Test Product',
  description: 'This is a test product description',
  price: 99.99,
  compareAtPrice: 129.99,
  sku: 'TEST-001',
  inventory: 10,
  isActive: true,
  categoryId: 'cat-1',
  images: ['/test-image.jpg'],
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
};

describe('ProductCard', () => {
  it('renders product information correctly in grid view', () => {
    render(<ProductCard product={mockProduct} viewMode="grid" />);
    
    expect(screen.getByText('Test Product')).toBeInTheDocument();
    expect(screen.getByText('This is a test product description')).toBeInTheDocument();
    expect(screen.getByText('$99.99')).toBeInTheDocument();
    expect(screen.getByText('$129.99')).toBeInTheDocument();
    expect(screen.getByText('Add to Cart')).toBeInTheDocument();
  });

  it('renders product information correctly in list view', () => {
    render(<ProductCard product={mockProduct} viewMode="list" />);
    
    expect(screen.getByText('Test Product')).toBeInTheDocument();
    expect(screen.getByText('This is a test product description')).toBeInTheDocument();
    expect(screen.getByText('$99.99')).toBeInTheDocument();
    expect(screen.getByText('$129.99')).toBeInTheDocument();
    expect(screen.getByText('Add to Cart')).toBeInTheDocument();
  });

  it('calls onAddToCart when add to cart button is clicked in grid view', () => {
    const mockOnAddToCart = jest.fn();
    render(<ProductCard product={mockProduct} onAddToCart={mockOnAddToCart} viewMode="grid" />);
    
    const addToCartButton = screen.getByText('Add to Cart');
    fireEvent.click(addToCartButton);
    
    expect(mockOnAddToCart).toHaveBeenCalledWith('1');
  });

  it('calls onAddToCart when add to cart button is clicked in list view', () => {
    const mockOnAddToCart = jest.fn();
    render(<ProductCard product={mockProduct} onAddToCart={mockOnAddToCart} viewMode="list" />);
    
    const addToCartButton = screen.getByText('Add to Cart');
    fireEvent.click(addToCartButton);
    
    expect(mockOnAddToCart).toHaveBeenCalledWith('1');
  });

  it('shows out of stock when inventory is 0 in grid view', () => {
    const outOfStockProduct = { ...mockProduct, inventory: 0 };
    render(<ProductCard product={outOfStockProduct} viewMode="grid" />);
    
    expect(screen.getAllByText('Out of Stock')).toHaveLength(2);
    expect(screen.getByRole('button', { name: /out of stock/i })).toBeDisabled();
  });

  it('shows out of stock when inventory is 0 in list view', () => {
    const outOfStockProduct = { ...mockProduct, inventory: 0 };
    render(<ProductCard product={outOfStockProduct} viewMode="list" />);
    
    expect(screen.getAllByText('Out of Stock')).toHaveLength(2);
    expect(screen.getByRole('button', { name: /out of stock/i })).toBeDisabled();
  });

  it('shows low stock warning when inventory is 5 or less in grid view', () => {
    const lowStockProduct = { ...mockProduct, inventory: 3 };
    render(<ProductCard product={lowStockProduct} viewMode="grid" />);
    
    expect(screen.getByText('Only 3 left in stock')).toBeInTheDocument();
  });

  it('shows low stock warning when inventory is 5 or less in list view', () => {
    const lowStockProduct = { ...mockProduct, inventory: 3 };
    render(<ProductCard product={lowStockProduct} viewMode="list" />);
    
    expect(screen.getByText('Only 3 left')).toBeInTheDocument();
  });

  it('renders product image with correct alt text in grid view', () => {
    render(<ProductCard product={mockProduct} viewMode="grid" />);
    
    const image = screen.getByAltText('Test Product');
    expect(image).toBeInTheDocument();
    expect(image).toHaveAttribute('src', '/test-image.jpg');
  });

  it('renders product image with correct alt text in list view', () => {
    render(<ProductCard product={mockProduct} viewMode="list" />);
    
    const image = screen.getByAltText('Test Product');
    expect(image).toBeInTheDocument();
    expect(image).toHaveAttribute('src', '/test-image.jpg');
  });

  it('links to product detail page in both views', () => {
    const { rerender } = render(<ProductCard product={mockProduct} viewMode="grid" />);
    
    let productLinks = screen.getAllByRole('link');
    let productNameLink = productLinks.find(link => 
      link.getAttribute('href') === '/products/1'
    );
    expect(productNameLink).toBeInTheDocument();

    rerender(<ProductCard product={mockProduct} viewMode="list" />);
    productLinks = screen.getAllByRole('link');
    productNameLink = productLinks.find(link => 
      link.getAttribute('href') === '/products/1'
    );
    expect(productNameLink).toBeInTheDocument();
  });
});