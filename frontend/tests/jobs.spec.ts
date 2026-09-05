import { test, expect } from '@playwright/test';

test.describe('Jobs Discovery & Filtering', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.evaluate(() => {
      localStorage.setItem('token', 'local-demo-token');
      localStorage.setItem(
        'user',
        JSON.stringify({ id: 'local-demo-user', email: 'demo@skill-match.test', fullName: 'Jane Doe' })
      );
    });
  });

  test('should render search controls and input fields', async ({ page }) => {
    await page.goto('/discover');
    await expect(page.locator('input[aria-label="Search jobs"]')).toBeVisible();
    await expect(page.locator('select[aria-label="Filter by location"]')).toBeVisible();
  });

  test('should display unsupported filters as disabled with Coming Soon indicators', async ({ page }) => {
    await page.goto('/discover');

    const senioritySelect = page.locator('select[aria-label*="seniority" i]');
    await expect(senioritySelect).toBeDisabled();
    await expect(senioritySelect).toContainText('Coming soon');

    const workTypeSelect = page.locator('select[aria-label*="work type" i]');
    await expect(workTypeSelect).toBeDisabled();
    await expect(workTypeSelect).toContainText('Coming soon');
  });

  test('should allow filtering by location', async ({ page }) => {
    await page.goto('/discover');
    const locationSelect = page.locator('select[aria-label="Filter by location"]');
    await locationSelect.selectOption('remote');
    await expect(locationSelect).toHaveValue('remote');
  });
});
