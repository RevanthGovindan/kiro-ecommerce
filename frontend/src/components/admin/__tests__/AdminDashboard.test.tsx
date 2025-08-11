import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { AdminDashboard } from '../AdminDashboard';
import { apiClient } from '../../../lib/api';

// Mock the API client
jest.mock('../../../lib/api', () => ({
  apiClient: {
    getSalesMetrics: jest.fn(),
  },
}));

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

const mockMetrics = {
  totalRevenue: 150000,
  totalOrders: 250,
  averageOrderValue: 600,
  topProducts: [
    {
      product: {
        id: '1',
        name: 'Test Product 1',
        description: 'Test description',
        price: 100,
        sku: 'TEST-001',
        inventory: 50,
        isActive: true,
        categoryId: 'cat1',
        images: [],
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      },
      totalSold: 50,
      revenue: 5000,
    },
  ],
  revenueByMonth: [
    {
      month: '2024-01',
      revenue: 50000,
      orders: 100,
    },
  ],
};

describe('AdminDashboard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders loading state initially', () => {
    mockApiClient.getSalesMetrics.mockImplementation(() => new Promise(() => {}));
    
    render(<AdminDashboard />);
    
    expect(screen.getByRole('generic')).toHaveClass('animate-spin');
  });

  it('renders dashboard with metrics data', async () => {
    mockApiClient.getSalesMetrics.mockResolvedValue(mockMetrics);
    
    render(<AdminDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });

    expect(screen.getByText((content, element) => {
      return element?.textContent === '₹150,000';
    })).toBeInTheDocument();
    expect(screen.getByText('250')).toBeInTheDocument();
    expect(screen.getByText((content, element) => {
      return element?.textContent === '₹600.00';
    })).toBeInTheDocument();
    expect(screen.getByText('1')).toBeInTheDocument();
  });

  it('renders quick action buttons', async () => {
    mockApiClient.getSalesMetrics.mockResolvedValue(mockMetrics);
    
    render(<AdminDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('Quick Actions')).toBeInTheDocument();
    });

    expect(screen.getByText('Manage Products')).toBeInTheDocument();
    expect(screen.getByText('View Orders')).toBeInTheDocument();
    expect(screen.getByText('Manage Customers')).toBeInTheDocument();
  });

  it('renders top products section when data is available', async () => {
    mockApiClient.getSalesMetrics.mockResolvedValue(mockMetrics);
    
    render(<AdminDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('Top Selling Products')).toBeInTheDocument();
    });

    expect(screen.getByText('Test Product 1')).toBeInTheDocument();
    expect(screen.getByText('50 sold')).toBeInTheDocument();
    expect(screen.getByText((content, element) => {
      return element?.textContent === '₹5,000';
    })).toBeInTheDocument();
  });

  it('handles API error gracefully', async () => {
    mockApiClient.getSalesMetrics.mockRejectedValue(new Error('API Error'));
    
    render(<AdminDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('Failed to load dashboard metrics')).toBeInTheDocument();
    });
  });

  it('does not render top products section when no data', async () => {
    const metricsWithoutProducts = {
      ...mockMetrics,
      topProducts: [],
    };
    mockApiClient.getSalesMetrics.mockResolvedValue(metricsWithoutProducts);
    
    render(<AdminDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });

    expect(screen.queryByText('Top Selling Products')).not.toBeInTheDocument();
  });
});