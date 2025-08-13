import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { AdvancedSearch } from '../AdvancedSearch';
import { apiClient } from '../../../lib/api';

// Mock the API client
jest.mock('../../../lib/api', () => ({
  apiClient: {
    advancedSearchProducts: jest.fn(),
    getCategories: jest.fn(),
    getSearchSuggestions: jest.fn(),
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

const mockCategories = {
  categories: [
    {
      id: 'cat-1',
      name: 'Electronics',
      slug: 'electronics',
      isActive: true,
      sortOrder: 1,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z',
    },
    {
      id: 'cat-2',
      name: 'Clothing',
      slug: 'clothing',
      isActive: true,
      sortOrder: 2,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z',
    },
  ],
  total: 2,
};

const mockSearchResults = {
  products: [
    {
      id: '1',
      name: 'Advanced Search Product',
      description: 'This product matches the advanced search',
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
  facets: {
    categories: [
      { id: 'cat-1', name: 'Electronics', count: 5 },
      { id: 'cat-2', name: 'Clothing', count: 3 },
    ],
    priceRanges: [
      { range: '0-25', min: 0, max: 25, count: 2 },
      { range: '25-50', min: 25, max: 50, count: 3 },
      { range: '50-100', min: 50, max: 100, count: 4 },
    ],
  },
};

const mockSuggestions = {
  suggestions: ['Advanced Product', 'Advanced Search', 'Advanced Filter'],
};

describe('AdvancedSearch', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockApiClient.getCategories.mockResolvedValue(mockCategories);
    mockApiClient.advancedSearchProducts.mockResolvedValue(mockSearchResults);
    mockApiClient.getSearchSuggestions.mockResolvedValue(mockSuggestions);
  });

  it('renders search interface correctly', async () => {
    render(<AdvancedSearch />);
    
    expect(screen.getByText('Product Search')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Search products...')).toBeInTheDocument();
    expect(screen.getByText('Filters')).toBeInTheDocument();
    expect(screen.getByText('Clear All')).toBeInTheDocument();
    
    // Wait for categories to load
    await waitFor(() => {
      expect(screen.getByText('All Categories')).toBeInTheDocument();
    });
  });

  it('loads and displays categories', async () => {
    render(<AdvancedSearch />);
    
    await waitFor(() => {
      expect(mockApiClient.getCategories).toHaveBeenCalled();
      expect(screen.getByText('Electronics')).toBeInTheDocument();
      expect(screen.getByText('Clothing')).toBeInTheDocument();
    });
  });

  it('performs search when query is entered', async () => {
    render(<AdvancedSearch />);
    
    const searchInput = screen.getByPlaceholderText('Search products...');
    fireEvent.change(searchInput, { target: { value: 'advanced' } });
    
    await waitFor(() => {
      expect(mockApiClient.advancedSearchProducts).toHaveBeenCalledWith(
        expect.objectContaining({
          query: 'advanced',
          page: 1,
          pageSize: 20,
          includeFacets: true,
        })
      );
    });
    
    await waitFor(() => {
      expect(screen.getByText('Found 1 products')).toBeInTheDocument();
      expect(screen.getByText('Advanced Search Product')).toBeInTheDocument();
    });
  });

  it('shows search suggestions when typing', async () => {
    render(<AdvancedSearch />);
    
    const searchInput = screen.getByPlaceholderText('Search products...');
    fireEvent.change(searchInput, { target: { value: 'adv' } });
    fireEvent.focus(searchInput);
    
    await waitFor(() => {
      expect(mockApiClient.getSearchSuggestions).toHaveBeenCalledWith('adv', 5);
    });
    
    await waitFor(() => {
      expect(screen.getByText('Advanced Product')).toBeInTheDocument();
      expect(screen.getByText('Advanced Search')).toBeInTheDocument();
    });
  });

  it('applies category filter', async () => {
    render(<AdvancedSearch />);
    
    await waitFor(() => {
      expect(screen.getByText('Electronics')).toBeInTheDocument();
    });
    
    const categorySelect = screen.getByDisplayValue('All Categories');
    fireEvent.change(categorySelect, { target: { value: 'cat-1' } });
    
    await waitFor(() => {
      expect(mockApiClient.advancedSearchProducts).toHaveBeenCalledWith(
        expect.objectContaining({
          categoryId: 'cat-1',
        })
      );
    });
  });

  it('applies price range filters', async () => {
    render(<AdvancedSearch />);
    
    const minPriceInput = screen.getByPlaceholderText('Min');
    const maxPriceInput = screen.getByPlaceholderText('Max');
    
    fireEvent.change(minPriceInput, { target: { value: '10' } });
    fireEvent.change(maxPriceInput, { target: { value: '100' } });
    
    await waitFor(() => {
      expect(mockApiClient.advancedSearchProducts).toHaveBeenCalledWith(
        expect.objectContaining({
          minPrice: 10,
          maxPrice: 100,
        })
      );
    });
  });

  it('applies in stock filter', async () => {
    render(<AdvancedSearch />);
    
    const inStockCheckbox = screen.getByLabelText('In Stock Only');
    fireEvent.click(inStockCheckbox);
    
    await waitFor(() => {
      expect(mockApiClient.advancedSearchProducts).toHaveBeenCalledWith(
        expect.objectContaining({
          inStock: true,
        })
      );
    });
  });

  it('applies sorting options', async () => {
    render(<AdvancedSearch />);
    
    const sortSelect = screen.getByDisplayValue('Newest First');
    fireEvent.change(sortSelect, { target: { value: 'price-asc' } });
    
    await waitFor(() => {
      expect(mockApiClient.advancedSearchProducts).toHaveBeenCalledWith(
        expect.objectContaining({
          sortBy: 'price',
          sortOrder: 'asc',
        })
      );
    });
  });

  it('displays facets when available', async () => {
    render(<AdvancedSearch initialQuery="test" />);
    
    await waitFor(() => {
      expect(screen.getByText('Categories')).toBeInTheDocument();
      expect(screen.getByText('Price Ranges')).toBeInTheDocument();
      // Check for facet-specific Electronics (with count)
      expect(screen.getByText('(5)')).toBeInTheDocument();
      expect(screen.getByText('$0-25')).toBeInTheDocument();
    });
  });

  it('clears all filters when clear button is clicked', async () => {
    render(<AdvancedSearch initialQuery="test" initialCategoryId="cat-1" />);
    
    const clearButton = screen.getByText('Clear All');
    fireEvent.click(clearButton);
    
    await waitFor(() => {
      expect(screen.getByPlaceholderText('Search products...')).toHaveValue('');
      expect(screen.getByDisplayValue('All Categories')).toBeInTheDocument();
    });
  });

  it('calls onAddToCart when product add to cart is clicked', async () => {
    const mockOnAddToCart = jest.fn();
    render(<AdvancedSearch onAddToCart={mockOnAddToCart} initialQuery="test" />);
    
    await waitFor(() => {
      expect(screen.getByText('Advanced Search Product')).toBeInTheDocument();
    });
    
    const addToCartButton = screen.getByText('Add to Cart');
    fireEvent.click(addToCartButton);
    
    expect(mockOnAddToCart).toHaveBeenCalledWith('1');
  });

  it('handles search errors gracefully', async () => {
    mockApiClient.advancedSearchProducts.mockRejectedValue(new Error('Search failed'));
    
    render(<AdvancedSearch initialQuery="test" />);
    
    await waitFor(() => {
      expect(screen.getByText('Failed to search products')).toBeInTheDocument();
    });
  });
});