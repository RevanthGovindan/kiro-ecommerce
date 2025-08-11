'use client';

import React, { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { CheckoutData } from '@/app/checkout/page';
import { Cart, apiClient } from '@/lib/api';
import { useRouter } from 'next/navigation';
import { useCart } from '@/contexts/CartContext';

interface OrderReviewProps {
  checkoutData: CheckoutData;
  cart: Cart;
  onBack: () => void;
}

interface RazorpayOptions {
  key: string;
  amount: number;
  currency: string;
  name: string;
  description: string;
  order_id: string;
  handler: (response: RazorpayResponse) => void;
  prefill: {
    name: string;
    email: string;
    contact: string;
  };
  notes: {
    order_id: string;
  };
  theme: {
    color: string;
  };
  modal: {
    ondismiss: () => void;
  };
}

interface RazorpayResponse {
  razorpay_order_id: string;
  razorpay_payment_id: string;
  razorpay_signature: string;
}

declare global {
  interface Window {
    Razorpay: new (options: RazorpayOptions) => {
      open: () => void;
    };
  }
}

export const OrderReview: React.FC<OrderReviewProps> = ({ checkoutData, cart, onBack }) => {
  const [isPlacingOrder, setIsPlacingOrder] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const { clearCart } = useCart();

  const loadRazorpayScript = (): Promise<boolean> => {
    return new Promise((resolve) => {
      if (window.Razorpay) {
        resolve(true);
        return;
      }

      const script = document.createElement('script');
      script.src = 'https://checkout.razorpay.com/v1/checkout.js';
      script.onload = () => resolve(true);
      script.onerror = () => resolve(false);
      document.body.appendChild(script);
    });
  };

  const handlePlaceOrder = async () => {
    if (!checkoutData.shippingAddress || !checkoutData.billingAddress) {
      setError('Missing address information');
      return;
    }

    setIsPlacingOrder(true);
    setError(null);

    try {
      // Load Razorpay script
      const scriptLoaded = await loadRazorpayScript();
      if (!scriptLoaded) {
        throw new Error('Failed to load payment gateway');
      }

      // Create order
      const order = await apiClient.createOrder({
        shippingAddress: checkoutData.shippingAddress,
        billingAddress: checkoutData.billingAddress,
        notes: checkoutData.notes,
      });

      // Create Razorpay payment order
      const payment = await apiClient.createPaymentOrder(
        order.id,
        Math.round(cart.total * 100), // Convert to paise
        'INR'
      );

      // Configure Razorpay options
      const options = {
        key: process.env.NEXT_PUBLIC_RAZORPAY_KEY_ID || '',
        amount: Math.round(cart.total * 100),
        currency: 'INR',
        name: 'Your Store Name',
        description: `Order #${order.id}`,
        order_id: payment.razorpayOrderId,
        handler: async (response: RazorpayResponse) => {
          try {
            // Verify payment
            await apiClient.verifyPayment(
              response.razorpay_order_id,
              response.razorpay_payment_id,
              response.razorpay_signature
            );

            // Clear cart and redirect to success page
            await clearCart();
            router.push(`/order-confirmation/${order.id}`);
          } catch (error) {
            console.error('Payment verification failed:', error);
            setError('Payment verification failed. Please contact support.');
            setIsPlacingOrder(false);
          }
        },
        prefill: {
          name: `${checkoutData.shippingAddress.firstName} ${checkoutData.shippingAddress.lastName}`,
          email: '', // Add email if available
          contact: checkoutData.shippingAddress.phone || '',
        },
        notes: {
          order_id: order.id,
        },
        theme: {
          color: '#2563eb',
        },
        modal: {
          ondismiss: () => {
            setIsPlacingOrder(false);
          },
        },
      };

      const razorpay = new window.Razorpay(options);
      razorpay.open();
    } catch (error) {
      console.error('Error placing order:', error);
      setError(error instanceof Error ? error.message : 'Failed to place order');
      setIsPlacingOrder(false);
    }
  };

  return (
    <div className="p-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-6">Review Your Order</h2>
      
      {error && (
        <div className="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex">
            <svg className="h-5 w-5 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error</h3>
              <p className="text-sm text-red-700 mt-1">{error}</p>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-6">
        {/* Order Items */}
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">Order Items</h3>
          <div className="bg-gray-50 rounded-lg p-4">
            <div className="space-y-3">
              {cart.items.map((item) => (
                <div key={item.productId} className="flex justify-between items-center">
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
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">Shipping Address</h3>
          <div className="bg-gray-50 rounded-lg p-4">
            <div className="text-sm text-gray-900">
              <p className="font-medium">
                {checkoutData.shippingAddress?.firstName} {checkoutData.shippingAddress?.lastName}
              </p>
              {checkoutData.shippingAddress?.company && (
                <p>{checkoutData.shippingAddress.company}</p>
              )}
              <p>{checkoutData.shippingAddress?.address1}</p>
              {checkoutData.shippingAddress?.address2 && (
                <p>{checkoutData.shippingAddress.address2}</p>
              )}
              <p>
                {checkoutData.shippingAddress?.city}, {checkoutData.shippingAddress?.state} {checkoutData.shippingAddress?.postalCode}
              </p>
              <p>{checkoutData.shippingAddress?.country}</p>
              {checkoutData.shippingAddress?.phone && (
                <p className="mt-1">Phone: {checkoutData.shippingAddress.phone}</p>
              )}
            </div>
          </div>
        </div>

        {/* Billing Address */}
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">Billing Address</h3>
          <div className="bg-gray-50 rounded-lg p-4">
            {checkoutData.sameAsShipping ? (
              <p className="text-sm text-gray-600 italic">Same as shipping address</p>
            ) : (
              <div className="text-sm text-gray-900">
                <p className="font-medium">
                  {checkoutData.billingAddress?.firstName} {checkoutData.billingAddress?.lastName}
                </p>
                {checkoutData.billingAddress?.company && (
                  <p>{checkoutData.billingAddress.company}</p>
                )}
                <p>{checkoutData.billingAddress?.address1}</p>
                {checkoutData.billingAddress?.address2 && (
                  <p>{checkoutData.billingAddress.address2}</p>
                )}
                <p>
                  {checkoutData.billingAddress?.city}, {checkoutData.billingAddress?.state} {checkoutData.billingAddress?.postalCode}
                </p>
                <p>{checkoutData.billingAddress?.country}</p>
                {checkoutData.billingAddress?.phone && (
                  <p className="mt-1">Phone: {checkoutData.billingAddress.phone}</p>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Order Notes */}
        {checkoutData.notes && (
          <div>
            <h3 className="text-lg font-medium text-gray-900 mb-4">Order Notes</h3>
            <div className="bg-gray-50 rounded-lg p-4">
              <p className="text-sm text-gray-900">{checkoutData.notes}</p>
            </div>
          </div>
        )}

        {/* Payment Method */}
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">Payment Method</h3>
          <div className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-center">
              <div className="w-8 h-5 bg-blue-600 rounded text-white text-xs flex items-center justify-center font-bold mr-2">
                RZP
              </div>
              <span className="text-sm text-gray-900">Razorpay (Credit/Debit Card, UPI, Net Banking)</span>
            </div>
          </div>
        </div>

        {/* Order Summary */}
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">Order Summary</h3>
          <div className="bg-gray-50 rounded-lg p-4">
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Subtotal</span>
                <span className="text-gray-900">₹{cart.subtotal.toFixed(2)}</span>
              </div>
              
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Tax</span>
                <span className="text-gray-900">₹{cart.tax.toFixed(2)}</span>
              </div>
              
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Shipping</span>
                <span className="text-gray-900">Free</span>
              </div>
              
              <div className="border-t pt-2">
                <div className="flex justify-between font-semibold">
                  <span className="text-gray-900">Total</span>
                  <span className="text-gray-900">₹{cart.total.toFixed(2)}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex justify-between pt-6 border-t mt-6">
        <Button
          type="button"
          variant="outline"
          onClick={onBack}
          disabled={isPlacingOrder}
        >
          Back to Payment
        </Button>
        
        <Button
          onClick={handlePlaceOrder}
          size="lg"
          disabled={isPlacingOrder}
        >
          {isPlacingOrder ? 'Processing...' : 'Place Order'}
        </Button>
      </div>
    </div>
  );
};