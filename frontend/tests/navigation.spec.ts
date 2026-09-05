import { test, expect } from '@playwright/test';

test.describe('Application Navigation', () => {
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

  test('should navigate to dashboard and display core shell', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/.*dashboard/);
    await expect(page.locator('header, nav, aside').first()).toBeVisible();
  });

  test('should navigate to jobs discovery page', async ({ page }) => {
    await page.goto('/discover');
    await expect(page).toHaveURL(/.*discover/);
    await expect(page.getByText('Find your next role')).toBeVisible();
  });

  test('should navigate to resume management page', async ({ page }) => {
    await page.goto('/resume');
    await expect(page).toHaveURL(/.*resume/);
    await expect(page.getByText('Your master profile starts here.')).toBeVisible();
  });

  test('should navigate to AI-tailored CV page', async ({ page }) => {
    await page.goto('/cv-tailor');
    await expect(page).toHaveURL(/.*cv-tailor/);
    await expect(page.getByText('AI-Tailored CV')).toBeVisible();
  });

  test('should navigate to master profile page', async ({ page }) => {
    await page.goto('/profile');
    await expect(page).toHaveURL(/.*profile/);
    await expect(page.getByText('Your Master Profile')).toBeVisible();
  });

  test('should navigate to applications tracker', async ({ page }) => {
    await page.goto('/applications');
    await expect(page).toHaveURL(/.*applications/);
    await expect(page.locator('header, nav, aside, h1').first()).toBeVisible();
  });

  test('should navigate to saved jobs bookmarks', async ({ page }) => {
    await page.goto('/saved-jobs');
    await expect(page).toHaveURL(/.*saved-jobs/);
    await expect(page.locator('header, nav, aside, h1').first()).toBeVisible();
  });

  test('should navigate to AI chat assistant', async ({ page }) => {
    await page.goto('/chat');
    await expect(page).toHaveURL(/.*chat/);
    await expect(page.locator('input, textarea').first()).toBeVisible();
  });
});
