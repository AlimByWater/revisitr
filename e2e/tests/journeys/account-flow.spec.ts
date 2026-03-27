import { test, expect, waitForPageLoad } from '../../helpers/fixtures';

test.describe('Journey: Account Management', () => {
  test('account → billing → invoices', async ({ page }) => {
    // 1. Go to account settings
    await page.goto('dashboard/account');
    await waitForPageLoad(page);
    await page.waitForTimeout(2000);

    const accountBody = await page.locator('main').first().textContent();
    expect(accountBody?.trim().length).toBeGreaterThan(0);

    // 2. Navigate to billing
    await page.goto('dashboard/billing');
    await waitForPageLoad(page);
    await page.waitForTimeout(3000);

    const billingBody = await page.locator('main').first().textContent();
    expect(billingBody?.trim().length).toBeGreaterThan(10);

    // 3. Navigate to invoices
    await page.goto('dashboard/billing/invoices');
    await waitForPageLoad(page);

    const invoicesBody = await page.locator('main').first().textContent();
    expect(invoicesBody?.trim().length).toBeGreaterThan(0);
  });
});
