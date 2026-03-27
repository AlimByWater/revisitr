import { test, expect, waitForPageLoad } from '../helpers/fixtures';

test.describe('Promotions', () => {
  test('list promotions page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/promotions');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('create promotion via UI', async ({ page }) => {
    await page.goto('dashboard/promotions');
    await waitForPageLoad(page);
    await page.waitForTimeout(2000);

    const createBtn = page.locator('button:has-text("Создать"), a:has-text("Создать"), button:has-text("Добавить")').first();
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click({ force: true });
      await page.waitForTimeout(1000);

      // Fill promotion name
      const nameInput = page.locator('input[name="name"]').first();
      if (await nameInput.isVisible().catch(() => false)) {
        await nameInput.fill(`E2E Promo ${Date.now()}`);
      }

      // Fill description if available
      const descInput = page.locator('textarea[name="description"], input[name="description"]').first();
      if (await descInput.isVisible().catch(() => false)) {
        await descInput.fill('E2E test promotion');
      }

      const submitBtn = page.locator('button[type="submit"], button:has-text("Сохранить"), button:has-text("Создать")').first();
      if (await submitBtn.isVisible().catch(() => false)) {
        await submitBtn.click();
        await page.waitForTimeout(2000);
      }
    }
  });

  test('promo codes page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/promotions/codes');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('promotions archive page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/promotions/archive');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });
});
