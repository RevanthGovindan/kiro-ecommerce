import { test, expect } from '@playwright/test';

test.describe('Admin Dashboard Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin user
    await page.goto('/auth/login');
    await page.fill('input[name="email"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'adminpassword123');
    await page.click('button[type="submit"]');
    
    // Navigate to admin dashboard
    await page.goto('/admin');
  });

  test('admin can manage products', async ({ page }) => {
    // Verify admin dashboard is accessible
    await expect(page).toHaveURL('/admin');
    await expect(page.locator('h1:has-text("Admin Dashboard")')).toBeVisible();

    // Navigate to product management
    await page.click('text=Products');
    await expect(page).toHaveURL('/admin/products');

    // Verify product list is displayed
    await expect(page.locator('[data-testid="products-table"]')).toBeVisible();

    // Add new product
    await page.click('button:has-text("Add Product")');
    await expect(page.locator('[data-testid="product-form"]')).toBeVisible();

    // Fill product form
    await page.fill('input[name="name"]', 'Test Product');
    await page.fill('textarea[name="description"]', 'This is a test product description');
    await page.fill('input[name="price"]', '99.99');
    await page.fill('input[name="sku"]', 'TEST-001');
    await page.fill('input[name="inventory"]', '50');
    await page.selectOption('select[name="categoryId"]', { label: 'Electronics' });

    // Upload product image (if file input exists)
    const fileInput = page.locator('input[type="file"]');
    if (await fileInput.isVisible()) {
      await fileInput.setInputFiles('test-fixtures/product-image.jpg');
    }

    // Save product
    await page.click('button:has-text("Save Product")');

    // Verify product was created
    await expect(page.locator('text=Product created successfully')).toBeVisible();
    await expect(page.locator('text=Test Product')).toBeVisible();

    // Edit the product
    await page.click('[data-testid="edit-product"]:has-text("Test Product")');
    await page.fill('input[name="price"]', '89.99');
    await page.click('button:has-text("Update Product")');

    // Verify product was updated
    await expect(page.locator('text=Product updated successfully')).toBeVisible();
    await expect(page.locator('text=$89.99')).toBeVisible();

    // Delete the product
    await page.click('[data-testid="delete-product"]:has-text("Test Product")');
    await page.click('button:has-text("Confirm Delete")');

    // Verify product was deleted
    await expect(page.locator('text=Product deleted successfully')).toBeVisible();
    await expect(page.locator('text=Test Product')).not.toBeVisible();
  });

  test('admin can manage orders', async ({ page }) => {
    // Navigate to order management
    await page.click('text=Orders');
    await expect(page).toHaveURL('/admin/orders');

    // Verify orders table is displayed
    await expect(page.locator('[data-testid="orders-table"]')).toBeVisible();

    // Filter orders by status
    await page.selectOption('[data-testid="status-filter"]', 'pending');
    await expect(page).toHaveURL(/status=pending/);

    // View order details
    const firstOrder = page.locator('[data-testid="order-row"]').first();
    await firstOrder.click();

    // Verify order details modal/page
    await expect(page.locator('[data-testid="order-details"]')).toBeVisible();
    await expect(page.locator('[data-testid="order-items"]')).toBeVisible();
    await expect(page.locator('[data-testid="customer-info"]')).toBeVisible();

    // Update order status
    await page.selectOption('[data-testid="order-status-select"]', 'processing');
    await page.click('button:has-text("Update Status")');

    // Verify status was updated
    await expect(page.locator('text=Order status updated')).toBeVisible();
    await expect(page.locator('text=Processing')).toBeVisible();

    // Add order note
    await page.fill('[data-testid="order-note"]', 'Order is being prepared for shipment');
    await page.click('button:has-text("Add Note")');

    // Verify note was added
    await expect(page.locator('text=Note added successfully')).toBeVisible();
  });

  test('admin can manage customers', async ({ page }) => {
    // Navigate to customer management
    await page.click('text=Customers');
    await expect(page).toHaveURL('/admin/customers');

    // Verify customers table is displayed
    await expect(page.locator('[data-testid="customers-table"]')).toBeVisible();

    // Search for customer
    await page.fill('[data-testid="customer-search"]', 'john@example.com');
    await page.press('[data-testid="customer-search"]', 'Enter');

    // View customer details
    const firstCustomer = page.locator('[data-testid="customer-row"]').first();
    await firstCustomer.click();

    // Verify customer details
    await expect(page.locator('[data-testid="customer-details"]')).toBeVisible();
    await expect(page.locator('[data-testid="customer-orders"]')).toBeVisible();

    // Deactivate customer account
    await page.click('button:has-text("Deactivate Account")');
    await page.click('button:has-text("Confirm")');

    // Verify account was deactivated
    await expect(page.locator('text=Account deactivated')).toBeVisible();
    await expect(page.locator('text=Inactive')).toBeVisible();
  });

  test('admin can view analytics dashboard', async ({ page }) => {
    // Navigate to analytics
    await page.click('text=Analytics');
    await expect(page).toHaveURL('/admin/analytics');

    // Verify analytics widgets are displayed
    await expect(page.locator('[data-testid="total-sales"]')).toBeVisible();
    await expect(page.locator('[data-testid="total-orders"]')).toBeVisible();
    await expect(page.locator('[data-testid="total-customers"]')).toBeVisible();
    await expect(page.locator('[data-testid="conversion-rate"]')).toBeVisible();

    // Verify charts are displayed
    await expect(page.locator('[data-testid="sales-chart"]')).toBeVisible();
    await expect(page.locator('[data-testid="orders-chart"]')).toBeVisible();

    // Test date range filter
    await page.click('[data-testid="date-range-picker"]');
    await page.click('text=Last 30 days');

    // Verify data updates
    await expect(page.locator('[data-testid="loading-indicator"]')).toBeVisible();
    await expect(page.locator('[data-testid="loading-indicator"]')).not.toBeVisible();

    // Export report
    await page.click('button:has-text("Export Report")');
    
    // Verify download started (check for download event)
    const downloadPromise = page.waitForEvent('download');
    await downloadPromise;
  });

  test('admin can manage categories', async ({ page }) => {
    // Navigate to categories (if available)
    await page.click('text=Categories');
    await expect(page).toHaveURL('/admin/categories');

    // Add new category
    await page.click('button:has-text("Add Category")');
    await page.fill('input[name="name"]', 'Test Category');
    await page.fill('input[name="slug"]', 'test-category');
    await page.fill('textarea[name="description"]', 'Test category description');
    await page.click('button:has-text("Save Category")');

    // Verify category was created
    await expect(page.locator('text=Category created successfully')).toBeVisible();
    await expect(page.locator('text=Test Category')).toBeVisible();

    // Edit category
    await page.click('[data-testid="edit-category"]:has-text("Test Category")');
    await page.fill('textarea[name="description"]', 'Updated category description');
    await page.click('button:has-text("Update Category")');

    // Verify category was updated
    await expect(page.locator('text=Category updated successfully')).toBeVisible();
  });

  test('admin permissions and access control', async ({ page }) => {
    // Verify admin-only sections are accessible
    await expect(page.locator('text=Admin Dashboard')).toBeVisible();
    await expect(page.locator('text=Products')).toBeVisible();
    await expect(page.locator('text=Orders')).toBeVisible();
    await expect(page.locator('text=Customers')).toBeVisible();

    // Test logout
    await page.click('button:has-text("Logout")');
    await expect(page).toHaveURL('/');

    // Try to access admin area after logout
    await page.goto('/admin');
    await expect(page).toHaveURL('/auth/login');
  });

  test('admin can handle bulk operations', async ({ page }) => {
    await page.goto('/admin/products');

    // Select multiple products
    await page.check('[data-testid="select-all-products"]');
    
    // Verify bulk actions are available
    await expect(page.locator('[data-testid="bulk-actions"]')).toBeVisible();

    // Test bulk status update
    await page.selectOption('[data-testid="bulk-action-select"]', 'deactivate');
    await page.click('button:has-text("Apply")');
    await page.click('button:has-text("Confirm")');

    // Verify bulk action was applied
    await expect(page.locator('text=Products updated successfully')).toBeVisible();
  });
});