import { test, expect } from '@playwright/test';

test.describe('Complete Shopping Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Start from the homepage
    await page.goto('/');
  });

  test('user can browse products, add to cart, and complete purchase', async ({ page }) => {
    // Browse products on homepage
    await expect(page.locator('h1')).toContainText('Welcome');
    await expect(page.locator('[data-testid="featured-products"]')).toBeVisible();

    // Navigate to products page
    await page.click('text=Shop Now');
    await expect(page).toHaveURL('/products');

    // Verify products are displayed
    await expect(page.locator('[data-testid="product-grid"]')).toBeVisible();
    await expect(page.locator('[data-testid="product-card"]').first()).toBeVisible();

    // Click on first product
    await page.locator('[data-testid="product-card"]').first().click();
    await expect(page).toHaveURL(/\/products\/[^\/]+$/);

    // Verify product details page
    await expect(page.locator('[data-testid="product-name"]')).toBeVisible();
    await expect(page.locator('[data-testid="product-price"]')).toBeVisible();
    await expect(page.locator('[data-testid="product-description"]')).toBeVisible();

    // Add product to cart
    await page.click('button:has-text("Add to Cart")');
    await expect(page.locator('text=Added to cart')).toBeVisible();

    // Verify cart icon shows item count
    await expect(page.locator('[data-testid="cart-count"]')).toContainText('1');

    // Go to cart
    await page.click('[data-testid="cart-icon"]');
    await expect(page).toHaveURL('/cart');

    // Verify cart contents
    await expect(page.locator('[data-testid="cart-item"]')).toBeVisible();
    await expect(page.locator('[data-testid="cart-total"]')).toBeVisible();

    // Update quantity
    await page.click('[data-testid="quantity-increase"]');
    await expect(page.locator('[data-testid="item-quantity"]')).toHaveValue('2');

    // Proceed to checkout
    await page.click('button:has-text("Proceed to Checkout")');
    
    // Should redirect to login if not authenticated
    await expect(page).toHaveURL('/auth/login');

    // Login first
    await page.fill('input[name="email"]', 'test@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');

    // Should redirect back to checkout
    await expect(page).toHaveURL('/checkout');

    // Fill shipping information
    await page.fill('input[name="firstName"]', 'John');
    await page.fill('input[name="lastName"]', 'Doe');
    await page.fill('input[name="address"]', '123 Main St');
    await page.fill('input[name="city"]', 'Anytown');
    await page.fill('input[name="state"]', 'CA');
    await page.fill('input[name="zipCode"]', '12345');
    await page.fill('input[name="phone"]', '555-1234');

    // Continue to payment
    await page.click('button:has-text("Continue to Payment")');

    // Verify payment section is visible
    await expect(page.locator('[data-testid="payment-section"]')).toBeVisible();
    await expect(page.locator('[data-testid="order-summary"]')).toBeVisible();

    // Fill payment information (test mode)
    await page.fill('input[name="cardNumber"]', '4111111111111111');
    await page.fill('input[name="expiryDate"]', '12/25');
    await page.fill('input[name="cvv"]', '123');
    await page.fill('input[name="cardName"]', 'John Doe');

    // Place order
    await page.click('button:has-text("Place Order")');

    // Should redirect to order confirmation
    await expect(page).toHaveURL(/\/order-confirmation\/[^\/]+$/);
    await expect(page.locator('text=Order Confirmed')).toBeVisible();
    await expect(page.locator('[data-testid="order-number"]')).toBeVisible();
  });

  test('user can search and filter products', async ({ page }) => {
    await page.goto('/products');

    // Test search functionality
    await page.fill('[data-testid="search-input"]', 'laptop');
    await page.press('[data-testid="search-input"]', 'Enter');

    // Verify search results
    await expect(page).toHaveURL('/products?search=laptop');
    await expect(page.locator('[data-testid="search-results"]')).toBeVisible();

    // Test category filter
    await page.click('[data-testid="category-filter"]');
    await page.click('text=Electronics');
    await expect(page).toHaveURL(/category=electronics/);

    // Test price filter
    await page.fill('[data-testid="min-price"]', '100');
    await page.fill('[data-testid="max-price"]', '500');
    await page.click('button:has-text("Apply Filters")');

    // Verify filtered results
    await expect(page).toHaveURL(/min_price=100.*max_price=500/);

    // Test sorting
    await page.selectOption('[data-testid="sort-select"]', 'price-asc');
    await expect(page).toHaveURL(/sort=price.*order=asc/);
  });

  test('cart persistence and management', async ({ page }) => {
    // Add item to cart
    await page.goto('/products');
    await page.locator('[data-testid="product-card"]').first().click();
    await page.click('button:has-text("Add to Cart")');

    // Navigate away and back
    await page.goto('/');
    await page.goto('/cart');

    // Verify item is still in cart
    await expect(page.locator('[data-testid="cart-item"]')).toBeVisible();

    // Test quantity update
    await page.click('[data-testid="quantity-increase"]');
    await expect(page.locator('[data-testid="item-quantity"]')).toHaveValue('2');

    // Test item removal
    await page.click('[data-testid="remove-item"]');
    await expect(page.locator('text=Your cart is empty')).toBeVisible();
  });

  test('guest checkout flow', async ({ page }) => {
    // Add item to cart
    await page.goto('/products');
    await page.locator('[data-testid="product-card"]').first().click();
    await page.click('button:has-text("Add to Cart")');

    // Go to checkout
    await page.goto('/cart');
    await page.click('button:has-text("Proceed to Checkout")');

    // Choose guest checkout
    await page.click('button:has-text("Continue as Guest")');

    // Fill guest information
    await page.fill('input[name="email"]', 'guest@example.com');
    await page.fill('input[name="firstName"]', 'Guest');
    await page.fill('input[name="lastName"]', 'User');
    await page.fill('input[name="address"]', '456 Guest St');
    await page.fill('input[name="city"]', 'Guesttown');
    await page.fill('input[name="state"]', 'NY');
    await page.fill('input[name="zipCode"]', '54321');
    await page.fill('input[name="phone"]', '555-5678');

    // Continue to payment
    await page.click('button:has-text("Continue to Payment")');

    // Verify guest checkout works
    await expect(page.locator('[data-testid="payment-section"]')).toBeVisible();
  });

  test('product availability and inventory', async ({ page }) => {
    await page.goto('/products');
    
    // Find a product with low inventory
    const lowStockProduct = page.locator('[data-testid="product-card"]:has-text("Low Stock")').first();
    if (await lowStockProduct.isVisible()) {
      await lowStockProduct.click();
      
      // Verify low stock warning
      await expect(page.locator('text=Only')).toBeVisible();
      await expect(page.locator('text=left in stock')).toBeVisible();
    }

    // Find an out of stock product
    const outOfStockProduct = page.locator('[data-testid="product-card"]:has-text("Out of Stock")').first();
    if (await outOfStockProduct.isVisible()) {
      await outOfStockProduct.click();
      
      // Verify add to cart button is disabled
      await expect(page.locator('button:has-text("Add to Cart")')).toBeDisabled();
      await expect(page.locator('text=Out of Stock')).toBeVisible();
    }
  });
});