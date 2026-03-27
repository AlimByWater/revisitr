import { test, expect, waitForPageLoad } from '../helpers/fixtures';

test.describe('Smoke Test', () => {
  test('dashboard loads with sidebar and content', async ({ page, consoleErrors }) => {
    await page.goto('dashboard');
    await waitForPageLoad(page);

    // Sidebar (aside) should be visible
    await expect(page.locator('aside').first()).toBeVisible({ timeout: 10_000 });

    // Main content area should be visible
    await expect(page.locator('main').first()).toBeVisible({ timeout: 10_000 });

    consoleErrors.assertNoErrors();
  });

  test('API healthz returns ok', async ({ request }) => {
    // healthz is at root path, not under /api/v1
    const baseOrigin = process.env.API_URL?.replace(/\/api\/v1$/, '') || 'http://localhost:8080';
    const res = await request.get(`${baseOrigin}/healthz`);
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.status).toBe('ok');
  });
});
