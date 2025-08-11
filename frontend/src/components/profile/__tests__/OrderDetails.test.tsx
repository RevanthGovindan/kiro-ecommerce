import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { OrderDetails } from '../OrderDetails';
import { useAuth } from '../../../contexts/AuthContext';
import { apiClient } from '../../../lib/api';

// Mock the dependencies
jest.mock('../../../contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}));

jest.mock('../../../lib/api', () => ({
  apiClient: {
    getOrder: jest.fn(),
  },
}));

const mockOrder = {
  id: 'order-123',
  userId: 'user-1',
  status: 'delivered',
  subtotal: 100,
  tax: 10,
  shipping: 5,
  total: 115,
  shippingAddress: {
    firstName: 'John',
    lastName: 'Doe',
    company: 'Test Company',
    address1: '123 Main St',
    address2: 'Apt 4B',
    city: 'New York',
    state: 'NY',
    postalCode: '10001',
    country: 'US',
    phone: '+1234567890',
  },
  billingAddress: {
    firstName: 'Jane',
    lastName: 'Smith',
    address1: '456 Oak Ave',
    city: 'Boston',
    state: 'MA',
    postalCode: '02101',
    country: 'US',
    phone: '+0987654321',
  },
  paymentIntentId: 'pi_test',
  createdAt: '2023-01-01T12:00:00Z',
  updatedAt: '2023-01-01T12:00:00Z',
  items: [
    {
      id: 'item-1',
      orderId: 'order-123',
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
        description: 'A test product',
        inventory: 10,
        isActive: true,
        categoryId: 'cat-1',
        specifications: {},
        createdAt: '2023-01-01T00:00:00Z',
        updatedAt: '2023-01-01T00:00:00Z',
      },
    },
  ],
};

beforeEach(() => {
  (useAuth as jest.Mock).mockReturnValue({
    isAuthenticated: true,
  });
  
  jest.clearAllMocks();
});

describe('OrderDetails', () => {
  const orderId = 'order-123';

  it('shows login prompt when user is not authenticated', () => {
    (useAuth as jest.Mock).mockReturnValue({
      isAuthenticated: false,
    });
    
    render(<OrderDetails orderId={orderId} />);
    
    expect(screen.getByText('Please log in to view order details.')).toBeInTheDocument();
    expect(screen.getByText('Login')).toBeInTheDocument();
  });

  it('shows loading state while fetching order', () => {
    (apiClient.getOrder as jest.Mock).mockImplementation(() => 
      new Promise(() => {}) // Never resolves
    );
    
    render(<OrderDetails orderId={orderId} />);
    
    // Should show loading skeleton
    expect(screen.getByRole('generic')).toBeInTheDocument();
  });

  it('displays order details when loaded successfully', async () => {
    (apiClient.getOrder as jest.Mock).mockResolvedValue(mockOrder);
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      expect(screen.getByText((content, element) => {
        return element?.tagName.toLowerCase() === 'h1' && content.includes('Order #');
      })).toBeInTheDocument();
      expect(screen.getByText((content) => content.includes('Placed on January 1, 2023'))).toBeInTheDocument();
      expect(screen.getByText('Delivered')).toBeInTheDocument();
      expect(screen.getAllByText('$115.00')).toHaveLength(2); // Order total appears twice
    });
  });

  it('displays order items correctly', async () => {
    (apiClient.getOrder as jest.Mock).mockResolvedValue(mockOrder);
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      expect(screen.getByText('Order Items')).toBeInTheDocument();
      expect(screen.getByText('Test Product')).toBeInTheDocument();
      expect(screen.getByText('SKU: TEST-001')).toBeInTheDocument();
      expect(screen.getByText('Quantity: 2')).toBeInTheDocument();
      expect(screen.getByText('$50.00')).toBeInTheDocument();
      expect(screen.getByText('each')).toBeInTheDocument();
      expect(screen.getByText('$100.00')).toBeInTheDocument();
    });
  });

  it('displays order summary correctly', async () => {
    (apiClient.getOrder as jest.Mock).mockResolvedValue(mockOrder);
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      expect(screen.getByText('Subtotal:')).toBeInTheDocument();
      expect(screen.getByText('Tax:')).toBeInTheDocument();
      expect(screen.getByText('Shipping:')).toBeInTheDocument();
      expect(screen.getByText('Total:')).toBeInTheDocument();
      
      // Check amounts
      const subtotalElements = screen.getAllByText('$100.00');
      const taxElements = screen.getAllByText('$10.00');
      const shippingElements = screen.getAllByText('$5.00');
      const totalElements = screen.getAllByText('$115.00');
      
      expect(subtotalElements.length).toBeGreaterThan(0);
      expect(taxElements.length).toBeGreaterThan(0);
      expect(shippingElements.length).toBeGreaterThan(0);
      expect(totalElements.length).toBeGreaterThan(0);
    });
  });

  it('displays shipping address correctly', async () => {
    (apiClient.getOrder as jest.Mock).mockResolvedValue(mockOrder);
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      expect(screen.getByText('Shipping Address')).toBeInTheDocument();
      expect(screen.getByText('John Doe')).toBeInTheDocument();
      expect(screen.getByText('Test Company')).toBeInTheDocument();
      expect(screen.getByText('123 Main St')).toBeInTheDocument();
      expect(screen.getByText('Apt 4B')).toBeInTheDocument();
      expect(screen.getByText('New York, NY 10001')).toBeInTheDocument();
      expect(screen.getAllByText('US')).toHaveLength(2); // Both shipping and billing
      expect(screen.getByText('Phone: +1234567890')).toBeInTheDocument();
    });
  });

  it('displays billing address correctly', async () => {
    (apiClient.getOrder as jest.Mock).mockResolvedValue(mockOrder);
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      expect(screen.getByText('Billing Address')).toBeInTheDocument();
      expect(screen.getByText('Jane Smith')).toBeInTheDocument();
      expect(screen.getByText('456 Oak Ave')).toBeInTheDocument();
      expect(screen.getByText('Boston, MA 02101')).toBeInTheDocument();
      expect(screen.getByText('Phone: +0987654321')).toBeInTheDocument();
    });
  });

  it('shows back to orders link', async () => {
    (apiClient.getOrder as jest.Mock).mockResolvedValue(mockOrder);
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      expect(screen.getByText('← Back to Orders')).toBeInTheDocument();
    });
  });

  it('shows action buttons', async () => {
    (apiClient.getOrder as jest.Mock).mockResolvedValue(mockOrder);
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      expect(screen.getByText('Leave Review')).toBeInTheDocument();
      expect(screen.getByText('Contact Support')).toBeInTheDocument();
    });
  });

  it('only shows leave review button for delivered orders', async () => {
    const pendingOrder = { ...mockOrder, status: 'pending' };
    (apiClient.getOrder as jest.Mock).mockResolvedValue(pendingOrder);
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      expect(screen.queryByText('Leave Review')).not.toBeInTheDocument();
      expect(screen.getByText('Contact Support')).toBeInTheDocument();
    });
  });

  it('handles API error gracefully', async () => {
    (apiClient.getOrder as jest.Mock).mockRejectedValue(new Error('Order not found'));
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      expect(screen.getByText('Failed to load order details. Please try again.')).toBeInTheDocument();
      expect(screen.getByText('Try Again')).toBeInTheDocument();
      expect(screen.getByText('Back to Orders')).toBeInTheDocument();
    });
  });

  it('retries loading order when try again is clicked', async () => {
    (apiClient.getOrder as jest.Mock)
      .mockRejectedValueOnce(new Error('API Error'))
      .mockResolvedValueOnce(mockOrder);
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      expect(screen.getByText('Failed to load order details. Please try again.')).toBeInTheDocument();
    });
    
    const tryAgainButton = screen.getByText('Try Again');
    fireEvent.click(tryAgainButton);
    
    await waitFor(() => {
      expect(screen.getByText((content, element) => {
        return element?.tagName.toLowerCase() === 'h1' && content.includes('Order #');
      })).toBeInTheDocument();
    });
  });

  it('shows placeholder image when product has no images', async () => {
    const orderWithoutImages = {
      ...mockOrder,
      items: [
        {
          ...mockOrder.items[0],
          product: {
            ...mockOrder.items[0].product!,
            images: [],
          },
        },
      ],
    };
    
    (apiClient.getOrder as jest.Mock).mockResolvedValue(orderWithoutImages);
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      // Should show SVG placeholder icon - check for the SVG element
      const container = screen.getByText('Test Product').closest('.flex');
      const svgElement = container?.querySelector('svg');
      expect(svgElement).toBeInTheDocument();
    });
  });

  it('handles missing optional address fields', async () => {
    const orderWithMinimalAddress = {
      ...mockOrder,
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
        firstName: 'Jane',
        lastName: 'Smith',
        address1: '456 Oak Ave',
        city: 'Boston',
        state: 'MA',
        postalCode: '02101',
        country: 'US',
      },
    };
    
    (apiClient.getOrder as jest.Mock).mockResolvedValue(orderWithMinimalAddress);
    
    render(<OrderDetails orderId={orderId} />);
    
    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument();
      expect(screen.getByText('Jane Smith')).toBeInTheDocument();
      // Should not show company or phone fields
      expect(screen.queryByText('Test Company')).not.toBeInTheDocument();
      expect(screen.queryByText('Phone:')).not.toBeInTheDocument();
    });
  });
});