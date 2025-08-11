'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { useAuth } from '@/contexts/AuthContext';
import { apiClient, Order } from '@/lib/api';

interface OrderDetailsProps {
  orderId: string;
}

export const OrderDetails: React.FC<OrderDetailsProps> = ({ orderId }) => {
  const { isAuthenticated } = useAuth();
  const [order, setOrder] = useState<Order | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    if (isAuthenticated && orderId) {
      fetchOrder();
    }
  }, [isAuthenticated, orderId]);

  const fetchOrder = async () => {
    try {
      setIsLoading(true);
      const orderData = await apiClient.getOrder(orderId);
      setOrder(orderData);
    } catch (error) {
      console.error('Failed to fetch order:', error);
      setError('Failed to load order details. Please try again.');
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
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
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
        <p className="text-gray-600 mb-4">Please log in to view order details.</p>
        <Link href="/auth/login">
          <Button>Login</Button>
        </Link>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-white p-6 rounded-lg shadow-md animate-pulse">
          <div className="space-y-4">
            <div className="h-8 bg-gray-200 rounded w-1/3"></div>
            <div className="h-4 bg-gray-200 rounded w-1/4"></div>
            <div className="space-y-2">
              <div className="h-4 bg-gray-200 rounded w-full"></div>
              <div className="h-4 bg-gray-200 rounded w-3/4"></div>
              <div className="h-4 bg-gray-200 rounded w-1/2"></div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !order) {
    return (
      <div className="text-center py-8">
        <p className="text-red-600 mb-4">{error || 'Order not found'}</p>
        <div className="space-x-4">
          <Button onClick={fetchOrder}>Try Again</Button>
          <Link href="/account/orders">
            <Button variant="outline">Back to Orders</Button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Header */}
      <div className="bg-white p-6 rounded-lg shadow-md">
        <div className="flex justify-between items-start mb-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">
              Order #{order.id.slice(-8).toUpperCase()}
            </h1>
            <p className="text-gray-600">
              Placed on {formatDate(order.createdAt)}
            </p>
          </div>
          <span className={`px-4 py-2 rounded-full text-sm font-medium ${getStatusColor(order.status)}`}>
            {order.status.charAt(0).toUpperCase() + order.status.slice(1)}
          </span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
          <div>
            <span className="text-gray-600">Order Total:</span>
            <p className="font-semibold text-lg">{formatCurrency(order.total)}</p>
          </div>
          <div>
            <span className="text-gray-600">Payment Method:</span>
            <p className="font-medium">Credit Card</p>
          </div>
          <div>
            <span className="text-gray-600">Payment Status:</span>
            <p className="font-medium text-green-600">Paid</p>
          </div>
        </div>
      </div>

      {/* Order Items */}
      <div className="bg-white p-6 rounded-lg shadow-md">
        <h2 className="text-lg font-semibold mb-4">Order Items</h2>
        <div className="space-y-4">
          {order.items?.map((item) => (
            <div key={item.id} className="flex items-center space-x-4 py-4 border-b border-gray-200 last:border-b-0">
              <div className="w-16 h-16 bg-gray-200 rounded-md flex items-center justify-center">
                {item.product?.images && item.product.images.length > 0 ? (
                  <img
                    src={item.product.images[0]}
                    alt={item.product.name}
                    className="w-full h-full object-cover rounded-md"
                  />
                ) : (
                  <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                )}
              </div>
              <div className="flex-1">
                <h3 className="font-medium text-gray-900">
                  {item.product?.name || 'Product'}
                </h3>
                <p className="text-sm text-gray-600">
                  SKU: {item.product?.sku || 'N/A'}
                </p>
                <p className="text-sm text-gray-600">
                  Quantity: {item.quantity}
                </p>
              </div>
              <div className="text-right">
                <p className="font-medium">{formatCurrency(item.price)}</p>
                <p className="text-sm text-gray-600">each</p>
                <p className="font-semibold">{formatCurrency(item.total)}</p>
              </div>
            </div>
          ))}
        </div>

        {/* Order Summary */}
        <div className="mt-6 pt-6 border-t border-gray-200">
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600">Subtotal:</span>
              <span>{formatCurrency(order.subtotal)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600">Tax:</span>
              <span>{formatCurrency(order.tax)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600">Shipping:</span>
              <span>{formatCurrency(order.shipping)}</span>
            </div>
            <div className="flex justify-between text-lg font-semibold pt-2 border-t border-gray-200">
              <span>Total:</span>
              <span>{formatCurrency(order.total)}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Shipping Information */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white p-6 rounded-lg shadow-md">
          <h2 className="text-lg font-semibold mb-4">Shipping Address</h2>
          <div className="text-sm space-y-1">
            <p className="font-medium">
              {order.shippingAddress.firstName} {order.shippingAddress.lastName}
            </p>
            {order.shippingAddress.company && (
              <p>{order.shippingAddress.company}</p>
            )}
            <p>{order.shippingAddress.address1}</p>
            {order.shippingAddress.address2 && (
              <p>{order.shippingAddress.address2}</p>
            )}
            <p>
              {order.shippingAddress.city}, {order.shippingAddress.state} {order.shippingAddress.postalCode}
            </p>
            <p>{order.shippingAddress.country}</p>
            {order.shippingAddress.phone && (
              <p className="pt-2">Phone: {order.shippingAddress.phone}</p>
            )}
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <h2 className="text-lg font-semibold mb-4">Billing Address</h2>
          <div className="text-sm space-y-1">
            <p className="font-medium">
              {order.billingAddress.firstName} {order.billingAddress.lastName}
            </p>
            {order.billingAddress.company && (
              <p>{order.billingAddress.company}</p>
            )}
            <p>{order.billingAddress.address1}</p>
            {order.billingAddress.address2 && (
              <p>{order.billingAddress.address2}</p>
            )}
            <p>
              {order.billingAddress.city}, {order.billingAddress.state} {order.billingAddress.postalCode}
            </p>
            <p>{order.billingAddress.country}</p>
            {order.billingAddress.phone && (
              <p className="pt-2">Phone: {order.billingAddress.phone}</p>
            )}
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex justify-between items-center">
        <Link href="/account/orders">
          <Button variant="outline">← Back to Orders</Button>
        </Link>
        
        <div className="space-x-4">
          {order.status === 'delivered' && (
            <Button variant="outline">Leave Review</Button>
          )}
          <Button variant="outline">Contact Support</Button>
        </div>
      </div>
    </div>
  );
};