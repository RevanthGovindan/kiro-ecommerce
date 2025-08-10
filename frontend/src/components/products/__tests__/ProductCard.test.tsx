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
  const MockImage = ({ src, alt, ...props }: { src: string; alt: string; [key: string]: unknown }) => (
    // eslint-disable-next-line @next/next/no-img-element
    <img src={src} alt={alt} {...props} />
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
  it('renders product information correctly', () => {
    render(<ProductCard product={mockProduct} />);
    
    expect(screen.getByText('Test Product')).toBeInTheDocument();
    expect(screen.getByText('This is a test product description')).toBeInTheDocument();
    expect(screen.getByText('$99.99')).toBeInTheDocument();
    expect(screen.getByText('$129.99')).toBeInTheDocument();
    expect(screen.getByText('Add to Cart')).toBeInTheDocument();
  });

  it('calls onAddToCart when add to cart button is clicked', () => {
    const mockOnAddToCart = jest.fn();
    render(<ProductCard product={mockProduct} onAddToCart={mockOnAddToCart} />);
    
    const addToCartButton = screen.getByText('Add to Cart');
    fireEvent.click(addToCartButton);
    
    expect(mockOnAddToCart).toHaveBeenCalledWith('1');
  });

  it('shows out of stock when inventory is 0', () => {
    const outOfStockProduct = { ...mockProduct, inventory: 0 };
    render(<ProductCard product={outOfStockProduct} />);
    
    expect(screen.getAllByText('Out of Stock')).toHaveLength(2);
    expect(screen.getByRole('button', { name: /out of stock/i })).toBeDisabled();
  });

  it('shows low stock warning when inventory is 5 or less', () => {
    const lowStockProduct = { ...mockProduct, inventory: 3 };
    render(<ProductCard product={lowStockProduct} />);
    
    expect(screen.getByText('Only 3 left in stock')).toBeInTheDocument();
  });

  it('renders product image with correct alt text', () => {
    render(<ProductCard product={mockProduct} />);
    
    const image = screen.getByAltText('Test Product');
    expect(image).toBeInTheDocument();
    expect(image).toHaveAttribute('src', '/test-image.jpg');
  });

  it('links to product detail page', () => {
    render(<ProductCard product={mockProduct} />);
    
    const productLinks = screen.getAllByRole('link');
    const productNameLink = productLinks.find(link => 
      link.getAttribute('href') === '/products/1'
    );
    
    expect(productNameLink).toBeInTheDocument();
  });
});