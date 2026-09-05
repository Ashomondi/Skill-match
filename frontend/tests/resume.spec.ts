import { test, expect } from '@playwright/test';

test.describe('Resume Management Flow', () => {
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

  test('should render upload zone and file requirements', async ({ page }) => {
    await page.goto('/resume');
    await expect(page.getByText('Drag and drop your CV here')).toBeVisible();
    await expect(page.getByText(/PDF.*DOCX.*TXT/i)).toBeVisible();
    await expect(page.locator('input[type="file"]')).toBeAttached();
  });

  test('should display resume management header and description', async ({ page }) => {
    await page.goto('/resume');
    await expect(page.getByText('Your master profile starts here.')).toBeVisible();
    await expect(page.getByText('Upload the resume that best represents your experience.')).toBeVisible();
  });
});
