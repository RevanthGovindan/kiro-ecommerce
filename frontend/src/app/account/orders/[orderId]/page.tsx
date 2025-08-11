import { Metadata } from 'next';
import { OrderDetails } from '@/components/profile/OrderDetails';

interface OrderDetailsPageProps {
  params: Promise<{
    orderId: string;
  }>;
}

export const metadata: Metadata = {
  title: 'Order Details - Ecommerce Store',
  description: 'View detailed information about your order.',
};

export default async function OrderDetailsPage({ params }: OrderDetailsPageProps) {
  const { orderId } = await params;
  
  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Order Details</h1>
          <p className="mt-2 text-gray-600">
            Detailed information about your order.
          </p>
        </div>
        <OrderDetails orderId={orderId} />
      </div>
    </div>
  );
}