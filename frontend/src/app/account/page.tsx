'use client';

// import { Metadata } from 'next';
import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { useAuth } from '@/contexts/AuthContext';
import { OrderHistory } from '@/components/profile/OrderHistory';

export default function AccountPage() {
  const { user, isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto text-center">
          <h1 className="text-3xl font-bold text-gray-900 mb-4">Account Dashboard</h1>
          <p className="text-gray-600 mb-8">Please log in to access your account.</p>
          <div className="space-x-4">
            <Link href="/auth/login">
              <Button>Login</Button>
            </Link>
            <Link href="/auth/register">
              <Button variant="outline">Sign Up</Button>
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        {/* Welcome Section */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">
            Welcome back, {user?.firstName}!
          </h1>
          <p className="mt-2 text-gray-600">
            Manage your account, view orders, and update your preferences.
          </p>
        </div>

        {/* Quick Actions */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <Link href="/account/profile">
            <div className="bg-white p-6 rounded-lg shadow-md hover:shadow-lg transition-shadow cursor-pointer">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <svg className="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                </div>
                <div className="ml-4">
                  <h3 className="text-lg font-medium text-gray-900">Profile</h3>
                  <p className="text-sm text-gray-600">Update your personal information</p>
                </div>
              </div>
            </div>
          </Link>

          <Link href="/account/orders">
            <div className="bg-white p-6 rounded-lg shadow-md hover:shadow-lg transition-shadow cursor-pointer">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
                  </svg>
                </div>
                <div className="ml-4">
                  <h3 className="text-lg font-medium text-gray-900">Orders</h3>
                  <p className="text-sm text-gray-600">View and track your orders</p>
                </div>
              </div>
            </div>
          </Link>

          <Link href="/products">
            <div className="bg-white p-6 rounded-lg shadow-md hover:shadow-lg transition-shadow cursor-pointer">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <svg className="w-8 h-8 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4" />
                  </svg>
                </div>
                <div className="ml-4">
                  <h3 className="text-lg font-medium text-gray-900">Shop</h3>
                  <p className="text-sm text-gray-600">Browse our products</p>
                </div>
              </div>
            </div>
          </Link>
        </div>

        {/* Account Information */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-1">
            <div className="bg-white p-6 rounded-lg shadow-md">
              <h2 className="text-lg font-semibold mb-4">Account Information</h2>
              <div className="space-y-3">
                <div>
                  <span className="text-sm text-gray-600">Name:</span>
                  <p className="font-medium">{user?.firstName} {user?.lastName}</p>
                </div>
                <div>
                  <span className="text-sm text-gray-600">Email:</span>
                  <p className="font-medium">{user?.email}</p>
                </div>
                {user?.phone && (
                  <div>
                    <span className="text-sm text-gray-600">Phone:</span>
                    <p className="font-medium">{user.phone}</p>
                  </div>
                )}
                <div>
                  <span className="text-sm text-gray-600">Member since:</span>
                  <p className="font-medium">
                    {user?.createdAt ? new Date(user.createdAt).toLocaleDateString('en-US', {
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                    }) : 'N/A'}
                  </p>
                </div>
              </div>
              <div className="mt-6">
                <Link href="/account/profile">
                  <Button variant="outline" className="w-full">
                    Edit Profile
                  </Button>
                </Link>
              </div>
            </div>
          </div>

          <div className="lg:col-span-2">
            <div className="bg-white p-6 rounded-lg shadow-md">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-lg font-semibold">Recent Orders</h2>
                <Link href="/account/orders">
                  <Button variant="outline" size="sm">
                    View All
                  </Button>
                </Link>
              </div>
              <OrderHistory limit={3} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}