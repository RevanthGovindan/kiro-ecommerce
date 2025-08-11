import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { Header } from '../Header';
import { CartProvider } from '../../../contexts/CartContext';

// Mock Next.js Link component
jest.mock('next/link', () => {
  const MockLink = ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  );
  MockLink.displayName = 'MockLink';
  return MockLink;
});

// Mock the API client
jest.mock('../../../lib/api', () => ({
  apiClient: {
    getCart: jest.fn().mockResolvedValue({
      sessionId: 'test-session',
      items: [],
      subtotal: 0,
      tax: 0,
      total: 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }),
  },
}));

const renderWithCartProvider = (component: React.ReactElement) => {
  return render(
    <CartProvider>
      {component}
    </CartProvider>
  );
};

describe('Header', () => {
  it('renders header with logo and navigation', () => {
    renderWithCartProvider(<Header />);
    
    expect(screen.getByText('Ecommerce Store')).toBeInTheDocument();
    expect(screen.getByText('Products')).toBeInTheDocument();
    expect(screen.getByText('Cart')).toBeInTheDocument();
    expect(screen.getByText('Account')).toBeInTheDocument();
  });

  it('renders search input', () => {
    renderWithCartProvider(<Header />);
    
    const searchInput = screen.getByPlaceholderText('Search products...');
    expect(searchInput).toBeInTheDocument();
    expect(screen.getByText('Search')).toBeInTheDocument();
  });

  it('calls onSearch when search form is submitted', () => {
    const mockOnSearch = jest.fn();
    renderWithCartProvider(<Header onSearch={mockOnSearch} />);
    
    const searchInput = screen.getByPlaceholderText('Search products...');
    const searchButton = screen.getByText('Search');
    
    fireEvent.change(searchInput, { target: { value: 'test query' } });
    fireEvent.click(searchButton);
    
    expect(mockOnSearch).toHaveBeenCalledWith('test query');
  });

  it('calls onSearch when search form is submitted with Enter key', () => {
    const mockOnSearch = jest.fn();
    renderWithCartProvider(<Header onSearch={mockOnSearch} />);
    
    const searchInput = screen.getByPlaceholderText('Search products...');
    
    fireEvent.change(searchInput, { target: { value: 'test query' } });
    fireEvent.submit(searchInput.closest('form')!);
    
    expect(mockOnSearch).toHaveBeenCalledWith('test query');
  });

  it('does not call onSearch with empty query', () => {
    const mockOnSearch = jest.fn();
    renderWithCartProvider(<Header onSearch={mockOnSearch} />);
    
    const searchButton = screen.getByText('Search');
    fireEvent.click(searchButton);
    
    expect(mockOnSearch).not.toHaveBeenCalled();
  });

  it('trims whitespace from search query', () => {
    const mockOnSearch = jest.fn();
    renderWithCartProvider(<Header onSearch={mockOnSearch} />);
    
    const searchInput = screen.getByPlaceholderText('Search products...');
    const searchButton = screen.getByText('Search');
    
    fireEvent.change(searchInput, { target: { value: '  test query  ' } });
    fireEvent.click(searchButton);
    
    expect(mockOnSearch).toHaveBeenCalledWith('test query');
  });

  it('has correct navigation links', () => {
    renderWithCartProvider(<Header />);
    
    const logoLink = screen.getByRole('link', { name: /ecommerce store/i });
    expect(logoLink).toHaveAttribute('href', '/');
    
    const productsLink = screen.getByRole('link', { name: /products/i });
    expect(productsLink).toHaveAttribute('href', '/products');
    
    const cartLink = screen.getByRole('link', { name: /cart/i });
    expect(cartLink).toHaveAttribute('href', '/cart');
    
    const accountLink = screen.getByRole('link', { name: /account/i });
    expect(accountLink).toHaveAttribute('href', '/account');
  });

  it('renders cart link with test id', () => {
    renderWithCartProvider(<Header />);
    
    const cartLink = screen.getByTestId('cart-link');
    expect(cartLink).toBeInTheDocument();
    expect(cartLink).toHaveTextContent('Cart');
  });
});