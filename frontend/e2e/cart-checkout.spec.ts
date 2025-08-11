import { test, expect } from '@playwright/test';

test.describe('Shopping Cart and Checkout Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the homepage
    await page.goto('/');
    
    // Mock Razorpay script loading
    await page.addInitScript(() => {
      // Mock Razorpay constructor
      (window as any).Razorpay = class MockRazorpay {
        constructor(options: any) {
          this.options = options;
        }
        
        open() {
          // Simulate successful payment
          setTimeout(() => {
            if (this.options.handler) {
              this.options.handler({
                razorpay_order_id: 'order_mock_123',
                razorpay_payment_id: 'pay_mock_123',
                razorpay_signature: 'signature_mock_123'
              });
            }
          }, 100);
        }
      };
    });
  });

  test('should add product to cart and view cart page', async ({ page }) => {
    // Navigate to products page
    await page.waitForSelector('text=Shop Now', { timeout: 10000 });
    await page.click('text=Shop Now');
    await page.waitForURL('/products', { timeout: 10000 });

    // Wait for products to load
    await page.waitForSelector('[data-testid="product-card"]', { timeout: 10000 });

    // Click on the first product
    const firstProduct = page.locator('[data-testid="product-card"]').first();
    await firstProduct.click();

    // Add product to cart
    await page.waitForSelector('text=Add to Cart', { timeout: 5000 });
    await page.click('text=Add to Cart');
    
    // Wait for success message or cart update
    await page.waitForTimeout(2000);

    // Navigate to cart
    await page.click('[data-testid="cart-link"]');
    await page.waitForURL('/cart', { timeout: 10000 });

    // Verify cart has items
    await expect(page.locator('text=Cart Items')).toBeVisible();
    await expect(page.locator('[data-testid="cart-item"]')).toHaveCount(1);
  });

  test('should update item quantity in cart', async ({ page }) => {
    // First add a product to cart (assuming we have a helper or setup)
    await page.goto('/products');
    await page.waitForSelector('[data-testid="product-card"]', { timeout: 10000 });
    
    const firstProduct = page.locator('[data-testid="product-card"]').first();
    await firstProduct.click();
    await page.click('text=Add to Cart');
    await page.waitForTimeout(1000);

    // Go to cart
    await page.click('[data-testid="cart-link"]');
    
    // Find quantity controls
    const increaseButton = page.locator('button:has(svg)').filter({ hasText: '+' }).first();
    const quantityDisplay = page.locator('text=/^\\d+$/', { hasText: '1' });
    
    // Increase quantity
    await increaseButton.click();
    await page.waitForTimeout(500);
    
    // Verify quantity updated
    await expect(page.locator('text=2')).toBeVisible();
  });

  test('should remove item from cart', async ({ page }) => {
    // Add product to cart first
    await page.goto('/products');
    await page.waitForSelector('[data-testid="product-card"]', { timeout: 10000 });
    
    const firstProduct = page.locator('[data-testid="product-card"]').first();
    await firstProduct.click();
    await page.click('text=Add to Cart');
    await page.waitForTimeout(1000);

    // Go to cart
    await page.click('[data-testid="cart-link"]');
    
    // Remove item
    await page.click('text=Remove');
    await page.waitForTimeout(500);
    
    // Verify cart is empty
    await expect(page.locator('text=Your cart is empty')).toBeVisible();
  });

  test('should proceed through checkout steps', async ({ page }) => {
    // Add product to cart first
    await page.goto('/products');
    await page.waitForSelector('[data-testid="product-card"]', { timeout: 10000 });
    
    const firstProduct = page.locator('[data-testid="product-card"]').first();
    await firstProduct.click();
    await page.click('text=Add to Cart');
    await page.waitForTimeout(1000);

    // Go to cart and proceed to checkout
    await page.click('[data-testid="cart-link"]');
    await page.click('text=Proceed to Checkout');
    await expect(page).toHaveURL('/checkout');

    // Fill shipping form
    await page.fill('input[name="firstName"]', 'John');
    await page.fill('input[name="lastName"]', 'Doe');
    await page.fill('input[name="address1"]', '123 Main Street');
    await page.fill('input[name="city"]', 'Mumbai');
    await page.fill('input[name="state"]', 'Maharashtra');
    await page.fill('input[name="postalCode"]', '400001');
    await page.fill('input[name="country"]', 'India');
    await page.fill('input[name="phone"]', '+91 9876543210');

    // Continue to payment
    await page.click('text=Continue to Payment');
    await page.waitForTimeout(500);

    // Verify we're on payment step
    await expect(page.locator('text=Payment Method')).toBeVisible();
    await expect(page.locator('text=Razorpay')).toBeVisible();

    // Continue to review
    await page.click('text=Review Order');
    await page.waitForTimeout(500);

    // Verify we're on review step
    await expect(page.locator('text=Review Your Order')).toBeVisible();
    await expect(page.locator('text=Order Items')).toBeVisible();
    await expect(page.locator('text=Shipping Address')).toBeVisible();
    await expect(page.locator('text=John Doe')).toBeVisible();
  });

  test('should validate required fields in shipping form', async ({ page }) => {
    // Add product to cart first
    await page.goto('/products');
    await page.waitForSelector('[data-testid="product-card"]', { timeout: 10000 });
    
    const firstProduct = page.locator('[data-testid="product-card"]').first();
    await firstProduct.click();
    await page.click('text=Add to Cart');
    await page.waitForTimeout(1000);

    // Go to checkout
    await page.click('[data-testid="cart-link"]');
    await page.click('text=Proceed to Checkout');

    // Try to submit without filling required fields
    await page.click('text=Continue to Payment');
    await page.waitForTimeout(500);

    // Verify validation errors appear
    await expect(page.locator('text=First name is required')).toBeVisible();
    await expect(page.locator('text=Last name is required')).toBeVisible();
    await expect(page.locator('text=Address is required')).toBeVisible();
  });

  test('should handle empty cart checkout redirect', async ({ page }) => {
    // Try to access checkout with empty cart
    await page.goto('/checkout');
    
    // Should redirect to cart page
    await expect(page).toHaveURL('/cart');
    await expect(page.locator('text=Your cart is empty')).toBeVisible();
  });

  test('should display order summary correctly', async ({ page }) => {
    // Add product to cart
    await page.goto('/products');
    await page.waitForSelector('[data-testid="product-card"]', { timeout: 10000 });
    
    const firstProduct = page.locator('[data-testid="product-card"]').first();
    await firstProduct.click();
    await page.click('text=Add to Cart');
    await page.waitForTimeout(1000);

    // Go to cart
    await page.click('[data-testid="cart-link"]');
    
    // Verify order summary elements
    await expect(page.locator('text=Order Summary')).toBeVisible();
    await expect(page.locator('text=Subtotal')).toBeVisible();
    await expect(page.locator('text=Tax')).toBeVisible();
    await expect(page.locator('text=Total')).toBeVisible();
    
    // Verify price format (₹ symbol)
    await expect(page.locator('text=/₹\\d+\\.\\d{2}/')).toBeVisible();
  });

  test('should handle billing address same as shipping', async ({ page }) => {
    // Add product and go to checkout
    await page.goto('/products');
    await page.waitForSelector('[data-testid="product-card"]', { timeout: 10000 });
    
    const firstProduct = page.locator('[data-testid="product-card"]').first();
    await firstProduct.click();
    await page.click('text=Add to Cart');
    await page.waitForTimeout(1000);

    await page.click('[data-testid="cart-link"]');
    await page.click('text=Proceed to Checkout');

    // Fill shipping address
    await page.fill('input[name="firstName"]', 'Jane');
    await page.fill('input[name="lastName"]', 'Smith');
    await page.fill('input[name="address1"]', '456 Oak Avenue');
    await page.fill('input[name="city"]', 'Delhi');
    await page.fill('input[name="state"]', 'Delhi');
    await page.fill('input[name="postalCode"]', '110001');
    await page.fill('input[name="country"]', 'India');

    // Verify "same as shipping" is checked by default
    const sameAsShippingCheckbox = page.locator('input[type="checkbox"]');
    await expect(sameAsShippingCheckbox).toBeChecked();

    // Uncheck to show billing form
    await sameAsShippingCheckbox.uncheck();
    await page.waitForTimeout(500);

    // Verify billing address form appears
    await expect(page.locator('text=Billing Address')).toBeVisible();
    
    // Check again to hide billing form
    await sameAsShippingCheckbox.check();
    await page.waitForTimeout(500);

    // Continue and verify in review that billing shows "Same as shipping"
    await page.click('text=Continue to Payment');
    await page.click('text=Review Order');
    
    await expect(page.locator('text=Same as shipping address')).toBeVisible();
  });

  test('should complete full purchase flow with payment', async ({ page }) => {
    // Mock API responses for order creation and payment
    await page.route('**/api/orders/create', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 'order_test_123',
            userId: 'user_test_123',
            status: 'pending',
            subtotal: 999.00,
            tax: 179.82,
            shipping: 0,
            total: 1178.82,
            shippingAddress: {
              firstName: 'John',
              lastName: 'Doe',
              address1: '123 Main Street',
              city: 'Mumbai',
              state: 'Maharashtra',
              postalCode: '400001',
              country: 'India',
              phone: '+91 9876543210'
            },
            billingAddress: {
              firstName: 'John',
              lastName: 'Doe',
              address1: '123 Main Street',
              city: 'Mumbai',
              state: 'Maharashtra',
              postalCode: '400001',
              country: 'India',
              phone: '+91 9876543210'
            },
            paymentIntentId: '',
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
          }
        })
      });
    });

    await page.route('**/api/payments/create-order', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 'payment_test_123',
            orderId: 'order_test_123',
            razorpayOrderId: 'order_razorpay_123',
            amount: 117882,
            currency: 'INR',
            status: 'created',
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
          }
        })
      });
    });

    await page.route('**/api/payments/verify', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: { verified: true }
        })
      });
    });

    await page.route('**/api/orders/order_test_123', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 'order_test_123',
            userId: 'user_test_123',
            status: 'confirmed',
            subtotal: 999.00,
            tax: 179.82,
            shipping: 0,
            total: 1178.82,
            shippingAddress: {
              firstName: 'John',
              lastName: 'Doe',
              address1: '123 Main Street',
              city: 'Mumbai',
              state: 'Maharashtra',
              postalCode: '400001',
              country: 'India',
              phone: '+91 9876543210'
            },
            billingAddress: {
              firstName: 'John',
              lastName: 'Doe',
              address1: '123 Main Street',
              city: 'Mumbai',
              state: 'Maharashtra',
              postalCode: '400001',
              country: 'India',
              phone: '+91 9876543210'
            },
            paymentIntentId: 'pay_mock_123',
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            items: [{
              id: 'item_test_123',
              orderId: 'order_test_123',
              productId: 'product_test_123',
              quantity: 1,
              price: 999.00,
              total: 999.00,
              product: {
                id: 'product_test_123',
                name: 'Test Product',
                sku: 'TEST-001'
              }
            }]
          }
        })
      });
    });

    await page.route('**/api/cart/clear', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: null
        })
      });
    });

    // Add product to cart
    await page.goto('/products');
    await page.waitForSelector('[data-testid="product-card"]', { timeout: 10000 });
    
    const firstProduct = page.locator('[data-testid="product-card"]').first();
    await firstProduct.click();
    await page.click('text=Add to Cart');
    await page.waitForTimeout(1000);

    // Go to cart and proceed to checkout
    await page.click('[data-testid="cart-link"]');
    await page.click('text=Proceed to Checkout');

    // Fill shipping form
    await page.fill('input[name="firstName"]', 'John');
    await page.fill('input[name="lastName"]', 'Doe');
    await page.fill('input[name="address1"]', '123 Main Street');
    await page.fill('input[name="city"]', 'Mumbai');
    await page.fill('input[name="state"]', 'Maharashtra');
    await page.fill('input[name="postalCode"]', '400001');
    await page.fill('input[name="country"]', 'India');
    await page.fill('input[name="phone"]', '+91 9876543210');

    // Continue to payment
    await page.click('text=Continue to Payment');
    await page.waitForTimeout(500);

    // Add order notes
    await page.fill('textarea[id="notes"]', 'Please handle with care');

    // Continue to review
    await page.click('text=Review Order');
    await page.waitForTimeout(500);

    // Verify order review details
    await expect(page.locator('text=Review Your Order')).toBeVisible();
    await expect(page.locator('text=John Doe')).toBeVisible();
    await expect(page.locator('text=123 Main Street')).toBeVisible();
    await expect(page.locator('text=Please handle with care')).toBeVisible();

    // Place order (this will trigger Razorpay mock)
    await page.click('text=Place Order');
    await page.waitForTimeout(1000);

    // Should redirect to order confirmation page
    await expect(page).toHaveURL(/\/order-confirmation\/order_test_123/);
    await expect(page.locator('text=Order Confirmed!')).toBeVisible();
    await expect(page.locator('text=#order_test_123')).toBeVisible();
  });

  test('should display cart count in header', async ({ page }) => {
    // Add product to cart
    await page.goto('/products');
    await page.waitForSelector('[data-testid="product-card"]', { timeout: 10000 });
    
    const firstProduct = page.locator('[data-testid="product-card"]').first();
    await firstProduct.click();
    await page.click('text=Add to Cart');
    await page.waitForTimeout(1000);

    // Check cart count in header
    const cartLink = page.locator('[data-testid="cart-link"]');
    await expect(cartLink).toBeVisible();
    
    // Should show cart count badge
    await expect(page.locator('[data-testid="cart-link"] span')).toHaveText('1');
  });

  test('should handle payment failure gracefully', async ({ page }) => {
    // Mock payment failure
    await page.addInitScript(() => {
      (window as any).Razorpay = class MockRazorpay {
        constructor(options: any) {
          this.options = options;
        }
        
        open() {
          // Simulate payment failure by calling ondismiss
          setTimeout(() => {
            if (this.options.modal && this.options.modal.ondismiss) {
              this.options.modal.ondismiss();
            }
          }, 100);
        }
      };
    });

    // Mock API responses
    await page.route('**/api/orders/create', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 'order_test_123',
            status: 'pending',
            total: 1178.82
          }
        })
      });
    });

    await page.route('**/api/payments/create-order', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            razorpayOrderId: 'order_razorpay_123',
            amount: 117882,
            currency: 'INR'
          }
        })
      });
    });

    // Complete checkout flow
    await page.goto('/products');
    await page.waitForSelector('[data-testid="product-card"]', { timeout: 10000 });
    
    const firstProduct = page.locator('[data-testid="product-card"]').first();
    await firstProduct.click();
    await page.click('text=Add to Cart');
    await page.waitForTimeout(1000);

    await page.click('[data-testid="cart-link"]');
    await page.click('text=Proceed to Checkout');

    // Fill shipping form
    await page.fill('input[name="firstName"]', 'John');
    await page.fill('input[name="lastName"]', 'Doe');
    await page.fill('input[name="address1"]', '123 Main Street');
    await page.fill('input[name="city"]', 'Mumbai');
    await page.fill('input[name="state"]', 'Maharashtra');
    await page.fill('input[name="postalCode"]', '400001');
    await page.fill('input[name="country"]', 'India');

    await page.click('text=Continue to Payment');
    await page.click('text=Review Order');

    // Place order (payment will fail)
    await page.click('text=Place Order');
    await page.waitForTimeout(1000);

    // Should remain on checkout page and show Place Order button again
    await expect(page).toHaveURL('/checkout');
    await expect(page.locator('text=Place Order')).toBeVisible();
  });
});