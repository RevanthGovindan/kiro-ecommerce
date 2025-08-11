'use client';

import React, { useState, useEffect } from 'react';
import { useCart } from '@/contexts/CartContext';
import { useRouter } from 'next/navigation';
import { CheckoutSteps } from '@/components/checkout/CheckoutSteps';
import { ShippingForm } from '@/components/checkout/ShippingForm';
import { PaymentForm } from '@/components/checkout/PaymentForm';
import { OrderReview } from '@/components/checkout/OrderReview';
import { OrderAddress } from '@/lib/api';

export type CheckoutStep = 'shipping' | 'payment' | 'review';

export interface CheckoutData {
  shippingAddress: OrderAddress | null;
  billingAddress: OrderAddress | null;
  sameAsShipping: boolean;
  notes?: string;
}

const CheckoutPage: React.FC = () => {
  const { cart, loading } = useCart();
  const router = useRouter();
  const [currentStep, setCurrentStep] = useState<CheckoutStep>('shipping');
  const [checkoutData, setCheckoutData] = useState<CheckoutData>({
    shippingAddress: null,
    billingAddress: null,
    sameAsShipping: true,
    notes: '',
  });

  useEffect(() => {
    // Redirect to cart if cart is empty
    if (!loading && (!cart || cart.items.length === 0)) {
      router.push('/cart');
    }
  }, [cart, loading, router]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-4xl mx-auto px-4">
          <div className="animate-pulse">
            <div className="h-8 bg-gray-300 rounded w-48 mb-8"></div>
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

  if (!cart || cart.items.length === 0) {
    return null; // Will redirect in useEffect
  }

  const handleStepComplete = (step: CheckoutStep, data: Partial<CheckoutData>) => {
    setCheckoutData(prev => ({ ...prev, ...data }));
    
    // Move to next step
    if (step === 'shipping') {
      setCurrentStep('payment');
    } else if (step === 'payment') {
      setCurrentStep('review');
    }
  };

  const handleStepBack = () => {
    if (currentStep === 'payment') {
      setCurrentStep('shipping');
    } else if (currentStep === 'review') {
      setCurrentStep('payment');
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Checkout</h1>
        
        <CheckoutSteps currentStep={currentStep} />
        
        <div className="mt-8 grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-lg shadow">
              {currentStep === 'shipping' && (
                <ShippingForm
                  initialData={checkoutData}
                  onComplete={(data) => handleStepComplete('shipping', data)}
                />
              )}
              
              {currentStep === 'payment' && (
                <PaymentForm
                  checkoutData={checkoutData}
                  onComplete={(data) => handleStepComplete('payment', data)}
                  onBack={handleStepBack}
                />
              )}
              
              {currentStep === 'review' && (
                <OrderReview
                  checkoutData={checkoutData}
                  cart={cart}
                  onBack={handleStepBack}
                />
              )}
            </div>
          </div>

          {/* Order Summary Sidebar */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow p-6 sticky top-8">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Order Summary</h2>
              
              <div className="space-y-3 mb-4">
                {cart.items.map((item) => (
                  <div key={item.productId} className="flex justify-between text-sm">
                    <div className="flex-1">
                      <span className="font-medium">{item.product?.name || 'Product'}</span>
                      <span className="text-gray-500 ml-2">× {item.quantity}</span>
                    </div>
                    <span className="font-medium">₹{item.total.toFixed(2)}</span>
                  </div>
                ))}
              </div>
              
              <div className="border-t pt-3 space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">Subtotal</span>
                  <span>₹{cart.subtotal.toFixed(2)}</span>
                </div>
                
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">Tax</span>
                  <span>₹{cart.tax.toFixed(2)}</span>
                </div>
                
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">Shipping</span>
                  <span>Free</span>
                </div>
                
                <div className="border-t pt-2">
                  <div className="flex justify-between font-semibold">
                    <span>Total</span>
                    <span>₹{cart.total.toFixed(2)}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CheckoutPage;