'use client';

import React, { useEffect, useState } from 'react';
import { apiClient, SalesMetrics } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import Link from 'next/link';

export const AdminDashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<SalesMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        setLoading(true);
        const data = await apiClient.getSalesMetrics();
        setMetrics(data);
      } catch (err) {
        setError('Failed to load dashboard metrics');
        console.error('Error fetching metrics:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchMetrics();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4">
        <p className="text-red-800">{error}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <div className="text-sm text-gray-500">
          Welcome to the admin panel
        </div>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <div className="w-8 h-8 bg-green-500 rounded-md flex items-center justify-center">
                <span className="text-white text-sm">💰</span>
              </div>
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-500">Total Revenue</p>
              <p className="text-2xl font-semibold text-gray-900">
                ₹{metrics?.totalRevenue?.toLocaleString() || '0'}
              </p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <div className="w-8 h-8 bg-blue-500 rounded-md flex items-center justify-center">
                <span className="text-white text-sm">📦</span>
              </div>
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-500">Total Orders</p>
              <p className="text-2xl font-semibold text-gray-900">
                {metrics?.totalOrders || 0}
              </p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <div className="w-8 h-8 bg-purple-500 rounded-md flex items-center justify-center">
                <span className="text-white text-sm">📊</span>
              </div>
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-500">Avg Order Value</p>
              <p className="text-2xl font-semibold text-gray-900">
                ₹{metrics?.averageOrderValue?.toFixed(2) || '0.00'}
              </p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <div className="w-8 h-8 bg-orange-500 rounded-md flex items-center justify-center">
                <span className="text-white text-sm">🏆</span>
              </div>
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-500">Top Products</p>
              <p className="text-2xl font-semibold text-gray-900">
                {metrics?.topProducts?.length || 0}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-medium text-gray-900">Quick Actions</h2>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <Link href="/admin/products">
              <Button variant="outline" className="w-full justify-start">
                <span className="mr-2">📦</span>
                Manage Products
              </Button>
            </Link>
            <Link href="/admin/orders">
              <Button variant="outline" className="w-full justify-start">
                <span className="mr-2">🛒</span>
                View Orders
              </Button>
            </Link>
            <Link href="/admin/customers">
              <Button variant="outline" className="w-full justify-start">
                <span className="mr-2">👥</span>
                Manage Customers
              </Button>
            </Link>
            <Button variant="outline" className="w-full justify-start" disabled>
              <span className="mr-2">📈</span>
              Analytics (Coming Soon)
            </Button>
          </div>
        </div>
      </div>

      {/* Top Products */}
      {metrics?.topProducts && metrics.topProducts.length > 0 && (
        <div className="bg-white rounded-lg shadow">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">Top Selling Products</h2>
          </div>
          <div className="p-6">
            <div className="space-y-4">
              {metrics.topProducts.slice(0, 5).map((item, index) => (
                <div key={item.product.id} className="flex items-center justify-between">
                  <div className="flex items-center">
                    <div className="flex-shrink-0 w-8 h-8 bg-gray-200 rounded-md flex items-center justify-center">
                      <span className="text-sm font-medium text-gray-600">#{index + 1}</span>
                    </div>
                    <div className="ml-4">
                      <p className="text-sm font-medium text-gray-900">{item.product.name}</p>
                      <p className="text-sm text-gray-500">{item.totalSold} sold</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-medium text-gray-900">
                      ₹{item.revenue.toLocaleString()}
                    </p>
                    <p className="text-sm text-gray-500">revenue</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};