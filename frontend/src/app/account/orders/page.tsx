import { Metadata } from 'next';
import { OrderHistory } from '@/components/profile/OrderHistory';

export const metadata: Metadata = {
  title: 'Order History - Ecommerce Store',
  description: 'View your past orders and track their status.',
};

export default function OrderHistoryPage() {
  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Order History</h1>
          <p className="mt-2 text-gray-600">
            View and track all your past orders.
          </p>
        </div>
        <OrderHistory />
      </div>
    </div>
  );
}