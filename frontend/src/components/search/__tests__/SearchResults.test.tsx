import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { SearchResults } from '../SearchResults';
import { apiClient } from '../../../lib/api';

// Mock the API client
jest.mock('../../../lib/api', () => ({
  apiClient: {
    searchProducts: jest.fn(),
    advancedSearchProducts: jest.fn(),
  },
}));

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

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

const mockSearchResults = {
  products: [
    {
      id: '1',
      name: 'Search Result Product',
      description: 'This product matches the search query',
      price: 99.99,
      sku: 'SEARCH-001',
      inventory: 10,
      isActive: true,
      categoryId: 'cat-1',
      images: ['/search-product.jpg'],
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z',
    },
  ],
  total: 1,
  page: 1,
  pageSize: 20,
  totalPages: 1,
};

const mockAdvancedSearchResults = {
  products: [
    {
      id: '1',
      name: 'Advanced Search Product',
      description: 'This product matches the advanced search query',
      price: 99.99,
      sku: 'ADV-001',
      inventory: 10,
      isActive: true,
      categoryId: 'cat-1',
      images: ['/advanced-product.jpg'],
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z',
    },
  ],
  total: 1,
  page: 1,
  pageSize: 20,
  totalPages: 1,
  suggestions: ['Advanced Product', 'Advanced Search'],
};

describe('SearchResults', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders nothing when query is empty', () => {
    const { container } = render(<SearchResults query="" />);
    expect(container.firstChild).toBeNull();
  });

  it('displays search results when query is provided', async () => {
    mockApiClient.searchProducts.mockResolvedValue(mockSearchResults);
    
    render(<SearchResults query="test query" />);
    
    expect(screen.getByText('Search Results for "test query"')).toBeInTheDocument();
    
    await waitFor(() => {
      expect(screen.getByText('Found 1 products')).toBeInTheDocument();
      expect(screen.getByText('Search Result Product')).toBeInTheDocument();
    });
  });

  it('displays error message when search fails', async () => {
    mockApiClient.searchProducts.mockRejectedValue(new Error('Search failed'));
    
    render(<SearchResults query="test query" />);
    
    await waitFor(() => {
      expect(screen.getByText('Failed to search products')).toBeInTheDocument();
    });
  });

  it('calls onAddToCart when product card add to cart is clicked', async () => {
    mockApiClient.searchProducts.mockResolvedValue(mockSearchResults);
    const mockOnAddToCart = jest.fn();
    
    render(<SearchResults query="test query" onAddToCart={mockOnAddToCart} />);
    
    await waitFor(() => {
      expect(screen.getByText('Search Result Product')).toBeInTheDocument();
    });
    
    const addToCartButton = screen.getByText('Add to Cart');
    fireEvent.click(addToCartButton);
    
    expect(mockOnAddToCart).toHaveBeenCalledWith('1');
  });

  it('debounces search requests', async () => {
    mockApiClient.searchProducts.mockResolvedValue(mockSearchResults);
    
    const { rerender } = render(<SearchResults query="test" />);
    
    // Change query quickly
    rerender(<SearchResults query="test query" />);
    rerender(<SearchResults query="test query updated" />);
    
    // Wait for debounce
    await waitFor(() => {
      expect(mockApiClient.searchProducts).toHaveBeenCalledTimes(1);
      expect(mockApiClient.searchProducts).toHaveBeenCalledWith('test query updated', 1, 20);
    });
  });

  it('uses advanced search when enabled', async () => {
    mockApiClient.advancedSearchProducts.mockResolvedValue(mockAdvancedSearchResults);
    
    render(<SearchResults query="test query" useAdvancedSearch={true} />);
    
    await waitFor(() => {
      expect(mockApiClient.advancedSearchProducts).toHaveBeenCalledWith({
        query: 'test query',
        page: 1,
        pageSize: 20,
        includeFacets: false,
      });
      expect(screen.getByText('Advanced Search Product')).toBeInTheDocument();
    });
  });

  it('displays suggestions when no results found', async () => {
    const noResultsResponse = {
      ...mockAdvancedSearchResults,
      products: [],
      total: 0,
      suggestions: ['Did you mean this?', 'Or maybe this?'],
    };
    
    mockApiClient.advancedSearchProducts.mockResolvedValue(noResultsResponse);
    
    render(<SearchResults query="test query" useAdvancedSearch={true} />);
    
    await waitFor(() => {
      expect(screen.getByText('Did you mean:')).toBeInTheDocument();
      expect(screen.getByText('Did you mean this?')).toBeInTheDocument();
      expect(screen.getByText('Or maybe this?')).toBeInTheDocument();
    });
  });
});