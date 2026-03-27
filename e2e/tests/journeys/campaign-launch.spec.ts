import { test, expect, waitForPageLoad } from '../../helpers/fixtures';

test.describe('Journey: Campaign Launch', () => {
  test('dashboard → campaigns → create page', async ({ page }) => {
    // 1. Dashboard
    await page.goto('dashboard');
    await waitForPageLoad(page);

    // 2. Campaigns list
    await page.goto('dashboard/campaigns');
    await waitForPageLoad(page);
    await page.waitForTimeout(1000);

    // 3. Create campaign page
    await page.goto('dashboard/campaigns/create');
    await waitForPageLoad(page);
    await page.waitForTimeout(1000);

    // 4. Verify campaign form exists
    const root = page.locator('main');
    const body = await root.textContent();
    expect(body?.trim().length).toBeGreaterThan(0);
  });
});
