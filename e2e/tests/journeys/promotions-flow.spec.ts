import { test, expect, waitForPageLoad } from '../../helpers/fixtures';

test.describe('Journey: Promotions Flow', () => {
  test('promotions → codes → archive', async ({ page }) => {
    // 1. Promotions list
    await page.goto('dashboard/promotions');
    await waitForPageLoad(page);
    await page.waitForTimeout(1000);

    const body1 = await page.locator('main').first().textContent();
    expect(body1?.trim().length).toBeGreaterThan(0);

    // 2. Promo codes
    await page.goto('dashboard/promotions/codes');
    await waitForPageLoad(page);
    await page.waitForTimeout(1000);

    const body2 = await page.locator('main').first().textContent();
    expect(body2?.trim().length).toBeGreaterThan(0);

    // 3. Archive
    await page.goto('dashboard/promotions/archive');
    await waitForPageLoad(page);

    const body3 = await page.locator('main').first().textContent();
    expect(body3?.trim().length).toBeGreaterThan(0);
  });
});
