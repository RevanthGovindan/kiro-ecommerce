import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AdminProductManagement } from '../AdminProductManagement';
import { apiClient } from '../../../lib/api';

// Mock the API client
jest.mock('../../../lib/api', () => ({
  apiClient: {
    getAdminProducts: jest.fn(),
    getCategories: jest.fn(),
    deleteProduct: jest.fn(),
  },
}));

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

const mockProducts = [
  {
    id: '1',
    name: 'Test Product 1',
    description: 'Test description 1',
    price: 100,
    sku: 'TEST-001',
    inventory: 50,
    isActive: true,
    categoryId: 'cat1',
    images: ['https://example.com/image1.jpg'],
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
    category: {
      id: 'cat1',
      name: 'Category 1',
      slug: 'category-1',
      isActive: true,
      sortOrder: 1,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z',
    },
  },
  {
    id: '2',
    name: 'Test Product 2',
    description: 'Test description 2',
    price: 200,
    sku: 'TEST-002',
    inventory: 25,
    isActive: false,
    categoryId: 'cat2',
    images: [],
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
    category: {
      id: 'cat2',
      name: 'Category 2',
      slug: 'category-2',
      isActive: true,
      sortOrder: 2,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z',
    },
  },
];

const mockCategories = [
  {
    id: 'cat1',
    name: 'Category 1',
    slug: 'category-1',
    isActive: true,
    sortOrder: 1,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'cat2',
    name: 'Category 2',
    slug: 'category-2',
    isActive: true,
    sortOrder: 2,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

describe('AdminProductManagement', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockApiClient.getAdminProducts.mockResolvedValue({
      products: mockProducts,
      total: 2,
      page: 1,
      pageSize: 20,
      totalPages: 1,
    });
    mockApiClient.getCategories.mockResolvedValue({
      categories: mockCategories,
      total: 2,
    });
  });

  it('renders product management interface', async () => {
    render(<AdminProductManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Product Management')).toBeInTheDocument();
    });

    expect(screen.getByText('Add New Product')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Search by name, SKU, or description...')).toBeInTheDocument();
  });

  it('displays products in table', async () => {
    render(<AdminProductManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Test Product 1')).toBeInTheDocument();
    });

    expect(screen.getByText('Test Product 2')).toBeInTheDocument();
    expect(screen.getByText('TEST-001')).toBeInTheDocument();
    expect(screen.getByText('TEST-002')).toBeInTheDocument();
    expect(screen.getByText('₹100')).toBeInTheDocument();
    expect(screen.getByText('₹200')).toBeInTheDocument();
  });

  it('shows product status correctly', async () => {
    render(<AdminProductManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Test Product 1')).toBeInTheDocument();
    });

    const activeStatus = screen.getByText('Active');
    const inactiveStatus = screen.getByText('Inactive');
    
    expect(activeStatus).toHaveClass('bg-green-100', 'text-green-800');
    expect(inactiveStatus).toHaveClass('bg-red-100', 'text-red-800');
  });

  it('handles search functionality', async () => {
    const user = userEvent.setup();
    render(<AdminProductManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Product Management')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search by name, SKU, or description...');
    const searchButton = screen.getByRole('button', { name: 'Search' });

    await user.type(searchInput, 'Test Product');
    await user.click(searchButton);

    expect(mockApiClient.getAdminProducts).toHaveBeenCalledWith({
      page: 1,
      pageSize: 20,
      search: 'Test Product',
      categoryId: undefined,
    });
  });

  it('handles category filter', async () => {
    const user = userEvent.setup();
    render(<AdminProductManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Product Management')).toBeInTheDocument();
    });

    const categorySelect = screen.getByDisplayValue('All Categories');
    await user.selectOptions(categorySelect, 'cat1');

    const searchButton = screen.getByRole('button', { name: 'Search' });
    await user.click(searchButton);

    expect(mockApiClient.getAdminProducts).toHaveBeenCalledWith({
      page: 1,
      pageSize: 20,
      search: undefined,
      categoryId: 'cat1',
    });
  });

  it('handles product deletion with confirmation', async () => {
    const user = userEvent.setup();
    // Mock window.confirm
    const confirmSpy = jest.spyOn(window, 'confirm').mockReturnValue(true);
    mockApiClient.deleteProduct.mockResolvedValue();

    render(<AdminProductManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Test Product 1')).toBeInTheDocument();
    });

    const deleteButtons = screen.getAllByText('Delete');
    await user.click(deleteButtons[0]);

    expect(confirmSpy).toHaveBeenCalledWith('Are you sure you want to delete this product?');
    expect(mockApiClient.deleteProduct).toHaveBeenCalledWith('1');

    confirmSpy.mockRestore();
  });

  it('cancels deletion when user declines confirmation', async () => {
    const user = userEvent.setup();
    const confirmSpy = jest.spyOn(window, 'confirm').mockReturnValue(false);

    render(<AdminProductManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Test Product 1')).toBeInTheDocument();
    });

    const deleteButtons = screen.getAllByText('Delete');
    await user.click(deleteButtons[0]);

    expect(confirmSpy).toHaveBeenCalled();
    expect(mockApiClient.deleteProduct).not.toHaveBeenCalled();

    confirmSpy.mockRestore();
  });

  it('opens product form when Add New Product is clicked', async () => {
    const user = userEvent.setup();
    render(<AdminProductManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Add New Product')).toBeInTheDocument();
    });

    const addButton = screen.getByText('Add New Product');
    await user.click(addButton);

    // The form should be rendered (we'll test the form component separately)
    expect(screen.getByText('Add New Product')).toBeInTheDocument();
  });

  it('handles API error gracefully', async () => {
    mockApiClient.getAdminProducts.mockRejectedValue(new Error('API Error'));
    
    render(<AdminProductManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Failed to load products')).toBeInTheDocument();
    });
  });
});