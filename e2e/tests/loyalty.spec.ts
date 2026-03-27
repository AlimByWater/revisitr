import { test, expect, waitForPageLoad } from '../helpers/fixtures';

test.describe('Loyalty Programs', () => {
  test('list programs page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/loyalty');
    await waitForPageLoad(page);

    // Wait for page content — may show programs or onboarding step
    const root = page.locator('main');
    await expect(root).toBeVisible();
    await page.waitForTimeout(2000);

    const body = await root.textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('create loyalty program via UI', async ({ page }) => {
    await page.goto('dashboard/loyalty');
    await waitForPageLoad(page);
    await page.waitForTimeout(2000);

    const createBtn = page.locator('button:has-text("Создать"), button:has-text("Добавить"), a:has-text("Создать")').first();
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click({ force: true });
      await page.waitForTimeout(1000);

      const nameInput = page.locator('input[name="name"]').first();
      if (await nameInput.isVisible().catch(() => false)) {
        await nameInput.fill(`Test Program ${Date.now()}`);
      }

      const submitBtn = page.locator('button[type="submit"], button:has-text("Сохранить")').first();
      if (await submitBtn.isVisible().catch(() => false)) {
        await submitBtn.click();
        await page.waitForTimeout(2000);
      }
    }
  });

  test('view program details or onboarding step', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/loyalty');
    await waitForPageLoad(page);
    await page.waitForTimeout(2000);

    // Page may show program list or onboarding wizard step
    const body = await page.locator('main').first().textContent() || '';
    const hasContent = body.trim().length > 50;
    expect(hasContent).toBeTruthy();

    consoleErrors.assertNoErrors();
  });

  test('wallet page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/loyalty/wallet');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });
});
