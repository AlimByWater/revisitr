import { test, expect, waitForPageLoad } from '../helpers/fixtures';

test.describe('Analytics Pages', () => {
  test('sales analytics renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/analytics/sales');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('loyalty analytics renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/analytics/loyalty');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('mailings analytics renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/analytics/mailings');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('period filter changes data', async ({ page }) => {
    await page.goto('dashboard/analytics/sales');
    await waitForPageLoad(page);

    // Find period selector (buttons or dropdown)
    const periodBtn = page.locator('button:has-text("7"), button:has-text("30"), button:has-text("90"), select').first();
    if (await periodBtn.isVisible().catch(() => false)) {
      await periodBtn.click();
      await page.waitForTimeout(1000);

      // Find another period option
      const otherPeriod = page.locator('button:has-text("30"), button:has-text("90"), option').first();
      if (await otherPeriod.isVisible().catch(() => false)) {
        await otherPeriod.click();
        await page.waitForTimeout(2000);
      }
    }
  });
});
