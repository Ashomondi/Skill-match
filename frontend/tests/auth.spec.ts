import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test('should render login page with form elements', async ({ page }) => {
    await page.goto('/login');
    await expect(page.locator('input#email')).toBeVisible();
    await expect(page.locator('input#password')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('should show validation error on invalid login attempt', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input#email', 'notanemail');
    await page.fill('input#password', 'short');
    await page.click('button[type="submit"]');

    await expect(page.getByText('Enter a valid email address.')).toBeVisible();
  });

  test('should allow demo login and redirect to dashboard', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input#email', 'demo@skill-match.test');
    await page.fill('input#password', 'Demo1234!');
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL(/.*dashboard/);
    await expect(page.locator('header, nav, aside').first()).toBeVisible();
  });

  test('should protect dashboard and redirect unauthenticated users to login', async ({ page }) => {
    await page.goto('/login');
    await page.evaluate(() => localStorage.clear());
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/.*login/);
  });
});
