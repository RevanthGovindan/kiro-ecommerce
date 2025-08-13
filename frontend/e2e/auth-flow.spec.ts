import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Start from the homepage
    await page.goto('/');
  });

  test('user can register, login, and logout', async ({ page }) => {
    const testEmail = `test-${Date.now()}@example.com`;
    const testPassword = 'testpassword123';

    // Navigate to register page
    await page.click('text=Sign Up');
    await expect(page).toHaveURL('/auth/register');

    // Fill registration form
    await page.fill('input[name="firstName"]', 'Test');
    await page.fill('input[name="lastName"]', 'User');
    await page.fill('input[name="email"]', testEmail);
    await page.fill('input[name="password"]', testPassword);
    await page.fill('input[name="confirmPassword"]', testPassword);

    // Submit registration
    await page.click('button[type="submit"]');

    // Should redirect to dashboard or home after successful registration
    await expect(page).toHaveURL('/');
    await expect(page.locator('text=Welcome, Test')).toBeVisible();

    // Logout
    await page.click('button:has-text("Logout")');
    await expect(page.locator('text=Sign In')).toBeVisible();

    // Login with the same credentials
    await page.click('text=Sign In');
    await expect(page).toHaveURL('/auth/login');

    await page.fill('input[name="email"]', testEmail);
    await page.fill('input[name="password"]', testPassword);
    await page.click('button[type="submit"]');

    // Should be logged in again
    await expect(page).toHaveURL('/');
    await expect(page.locator('text=Welcome, Test')).toBeVisible();
  });

  test('registration validation works correctly', async ({ page }) => {
    await page.click('text=Sign Up');
    await expect(page).toHaveURL('/auth/register');

    // Try to submit empty form
    await page.click('button[type="submit"]');

    // Should show validation errors
    await expect(page.locator('text=First name is required')).toBeVisible();
    await expect(page.locator('text=Last name is required')).toBeVisible();
    await expect(page.locator('text=Email is required')).toBeVisible();
    await expect(page.locator('text=Password is required')).toBeVisible();

    // Test invalid email
    await page.fill('input[name="email"]', 'invalid-email');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Please enter a valid email')).toBeVisible();

    // Test password too short
    await page.fill('input[name="email"]', 'test@example.com');
    await page.fill('input[name="password"]', '123');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Password must be at least 8 characters')).toBeVisible();

    // Test password mismatch
    await page.fill('input[name="password"]', 'password123');
    await page.fill('input[name="confirmPassword"]', 'different123');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Passwords do not match')).toBeVisible();
  });

  test('login validation works correctly', async ({ page }) => {
    await page.click('text=Sign In');
    await expect(page).toHaveURL('/auth/login');

    // Try to submit empty form
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Email is required')).toBeVisible();
    await expect(page.locator('text=Password is required')).toBeVisible();

    // Test invalid credentials
    await page.fill('input[name="email"]', 'nonexistent@example.com');
    await page.fill('input[name="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=Invalid email or password')).toBeVisible();
  });

  test('forgot password flow works', async ({ page }) => {
    await page.click('text=Sign In');
    await page.click('text=Forgot Password?');
    await expect(page).toHaveURL('/auth/forgot-password');

    // Submit forgot password form
    await page.fill('input[name="email"]', 'test@example.com');
    await page.click('button[type="submit"]');

    // Should show success message
    await expect(page.locator('text=Password reset instructions sent')).toBeVisible();
  });

  test('protected routes redirect to login', async ({ page }) => {
    // Try to access protected route without authentication
    await page.goto('/account');
    await expect(page).toHaveURL('/auth/login');

    await page.goto('/account/orders');
    await expect(page).toHaveURL('/auth/login');

    await page.goto('/admin');
    await expect(page).toHaveURL('/auth/login');
  });
});