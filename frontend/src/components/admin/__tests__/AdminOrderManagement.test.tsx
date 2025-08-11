import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AdminOrderManagement } from '../AdminOrderManagement';
import { apiClient } from '../../../lib/api';

// Mock the API client
jest.mock('../../../lib/api', () => ({
  apiClient: {
    getAdminOrders: jest.fn(),
    updateOrderStatus: jest.fn(),
  },
}));

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

const mockOrders = [
  {
    id: 'order-1',
    userId: 'user-1',
    status: 'pending',
    subtotal: 100,
    tax: 10,
    shipping: 5,
    total: 115,
    shippingAddress: {
      firstName: 'John',
      lastName: 'Doe',
      company: '',
      address1: '123 Main St',
      address2: '',
      city: 'Mumbai',
      state: 'Maharashtra',
      postalCode: '400001',
      country: 'India',
      phone: '+91-9876543210',
    },
    billingAddress: {
      firstName: 'John',
      lastName: 'Doe',
      company: '',
      address1: '123 Main St',
      address2: '',
      city: 'Mumbai',
      state: 'Maharashtra',
      postalCode: '400001',
      country: 'India',
      phone: '+91-9876543210',
    },
    paymentIntentId: 'pi_test_123',
    notes: '',
    createdAt: '2024-01-01T10:00:00Z',
    updatedAt: '2024-01-01T10:00:00Z',
  },
  {
    id: 'order-2',
    userId: 'user-2',
    status: 'shipped',
    subtotal: 200,
    tax: 20,
    shipping: 10,
    total: 230,
    shippingAddress: {
      firstName: 'Jane',
      lastName: 'Smith',
      company: '',
      address1: '456 Oak Ave',
      address2: '',
      city: 'Delhi',
      state: 'Delhi',
      postalCode: '110001',
      country: 'India',
      phone: '+91-9876543211',
    },
    billingAddress: {
      firstName: 'Jane',
      lastName: 'Smith',
      company: '',
      address1: '456 Oak Ave',
      address2: '',
      city: 'Delhi',
      state: 'Delhi',
      postalCode: '110001',
      country: 'India',
      phone: '+91-9876543211',
    },
    paymentIntentId: 'pi_test_456',
    notes: '',
    createdAt: '2024-01-02T15:30:00Z',
    updatedAt: '2024-01-02T15:30:00Z',
  },
];

describe('AdminOrderManagement', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockApiClient.getAdminOrders.mockResolvedValue({
      orders: mockOrders,
      total: 2,
      page: 1,
      pageSize: 20,
      totalPages: 1,
    });
  });

  it('renders order management interface', async () => {
    render(<AdminOrderManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Order Management')).toBeInTheDocument();
    });

    expect(screen.getByText('Total Orders: 2')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Search by order ID, customer email, or phone...')).toBeInTheDocument();
  });

  it('displays orders in table', async () => {
    render(<AdminOrderManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('#order-1')).toBeInTheDocument();
    });

    expect(screen.getByText('#order-2')).toBeInTheDocument();
    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('Jane Smith')).toBeInTheDocument();
    expect(screen.getByText('₹115')).toBeInTheDocument();
    expect(screen.getByText('₹230')).toBeInTheDocument();
  });

  it('shows order status with correct styling', async () => {
    render(<AdminOrderManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Order Management')).toBeInTheDocument();
    });

    const pendingStatus = screen.getByText('Pending');
    const shippedStatus = screen.getByText('Shipped');
    
    expect(pendingStatus).toHaveClass('bg-yellow-100', 'text-yellow-800');
    expect(shippedStatus).toHaveClass('bg-indigo-100', 'text-indigo-800');
  });

  it('handles search functionality', async () => {
    const user = userEvent.setup();
    render(<AdminOrderManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Order Management')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search by order ID, customer email, or phone...');
    const searchButton = screen.getByRole('button', { name: 'Search' });

    await user.type(searchInput, 'John');
    await user.click(searchButton);

    expect(mockApiClient.getAdminOrders).toHaveBeenCalledWith({
      page: 1,
      pageSize: 20,
      search: 'John',
      status: undefined,
    });
  });

  it('handles status filter', async () => {
    const user = userEvent.setup();
    render(<AdminOrderManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Order Management')).toBeInTheDocument();
    });

    const statusSelect = screen.getByDisplayValue('All Statuses');
    await user.selectOptions(statusSelect, 'pending');

    const searchButton = screen.getByRole('button', { name: 'Search' });
    await user.click(searchButton);

    expect(mockApiClient.getAdminOrders).toHaveBeenCalledWith({
      page: 1,
      pageSize: 20,
      search: undefined,
      status: 'pending',
    });
  });

  it('handles order status update with confirmation', async () => {
    const user = userEvent.setup();
    const confirmSpy = jest.spyOn(window, 'confirm').mockReturnValue(true);
    mockApiClient.updateOrderStatus.mockResolvedValue({
      ...mockOrders[0],
      status: 'confirmed',
    });

    render(<AdminOrderManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Order Management')).toBeInTheDocument();
    });

    const statusSelects = screen.getAllByRole('combobox');
    // Find the select that has 'pending' as selected value
    const pendingSelect = statusSelects.find(select => 
      (select as HTMLSelectElement).value === 'pending'
    );
    
    if (pendingSelect) {
      await user.selectOptions(pendingSelect, 'confirmed');
    }

    expect(confirmSpy).toHaveBeenCalledWith('Are you sure you want to update this order status to "confirmed"?');
    expect(mockApiClient.updateOrderStatus).toHaveBeenCalledWith('order-1', 'confirmed');

    confirmSpy.mockRestore();
  });

  it('cancels status update when user declines confirmation', async () => {
    const user = userEvent.setup();
    const confirmSpy = jest.spyOn(window, 'confirm').mockReturnValue(false);

    render(<AdminOrderManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Order Management')).toBeInTheDocument();
    });

    const statusSelects = screen.getAllByRole('combobox');
    // Find the select that has 'pending' as selected value
    const pendingSelect = statusSelects.find(select => 
      (select as HTMLSelectElement).value === 'pending'
    );
    
    if (pendingSelect) {
      await user.selectOptions(pendingSelect, 'confirmed');
    }

    expect(confirmSpy).toHaveBeenCalled();
    expect(mockApiClient.updateOrderStatus).not.toHaveBeenCalled();

    confirmSpy.mockRestore();
  });

  it('formats dates correctly', async () => {
    render(<AdminOrderManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Order Management')).toBeInTheDocument();
    });

    // Check if dates are formatted (exact format may vary based on locale)
    expect(screen.getByText(/1 Jan 2024/)).toBeInTheDocument();
    expect(screen.getByText(/2 Jan 2024/)).toBeInTheDocument();
  });

  it('shows empty state when no orders found', async () => {
    mockApiClient.getAdminOrders.mockResolvedValue({
      orders: [],
      total: 0,
      page: 1,
      pageSize: 20,
      totalPages: 0,
    });

    render(<AdminOrderManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('No orders found')).toBeInTheDocument();
    });

    expect(screen.getByText('Orders will appear here once customers start placing them')).toBeInTheDocument();
  });

  it('handles API error gracefully', async () => {
    mockApiClient.getAdminOrders.mockRejectedValue(new Error('API Error'));
    
    render(<AdminOrderManagement />);
    
    await waitFor(() => {
      expect(screen.getByText('Failed to load orders')).toBeInTheDocument();
    });
  });
});