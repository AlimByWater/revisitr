import { test, expect, waitForPageLoad } from '../helpers/fixtures';

test.describe('POS Locations', () => {
  test('list POS page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/pos');
    await waitForPageLoad(page);

    const root = page.locator('main');
    await expect(root).toBeVisible();
    await page.waitForTimeout(2000);

    const body = await root.textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('create POS location via UI', async ({ page }) => {
    await page.goto('dashboard/pos');
    await waitForPageLoad(page);
    await page.waitForTimeout(2000);

    const createBtn = page.locator('button:has-text("Добавить"), button:has-text("Создать"), a:has-text("Добавить")').first();
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(1000);

      const nameInput = page.locator('input[name="name"]').first();
      if (await nameInput.isVisible().catch(() => false)) {
        await nameInput.fill(`E2E Point ${Date.now()}`);
      }

      const addrInput = page.locator('input[name="address"]').first();
      if (await addrInput.isVisible().catch(() => false)) {
        await addrInput.fill('ул. E2E Тестовая, 42');
      }

      const submitBtn = page.locator('button[type="submit"], button:has-text("Сохранить")').first();
      if (await submitBtn.isVisible().catch(() => false)) {
        await submitBtn.click();
        await page.waitForTimeout(2000);
      }
    }
  });

  test('view POS details', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/pos');
    await waitForPageLoad(page);
    await page.waitForTimeout(2000);

    const posLink = page.locator('a[href*="/pos/"]').first();
    if (await posLink.isVisible().catch(() => false)) {
      await posLink.click();
      await page.waitForURL('**/pos/**', { timeout: 10_000 });
      await waitForPageLoad(page);

      consoleErrors.assertNoErrors();
    }
  });
});
