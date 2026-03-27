import { test, expect, waitForPageLoad } from '../../helpers/fixtures';

test.describe('Journey: New User Registration to Dashboard', () => {
  test.use({ storageState: undefined });

  test('register → dashboard', async ({ page }) => {
    // 1. Register
    await page.goto('auth/register');
    await page.waitForTimeout(1000);
    const email = `e2e-journey-${Date.now()}@test.local`;

    // Fill email and password by type
    await page.locator('input[type="email"], input[name="email"]').first().fill(email);
    await page.locator('input[type="password"], input[name="password"]').first().fill('JourneyPass123!');

    // Fill text fields (name, organization)
    const textInputs = page.locator('input[type="text"]:visible');
    const textCount = await textInputs.count();
    if (textCount >= 1) await textInputs.nth(0).fill('Journey Tester');
    if (textCount >= 2) await textInputs.nth(1).fill('Journey Org');

    // Phone
    const phoneInput = page.locator('input[type="tel"]').first();
    if (await phoneInput.isVisible().catch(() => false)) {
      await phoneInput.fill('+79990001122');
    }

    await page.locator('button[type="submit"]').click();

    // 2. Should land on dashboard or onboarding
    await page.waitForURL(/\/(dashboard|onboarding)/, { timeout: 15_000 });
    await waitForPageLoad(page);

    // 3. Verify page loaded
    const root = page.locator('main');
    await expect(root).toBeVisible();
    const body = await root.textContent();
    expect(body?.trim().length).toBeGreaterThan(0);
  });
});
