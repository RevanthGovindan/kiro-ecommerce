import React, { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { CheckoutData } from '@/app/checkout/page';

interface PaymentFormProps {
  checkoutData: CheckoutData;
  onComplete: (data: Partial<CheckoutData>) => void;
  onBack: () => void;
}

export const PaymentForm: React.FC<PaymentFormProps> = ({ checkoutData, onComplete, onBack }) => {
  const [notes, setNotes] = useState(checkoutData.notes || '');
  const [isProcessing, setIsProcessing] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsProcessing(true);
    
    try {
      // Save notes and proceed to review
      onComplete({ notes });
    } catch (error) {
      console.error('Error processing payment form:', error);
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <div className="p-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-6">Payment Method</h2>
      
      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Payment Method Selection */}
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">Select Payment Method</h3>
          
          <div className="space-y-3">
            {/* Razorpay Option */}
            <div className="border border-gray-200 rounded-lg p-4">
              <div className="flex items-center">
                <input
                  id="razorpay"
                  name="payment-method"
                  type="radio"
                  defaultChecked
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                />
                <label htmlFor="razorpay" className="ml-3 flex items-center">
                  <div>
                    <div className="text-sm font-medium text-gray-900">Razorpay</div>
                    <div className="text-sm text-gray-500">
                      Pay securely with credit card, debit card, UPI, or net banking
                    </div>
                  </div>
                </label>
              </div>
              
              <div className="mt-3 flex items-center space-x-2">
                <div className="flex items-center space-x-1">
                  <div className="w-8 h-5 bg-blue-600 rounded text-white text-xs flex items-center justify-center font-bold">
                    VISA
                  </div>
                  <div className="w-8 h-5 bg-red-600 rounded text-white text-xs flex items-center justify-center font-bold">
                    MC
                  </div>
                  <div className="w-8 h-5 bg-orange-500 rounded text-white text-xs flex items-center justify-center font-bold">
                    UPI
                  </div>
                </div>
                <span className="text-xs text-gray-500">and more</span>
              </div>
            </div>

            {/* Cash on Delivery (Disabled for now) */}
            <div className="border border-gray-200 rounded-lg p-4 opacity-50">
              <div className="flex items-center">
                <input
                  id="cod"
                  name="payment-method"
                  type="radio"
                  disabled
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                />
                <label htmlFor="cod" className="ml-3">
                  <div className="text-sm font-medium text-gray-900">Cash on Delivery</div>
                  <div className="text-sm text-gray-500">
                    Pay when your order is delivered (Currently unavailable)
                  </div>
                </label>
              </div>
            </div>
          </div>
        </div>

        {/* Order Notes */}
        <div>
          <label htmlFor="notes" className="block text-sm font-medium text-gray-700 mb-2">
            Order Notes (Optional)
          </label>
          <textarea
            id="notes"
            rows={3}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            placeholder="Any special instructions for your order..."
          />
        </div>

        {/* Security Notice */}
        <div className="bg-green-50 border border-green-200 rounded-lg p-4">
          <div className="flex">
            <svg className="h-5 w-5 text-green-400 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
            <div className="ml-3">
              <h4 className="text-sm font-medium text-green-800">Secure Payment</h4>
              <p className="text-sm text-green-700 mt-1">
                Your payment information is encrypted and secure. We never store your card details.
              </p>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex justify-between pt-6 border-t">
          <Button
            type="button"
            variant="outline"
            onClick={onBack}
            disabled={isProcessing}
          >
            Back to Shipping
          </Button>
          
          <Button
            type="submit"
            size="lg"
            disabled={isProcessing}
          >
            {isProcessing ? 'Processing...' : 'Review Order'}
          </Button>
        </div>
      </form>
    </div>
  );
};