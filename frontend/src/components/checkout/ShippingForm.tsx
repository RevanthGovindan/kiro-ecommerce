import React, { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { OrderAddress } from '@/lib/api';
import { CheckoutData } from '@/app/checkout/page';

interface ShippingFormProps {
  initialData: CheckoutData;
  onComplete: (data: Partial<CheckoutData>) => void;
}

interface FormErrors {
  [key: string]: string;
}

const initialAddress: OrderAddress = {
  firstName: '',
  lastName: '',
  company: '',
  address1: '',
  address2: '',
  city: '',
  state: '',
  postalCode: '',
  country: 'India',
  phone: '',
};

export const ShippingForm: React.FC<ShippingFormProps> = ({ initialData, onComplete }) => {
  const [shippingAddress, setShippingAddress] = useState<OrderAddress>(
    initialData.shippingAddress || initialAddress
  );
  const [billingAddress, setBillingAddress] = useState<OrderAddress>(
    initialData.billingAddress || initialAddress
  );
  const [sameAsShipping, setSameAsShipping] = useState(initialData.sameAsShipping);
  const [errors, setErrors] = useState<FormErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const validateAddress = (address: OrderAddress, prefix: string = ''): FormErrors => {
    const addressErrors: FormErrors = {};
    
    if (!address.firstName.trim()) {
      addressErrors[`${prefix}firstName`] = 'First name is required';
    }
    
    if (!address.lastName.trim()) {
      addressErrors[`${prefix}lastName`] = 'Last name is required';
    }
    
    if (!address.address1.trim()) {
      addressErrors[`${prefix}address1`] = 'Address is required';
    }
    
    if (!address.city.trim()) {
      addressErrors[`${prefix}city`] = 'City is required';
    }
    
    if (!address.state.trim()) {
      addressErrors[`${prefix}state`] = 'State is required';
    }
    
    if (!address.postalCode.trim()) {
      addressErrors[`${prefix}postalCode`] = 'Postal code is required';
    } else if (!/^\d{6}$/.test(address.postalCode)) {
      addressErrors[`${prefix}postalCode`] = 'Please enter a valid 6-digit postal code';
    }
    
    if (!address.country.trim()) {
      addressErrors[`${prefix}country`] = 'Country is required';
    }
    
    if (address.phone && !/^\+?[\d\s\-\(\)]{10,}$/.test(address.phone)) {
      addressErrors[`${prefix}phone`] = 'Please enter a valid phone number';
    }
    
    return addressErrors;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    
    // Validate shipping address
    const shippingErrors = validateAddress(shippingAddress, 'shipping_');
    let allErrors = { ...shippingErrors };
    
    // Validate billing address if different from shipping
    if (!sameAsShipping) {
      const billingErrors = validateAddress(billingAddress, 'billing_');
      allErrors = { ...allErrors, ...billingErrors };
    }
    
    setErrors(allErrors);
    
    if (Object.keys(allErrors).length === 0) {
      const finalBillingAddress = sameAsShipping ? shippingAddress : billingAddress;
      
      onComplete({
        shippingAddress,
        billingAddress: finalBillingAddress,
        sameAsShipping,
      });
    }
    
    setIsSubmitting(false);
  };

  const handleShippingChange = (field: keyof OrderAddress, value: string) => {
    setShippingAddress(prev => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[`shipping_${field}`]) {
      setErrors(prev => ({ ...prev, [`shipping_${field}`]: '' }));
    }
  };

  const handleBillingChange = (field: keyof OrderAddress, value: string) => {
    setBillingAddress(prev => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[`billing_${field}`]) {
      setErrors(prev => ({ ...prev, [`billing_${field}`]: '' }));
    }
  };

  const handleSameAsShippingChange = (checked: boolean) => {
    setSameAsShipping(checked);
    if (checked) {
      setBillingAddress(shippingAddress);
      // Clear billing errors
      const newErrors = { ...errors };
      Object.keys(newErrors).forEach(key => {
        if (key.startsWith('billing_')) {
          delete newErrors[key];
        }
      });
      setErrors(newErrors);
    }
  };

  return (
    <div className="p-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-6">Shipping Information</h2>
      
      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Shipping Address */}
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">Shipping Address</h3>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="First Name *"
              value={shippingAddress.firstName}
              onChange={(e) => handleShippingChange('firstName', e.target.value)}
              error={errors.shipping_firstName}
            />
            
            <Input
              label="Last Name *"
              value={shippingAddress.lastName}
              onChange={(e) => handleShippingChange('lastName', e.target.value)}
              error={errors.shipping_lastName}
            />
            
            <div className="md:col-span-2">
              <Input
                label="Company (Optional)"
                value={shippingAddress.company || ''}
                onChange={(e) => handleShippingChange('company', e.target.value)}
              />
            </div>
            
            <div className="md:col-span-2">
              <Input
                label="Address Line 1 *"
                value={shippingAddress.address1}
                onChange={(e) => handleShippingChange('address1', e.target.value)}
                error={errors.shipping_address1}
              />
            </div>
            
            <div className="md:col-span-2">
              <Input
                label="Address Line 2 (Optional)"
                value={shippingAddress.address2 || ''}
                onChange={(e) => handleShippingChange('address2', e.target.value)}
              />
            </div>
            
            <Input
              label="City *"
              value={shippingAddress.city}
              onChange={(e) => handleShippingChange('city', e.target.value)}
              error={errors.shipping_city}
            />
            
            <Input
              label="State *"
              value={shippingAddress.state}
              onChange={(e) => handleShippingChange('state', e.target.value)}
              error={errors.shipping_state}
            />
            
            <Input
              label="Postal Code *"
              value={shippingAddress.postalCode}
              onChange={(e) => handleShippingChange('postalCode', e.target.value)}
              error={errors.shipping_postalCode}
              placeholder="123456"
            />
            
            <Input
              label="Country *"
              value={shippingAddress.country}
              onChange={(e) => handleShippingChange('country', e.target.value)}
              error={errors.shipping_country}
            />
            
            <div className="md:col-span-2">
              <Input
                label="Phone (Optional)"
                value={shippingAddress.phone || ''}
                onChange={(e) => handleShippingChange('phone', e.target.value)}
                error={errors.shipping_phone}
                placeholder="+91 9876543210"
              />
            </div>
          </div>
        </div>

        {/* Billing Address */}
        <div>
          <div className="flex items-center mb-4">
            <input
              id="same-as-shipping"
              type="checkbox"
              checked={sameAsShipping}
              onChange={(e) => handleSameAsShippingChange(e.target.checked)}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            />
            <label htmlFor="same-as-shipping" className="ml-2 text-sm text-gray-700">
              Billing address is the same as shipping address
            </label>
          </div>

          {!sameAsShipping && (
            <>
              <h3 className="text-lg font-medium text-gray-900 mb-4">Billing Address</h3>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="First Name *"
                  value={billingAddress.firstName}
                  onChange={(e) => handleBillingChange('firstName', e.target.value)}
                  error={errors.billing_firstName}
                />
                
                <Input
                  label="Last Name *"
                  value={billingAddress.lastName}
                  onChange={(e) => handleBillingChange('lastName', e.target.value)}
                  error={errors.billing_lastName}
                />
                
                <div className="md:col-span-2">
                  <Input
                    label="Company (Optional)"
                    value={billingAddress.company || ''}
                    onChange={(e) => handleBillingChange('company', e.target.value)}
                  />
                </div>
                
                <div className="md:col-span-2">
                  <Input
                    label="Address Line 1 *"
                    value={billingAddress.address1}
                    onChange={(e) => handleBillingChange('address1', e.target.value)}
                    error={errors.billing_address1}
                  />
                </div>
                
                <div className="md:col-span-2">
                  <Input
                    label="Address Line 2 (Optional)"
                    value={billingAddress.address2 || ''}
                    onChange={(e) => handleBillingChange('address2', e.target.value)}
                  />
                </div>
                
                <Input
                  label="City *"
                  value={billingAddress.city}
                  onChange={(e) => handleBillingChange('city', e.target.value)}
                  error={errors.billing_city}
                />
                
                <Input
                  label="State *"
                  value={billingAddress.state}
                  onChange={(e) => handleBillingChange('state', e.target.value)}
                  error={errors.billing_state}
                />
                
                <Input
                  label="Postal Code *"
                  value={billingAddress.postalCode}
                  onChange={(e) => handleBillingChange('postalCode', e.target.value)}
                  error={errors.billing_postalCode}
                  placeholder="123456"
                />
                
                <Input
                  label="Country *"
                  value={billingAddress.country}
                  onChange={(e) => handleBillingChange('country', e.target.value)}
                  error={errors.billing_country}
                />
                
                <div className="md:col-span-2">
                  <Input
                    label="Phone (Optional)"
                    value={billingAddress.phone || ''}
                    onChange={(e) => handleBillingChange('phone', e.target.value)}
                    error={errors.billing_phone}
                    placeholder="+91 9876543210"
                  />
                </div>
              </div>
            </>
          )}
        </div>

        {/* Submit Button */}
        <div className="flex justify-end pt-6 border-t">
          <Button
            type="submit"
            size="lg"
            disabled={isSubmitting}
          >
            {isSubmitting ? 'Validating...' : 'Continue to Payment'}
          </Button>
        </div>
      </form>
    </div>
  );
};