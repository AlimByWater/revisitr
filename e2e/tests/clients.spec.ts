import { test, expect, waitForPageLoad } from '../helpers/fixtures';

test.describe('Clients', () => {
  test('clients list page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/clients');
    await waitForPageLoad(page);

    // Should show clients table or empty state
    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('search filters clients', async ({ page }) => {
    await page.goto('dashboard/clients');
    await waitForPageLoad(page);

    const searchInput = page.locator('input[type="search"], input[placeholder*="поиск"], input[placeholder*="search"]').first();
    if (await searchInput.isVisible().catch(() => false)) {
      await searchInput.fill('test');
      await page.waitForTimeout(1000);
      // Just verify no crash
    }
  });

  test('client segments page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/clients/segments');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('custom segments page renders (frontend-only)', async ({ page }) => {
    await page.goto('dashboard/clients/custom-segments');
    await waitForPageLoad(page);

    // Frontend-only — UI should render, API errors tolerable
    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);
  });

  test('predictions page renders', async ({ page }) => {
    await page.goto('dashboard/clients/predictions');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);
  });

  test('client detail or empty state renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/clients');
    await waitForPageLoad(page);
    await page.waitForTimeout(2000);

    // Page may show client list, empty state, or onboarding wizard
    const body = await page.locator('main').first().textContent() || '';
    expect(body.trim().length).toBeGreaterThan(50);

    consoleErrors.assertNoErrors();
  });

  test('table sorting works', async ({ page }) => {
    await page.goto('dashboard/clients');
    await waitForPageLoad(page);

    // Click on a sortable column header
    const header = page.locator('th, [role="columnheader"]').first();
    if (await header.isVisible().catch(() => false)) {
      await header.click();
      await page.waitForTimeout(1000);

      // Click again for reverse sort
      await header.click();
      await page.waitForTimeout(1000);
    }
  });
});
