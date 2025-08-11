'use client';

import React, { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { Button } from '@/components/ui/Button';
import { Order, apiClient } from '@/lib/api';
import Link from 'next/link';

const OrderConfirmationPage: React.FC = () => {
  const params = useParams();
  const orderId = params.orderId as string;
  
  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchOrder = async () => {
      if (!orderId) {
        setError('Order ID not found');
        setLoading(false);
        return;
      }

      try {
        const orderData = await apiClient.getOrder(orderId);
        setOrder(orderData);
      } catch (error) {
        console.error('Error fetching order:', error);
        setError('Failed to load order details');
      } finally {
        setLoading(false);
      }
    };

    fetchOrder();
  }, [orderId]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-4xl mx-auto px-4">
          <div className="animate-pulse">
            <div className="h-8 bg-gray-300 rounded w-64 mb-8"></div>
            <div className="bg-white rounded-lg shadow p-6">
              <div className="space-y-4">
                <div className="h-4 bg-gray-300 rounded w-3/4"></div>
                <div className="h-4 bg-gray-300 rounded w-1/2"></div>
                <div className="h-4 bg-gray-300 rounded w-2/3"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !order) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-4xl mx-auto px-4">
          <div className="bg-red-50 border border-red-200 rounded-lg p-6">
            <h2 className="text-lg font-semibold text-red-800 mb-2">Error</h2>
            <p className="text-red-600 mb-4">{error || 'Order not found'}</p>
            <Link href="/">
              <Button>Return to Home</Button>
            </Link>
          </div>
        </div>
      </div>
    );
  }

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'confirmed':
      case 'paid':
        return 'text-green-800 bg-green-100';
      case 'pending':
        return 'text-yellow-800 bg-yellow-100';
      case 'processing':
        return 'text-blue-800 bg-blue-100';
      case 'shipped':
        return 'text-purple-800 bg-purple-100';
      case 'delivered':
        return 'text-green-800 bg-green-100';
      case 'cancelled':
        return 'text-red-800 bg-red-100';
      default:
        return 'text-gray-800 bg-gray-100';
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4">
        {/* Success Header */}
        <div className="text-center mb-8">
          <div className="w-16 h-16 mx-auto mb-4 bg-green-100 rounded-full flex items-center justify-center">
            <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Order Confirmed!</h1>
          <p className="text-lg text-gray-600">
            Thank you for your purchase. Your order has been successfully placed.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Order Details */}
          <div className="lg:col-span-2 space-y-6">
            {/* Order Info */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">Order Details</h2>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                <div>
                  <p className="text-sm text-gray-600">Order Number</p>
                  <p className="font-medium text-gray-900">#{order.id}</p>
                </div>
                
                <div>
                  <p className="text-sm text-gray-600">Order Date</p>
                  <p className="font-medium text-gray-900">
                    {new Date(order.createdAt).toLocaleDateString('en-IN', {
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                    })}
                  </p>
                </div>
                
                <div>
                  <p className="text-sm text-gray-600">Status</p>
                  <span className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(order.status)}`}>
                    {order.status.charAt(0).toUpperCase() + order.status.slice(1)}
                  </span>
                </div>
                
                <div>
                  <p className="text-sm text-gray-600">Total Amount</p>
                  <p className="font-medium text-gray-900">₹{order.total.toFixed(2)}</p>
                </div>
              </div>

              {/* Order Items */}
              <div>
                <h3 className="text-lg font-medium text-gray-900 mb-4">Items Ordered</h3>
                <div className="space-y-3">
                  {order.items?.map((item) => (
                    <div key={item.id} className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                      <div className="flex-1">
                        <h4 className="font-medium text-gray-900">
                          {item.product?.name || 'Product'}
                        </h4>
                        <p className="text-sm text-gray-500">
                          Quantity: {item.quantity} × ₹{item.price.toFixed(2)}
                        </p>
                      </div>
                      <div className="font-medium text-gray-900">
                        ₹{item.total.toFixed(2)}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {/* Shipping Address */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">Shipping Address</h2>
              <div className="text-sm text-gray-900">
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
                  <p className="mt-1">Phone: {order.shippingAddress.phone}</p>
                )}
              </div>
            </div>

            {/* Order Notes */}
            {order.notes && (
              <div className="bg-white rounded-lg shadow p-6">
                <h2 className="text-xl font-semibold text-gray-900 mb-4">Order Notes</h2>
                <p className="text-sm text-gray-900">{order.notes}</p>
              </div>
            )}
          </div>

          {/* Sidebar */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow p-6 sticky top-8">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">What&apos;s Next?</h2>
              
              <div className="space-y-4">
                <div className="flex items-start space-x-3">
                  <div className="w-6 h-6 bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                    <span className="text-xs font-medium text-blue-600">1</span>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">Order Confirmation</p>
                    <p className="text-xs text-gray-500">You&apos;ll receive an email confirmation shortly</p>
                  </div>
                </div>
                
                <div className="flex items-start space-x-3">
                  <div className="w-6 h-6 bg-gray-100 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                    <span className="text-xs font-medium text-gray-600">2</span>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">Processing</p>
                    <p className="text-xs text-gray-500">We&apos;ll prepare your order for shipment</p>
                  </div>
                </div>
                
                <div className="flex items-start space-x-3">
                  <div className="w-6 h-6 bg-gray-100 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                    <span className="text-xs font-medium text-gray-600">3</span>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">Shipping</p>
                    <p className="text-xs text-gray-500">Your order will be shipped within 2-3 business days</p>
                  </div>
                </div>
              </div>

              <div className="mt-6 space-y-3">
                <Link href="/products" className="block">
                  <Button className="w-full">Continue Shopping</Button>
                </Link>
                
                <Link href={`/orders/${order.id}`} className="block">
                  <Button variant="outline" className="w-full">Track Order</Button>
                </Link>
              </div>

              {/* Support */}
              <div className="mt-6 pt-6 border-t">
                <h3 className="text-sm font-medium text-gray-900 mb-2">Need Help?</h3>
                <p className="text-xs text-gray-500 mb-3">
                  If you have any questions about your order, please contact our support team.
                </p>
                <Button variant="outline" size="sm" className="w-full">
                  Contact Support
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default OrderConfirmationPage;