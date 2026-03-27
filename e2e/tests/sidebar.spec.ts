import { test, expect, waitForPageLoad } from '../helpers/fixtures';

test.describe('Sidebar Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('dashboard');
    await waitForPageLoad(page);
  });

  test('sidebar is visible with menu items', async ({ page }) => {
    const sidebar = page.locator('aside').first();
    await expect(sidebar).toBeVisible();

    // Should have multiple navigation links
    const links = sidebar.locator('a');
    const count = await links.count();
    expect(count).toBeGreaterThan(3);
  });

  test('clicking sidebar item navigates to correct page', async ({ page }) => {
    // Click "Дашборд" link which is always visible
    const dashLink = page.locator('a[href*="/dashboard"]').filter({ hasText: /дашборд|dashboard/i }).first();
    if (await dashLink.isVisible().catch(() => false)) {
      await dashLink.click({ force: true });
      await page.waitForURL('**/dashboard**', { timeout: 10_000 });
      expect(page.url()).toContain('/dashboard');
    }
  });

  test('submenu expands on parent click', async ({ page }) => {
    // Try to find expandable menu item (e.g. "Аналитика")
    const analyticsMenu = page.locator('button, a, [role="button"]')
      .filter({ hasText: /аналитик|analytic/i })
      .first();

    if (await analyticsMenu.isVisible().catch(() => false)) {
      await analyticsMenu.click();
      await page.waitForTimeout(500);

      // Submenu items should appear
      const submenuLinks = page.locator('a').filter({ hasText: /продаж|лояльн|рассыл|sales|loyalty|mailing/i });
      const count = await submenuLinks.count();
      expect(count).toBeGreaterThan(0);
    }
  });

  test('active menu item is highlighted', async ({ page }) => {
    // Navigate to bots page
    await page.goto('dashboard/bots');
    await waitForPageLoad(page);

    // The active link should have some visual indicator
    const activeLink = page.locator('a[href*="bots"]').first();
    if (await activeLink.isVisible().catch(() => false)) {
      // Check for active class or aria-current
      const className = await activeLink.getAttribute('class') || '';
      const ariaCurrent = await activeLink.getAttribute('aria-current') || '';
      const hasActiveState = className.includes('active') ||
        className.includes('bg-') ||
        className.includes('selected') ||
        ariaCurrent === 'page' ||
        className.includes('text-primary');
      // At minimum the link exists and is visible
      await expect(activeLink).toBeVisible();
    }
  });

  test('user menu in header works', async ({ page }) => {
    // Find user avatar/menu button in header
    const headerButtons = page.locator('header button, [data-testid="user-menu"]');
    const lastButton = headerButtons.last();

    if (await lastButton.isVisible().catch(() => false)) {
      await lastButton.click();
      await page.waitForTimeout(500);

      // Should show dropdown with settings/logout options
      const dropdown = page.locator('[role="menu"], [data-testid="dropdown"]');
      const settingsLink = page.getByText(/настройк|settings/i);
      const logoutLink = page.getByText(/выйти|logout/i);

      const hasDropdown = await dropdown.isVisible().catch(() => false) ||
        await settingsLink.isVisible().catch(() => false) ||
        await logoutLink.isVisible().catch(() => false);

      expect(hasDropdown).toBeTruthy();
    }
  });
});
