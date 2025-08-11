'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { useAuth } from '@/contexts/AuthContext';
import { apiClient, Order } from '@/lib/api';

interface OrderHistoryProps {
  limit?: number;
}

export const OrderHistory: React.FC<OrderHistoryProps> = ({ limit }) => {
  const { isAuthenticated } = useAuth();
  const [orders, setOrders] = useState<Order[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [pagination, setPagination] = useState({
    page: 1,
    totalPages: 1,
    total: 0,
  });

  useEffect(() => {
    if (isAuthenticated) {
      fetchOrders();
    }
  }, [isAuthenticated, pagination.page]);

  const fetchOrders = async () => {
    try {
      setIsLoading(true);
      const response = await apiClient.getUserOrders(pagination.page, limit || 10);
      setOrders(response.orders);
      setPagination({
        page: response.pagination.page,
        totalPages: response.pagination.totalPages,
        total: response.pagination.total,
      });
    } catch (error) {
      console.error('Failed to fetch orders:', error);
      setError('Failed to load order history. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'confirmed':
        return 'bg-blue-100 text-blue-800';
      case 'processing':
        return 'bg-purple-100 text-purple-800';
      case 'shipped':
        return 'bg-indigo-100 text-indigo-800';
      case 'delivered':
        return 'bg-green-100 text-green-800';
      case 'cancelled':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount);
  };

  if (!isAuthenticated) {
    return (
      <div className="text-center py-8">
        <p className="text-gray-600 mb-4">Please log in to view your order history.</p>
        <Link href="/auth/login">
          <Button>Login</Button>
        </Link>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        {[...Array(3)].map((_, i) => (
          <div key={i} className="bg-white p-6 rounded-lg shadow-md animate-pulse">
            <div className="flex justify-between items-start mb-4">
              <div className="space-y-2">
                <div className="h-4 bg-gray-200 rounded w-32"></div>
                <div className="h-3 bg-gray-200 rounded w-24"></div>
              </div>
              <div className="h-6 bg-gray-200 rounded w-20"></div>
            </div>
            <div className="space-y-2">
              <div className="h-3 bg-gray-200 rounded w-full"></div>
              <div className="h-3 bg-gray-200 rounded w-3/4"></div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8">
        <p className="text-red-600 mb-4">{error}</p>
        <Button onClick={fetchOrders}>Try Again</Button>
      </div>
    );
  }

  if (orders.length === 0) {
    return (
      <div className="text-center py-8">
        <div className="mb-4">
          <svg className="w-16 h-16 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
          </svg>
        </div>
        <h3 className="text-lg font-medium text-gray-900 mb-2">No orders yet</h3>
        <p className="text-gray-600 mb-4">You haven&apos;t placed any orders yet.</p>
        <Link href="/products">
          <Button>Start Shopping</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold">Order History</h2>
        {pagination.total > 0 && (
          <p className="text-sm text-gray-600">
            {pagination.total} order{pagination.total !== 1 ? 's' : ''} total
          </p>
        )}
      </div>

      <div className="space-y-4">
        {orders.map((order) => (
          <div key={order.id} className="bg-white p-6 rounded-lg shadow-md border">
            <div className="flex justify-between items-start mb-4">
              <div>
                <h3 className="font-medium text-gray-900">
                  Order #{order.id.slice(-8).toUpperCase()}
                </h3>
                <p className="text-sm text-gray-600">
                  Placed on {formatDate(order.createdAt)}
                </p>
              </div>
              <span className={`px-3 py-1 rounded-full text-xs font-medium ${getStatusColor(order.status)}`}>
                {order.status.charAt(0).toUpperCase() + order.status.slice(1)}
              </span>
            </div>

            <div className="border-t border-gray-200 pt-4">
              <div className="flex justify-between items-center mb-2">
                <span className="text-sm text-gray-600">Total Amount:</span>
                <span className="font-medium">{formatCurrency(order.total)}</span>
              </div>
              
              {order.items && order.items.length > 0 && (
                <div className="text-sm text-gray-600 mb-3">
                  {order.items.length} item{order.items.length !== 1 ? 's' : ''}
                  {order.items.length <= 2 && order.items.map((item, index) => (
                    <span key={item.id}>
                      {index === 0 ? ': ' : ', '}
                      {item.product?.name || 'Product'} (×{item.quantity})
                    </span>
                  ))}
                  {order.items.length > 2 && (
                    <span>: {order.items[0].product?.name || 'Product'} and {order.items.length - 1} more</span>
                  )}
                </div>
              )}

              <div className="flex justify-between items-center">
                <div className="text-sm text-gray-600">
                  <span>Shipping to: </span>
                  <span className="font-medium">
                    {order.shippingAddress.city}, {order.shippingAddress.state}
                  </span>
                </div>
                <Link href={`/account/orders/${order.id}`}>
                  <Button variant="outline" size="sm">
                    View Details
                  </Button>
                </Link>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Pagination */}
      {!limit && pagination.totalPages > 1 && (
        <div className="flex justify-center items-center space-x-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPagination(prev => ({ ...prev, page: prev.page - 1 }))}
            disabled={pagination.page <= 1}
          >
            Previous
          </Button>
          <span className="text-sm text-gray-600">
            Page {pagination.page} of {pagination.totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPagination(prev => ({ ...prev, page: prev.page + 1 }))}
            disabled={pagination.page >= pagination.totalPages}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
};