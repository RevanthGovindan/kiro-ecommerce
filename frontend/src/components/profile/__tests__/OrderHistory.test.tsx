import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { OrderHistory } from '../OrderHistory';
import { useAuth } from '../../../contexts/AuthContext';
import { apiClient } from '../../../lib/api';

// Mock the dependencies
jest.mock('../../../contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}));

jest.mock('../../../lib/api', () => ({
  apiClient: {
    getUserOrders: jest.fn(),
  },
}));

const mockOrders = [
  {
    id: 'order-1',
    userId: 'user-1',
    status: 'delivered',
    subtotal: 100,
    tax: 10,
    shipping: 5,
    total: 115,
    shippingAddress: {
      firstName: 'John',
      lastName: 'Doe',
      address1: '123 Main St',
      city: 'New York',
      state: 'NY',
      postalCode: '10001',
      country: 'US',
    },
    billingAddress: {
      firstName: 'John',
      lastName: 'Doe',
      address1: '123 Main St',
      city: 'New York',
      state: 'NY',
      postalCode: '10001',
      country: 'US',
    },
    paymentIntentId: 'pi_test',
    createdAt: '2023-01-01T00:00:00Z',
    updatedAt: '2023-01-01T00:00:00Z',
    items: [
      {
        id: 'item-1',
        orderId: 'order-1',
        productId: 'product-1',
        quantity: 2,
        price: 50,
        total: 100,
        product: {
          id: 'product-1',
          name: 'Test Product',
          sku: 'TEST-001',
          price: 50,
          images: ['test-image.jpg'],
        },
      },
    ],
  },
  {
    id: 'order-2',
    userId: 'user-1',
    status: 'pending',
    subtotal: 200,
    tax: 20,
    shipping: 10,
    total: 230,
    shippingAddress: {
      firstName: 'John',
      lastName: 'Doe',
      address1: '123 Main St',
      city: 'New York',
      state: 'NY',
      postalCode: '10001',
      country: 'US',
    },
    billingAddress: {
      firstName: 'John',
      lastName: 'Doe',
      address1: '123 Main St',
      city: 'New York',
      state: 'NY',
      postalCode: '10001',
      country: 'US',
    },
    paymentIntentId: 'pi_test_2',
    createdAt: '2023-01-02T00:00:00Z',
    updatedAt: '2023-01-02T00:00:00Z',
    items: [
      {
        id: 'item-2',
        orderId: 'order-2',
        productId: 'product-2',
        quantity: 1,
        price: 200,
        total: 200,
        product: {
          id: 'product-2',
          name: 'Another Product',
          sku: 'TEST-002',
          price: 200,
          images: ['test-image-2.jpg'],
        },
      },
    ],
  },
];

beforeEach(() => {
  (useAuth as jest.Mock).mockReturnValue({
    isAuthenticated: true,
  });
  
  jest.clearAllMocks();
});

describe('OrderHistory', () => {
  it('shows login prompt when user is not authenticated', () => {
    (useAuth as jest.Mock).mockReturnValue({
      isAuthenticated: false,
    });
    
    render(<OrderHistory />);
    
    expect(screen.getByText('Please log in to view your order history.')).toBeInTheDocument();
    expect(screen.getByText('Login')).toBeInTheDocument();
  });

  it('shows loading state while fetching orders', () => {
    (apiClient.getUserOrders as jest.Mock).mockImplementation(() => 
      new Promise(() => {}) // Never resolves
    );
    
    render(<OrderHistory />);
    
    // Should show loading skeletons
    expect(screen.getAllByRole('generic')).toHaveLength(3); // 3 skeleton items
  });

  it('displays orders when loaded successfully', async () => {
    (apiClient.getUserOrders as jest.Mock).mockResolvedValue({
      orders: mockOrders,
      pagination: {
        page: 1,
        totalPages: 1,
        total: 2,
      },
    });
    
    render(<OrderHistory />);
    
    await waitFor(() => {
      expect(screen.getByText('Order History')).toBeInTheDocument();
      expect(screen.getByText('2 orders total')).toBeInTheDocument();
      expect(screen.getByText('Order #' + mockOrders[0].id.slice(-8).toUpperCase())).toBeInTheDocument();
      expect(screen.getByText('Order #' + mockOrders[1].id.slice(-8).toUpperCase())).toBeInTheDocument();
    });
  });

  it('shows empty state when no orders exist', async () => {
    (apiClient.getUserOrders as jest.Mock).mockResolvedValue({
      orders: [],
      pagination: {
        page: 1,
        totalPages: 1,
        total: 0,
      },
    });
    
    render(<OrderHistory />);
    
    await waitFor(() => {
      expect(screen.getByText('No orders yet')).toBeInTheDocument();
      expect(screen.getByText('You haven\'t placed any orders yet.')).toBeInTheDocument();
      expect(screen.getByText('Start Shopping')).toBeInTheDocument();
    });
  });

  it('displays order status with correct styling', async () => {
    (apiClient.getUserOrders as jest.Mock).mockResolvedValue({
      orders: mockOrders,
      pagination: {
        page: 1,
        totalPages: 1,
        total: 2,
      },
    });
    
    render(<OrderHistory />);
    
    await waitFor(() => {
      expect(screen.getByText('Delivered')).toBeInTheDocument();
      expect(screen.getByText('Pending')).toBeInTheDocument();
    });
  });

  it('displays order totals correctly', async () => {
    (apiClient.getUserOrders as jest.Mock).mockResolvedValue({
      orders: mockOrders,
      pagination: {
        page: 1,
        totalPages: 1,
        total: 2,
      },
    });
    
    render(<OrderHistory />);
    
    await waitFor(() => {
      expect(screen.getByText('$115.00')).toBeInTheDocument();
      expect(screen.getByText('$230.00')).toBeInTheDocument();
    });
  });

  it('displays item counts correctly', async () => {
    (apiClient.getUserOrders as jest.Mock).mockResolvedValue({
      orders: mockOrders,
      pagination: {
        page: 1,
        totalPages: 1,
        total: 2,
      },
    });
    
    render(<OrderHistory />);
    
    await waitFor(() => {
      expect(screen.getByText('1 item: Test Product (×2)')).toBeInTheDocument();
      expect(screen.getByText('1 item: Another Product (×1)')).toBeInTheDocument();
    });
  });

  it('shows view details links for each order', async () => {
    (apiClient.getUserOrders as jest.Mock).mockResolvedValue({
      orders: mockOrders,
      pagination: {
        page: 1,
        totalPages: 1,
        total: 2,
      },
    });
    
    render(<OrderHistory />);
    
    await waitFor(() => {
      const viewDetailsButtons = screen.getAllByText('View Details');
      expect(viewDetailsButtons).toHaveLength(2);
    });
  });

  it('handles API error gracefully', async () => {
    (apiClient.getUserOrders as jest.Mock).mockRejectedValue(new Error('API Error'));
    
    render(<OrderHistory />);
    
    await waitFor(() => {
      expect(screen.getByText('Failed to load order history. Please try again.')).toBeInTheDocument();
      expect(screen.getByText('Try Again')).toBeInTheDocument();
    });
  });

  it('retries loading orders when try again is clicked', async () => {
    (apiClient.getUserOrders as jest.Mock)
      .mockRejectedValueOnce(new Error('API Error'))
      .mockResolvedValueOnce({
        orders: mockOrders,
        pagination: {
          page: 1,
          totalPages: 1,
          total: 2,
        },
      });
    
    render(<OrderHistory />);
    
    await waitFor(() => {
      expect(screen.getByText('Failed to load order history. Please try again.')).toBeInTheDocument();
    });
    
    const tryAgainButton = screen.getByText('Try Again');
    fireEvent.click(tryAgainButton);
    
    await waitFor(() => {
      expect(screen.getByText('Order History')).toBeInTheDocument();
    });
  });

  it('respects limit prop when provided', async () => {
    (apiClient.getUserOrders as jest.Mock).mockResolvedValue({
      orders: mockOrders.slice(0, 1),
      pagination: {
        page: 1,
        totalPages: 1,
        total: 1,
      },
    });
    
    render(<OrderHistory limit={1} />);
    
    await waitFor(() => {
      expect(apiClient.getUserOrders).toHaveBeenCalledWith(1, 1);
    });
  });

  it('shows pagination when there are multiple pages', async () => {
    (apiClient.getUserOrders as jest.Mock).mockResolvedValue({
      orders: mockOrders,
      pagination: {
        page: 1,
        totalPages: 3,
        total: 6,
      },
    });
    
    render(<OrderHistory />);
    
    await waitFor(() => {
      expect(screen.getByText('Page 1 of 3')).toBeInTheDocument();
      expect(screen.getByText('Previous')).toBeInTheDocument();
      expect(screen.getByText('Next')).toBeInTheDocument();
    });
  });

  it('does not show pagination when limit is provided', async () => {
    (apiClient.getUserOrders as jest.Mock).mockResolvedValue({
      orders: mockOrders,
      pagination: {
        page: 1,
        totalPages: 3,
        total: 6,
      },
    });
    
    render(<OrderHistory limit={5} />);
    
    await waitFor(() => {
      expect(screen.queryByText('Page 1 of 3')).not.toBeInTheDocument();
    });
  });
});