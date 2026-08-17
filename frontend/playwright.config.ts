import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E configuration for SkillMatch.
 *
 * The chat and resume lifecycle tests exercise the real stack
 * (React frontend -> Go API -> CockroachDB -> Bedrock), so a running
 * backend is expected. Point E2E_API_URL at it (defaults to the local
 * API on :8080). Set VITE_DEMO_AUTH_ENABLED=true in the web server env
 * to run against demo auth when no backend is available.
 */
export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['github'], ['list']] : 'list',
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  webServer: process.env.E2E_SKIP_WEB_SERVER === 'true' ? undefined : {
    command: 'npm run dev -- --host 127.0.0.1 --port 5173',
    url: 'http://127.0.0.1:5173',
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    env: {
      VITE_API_BASE_URL: process.env.E2E_API_URL || 'http://localhost:8080/api',
      VITE_DEMO_AUTH_ENABLED: process.env.E2E_DEMO_AUTH_ENABLED || 'true',
    },
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
});
