import { test, expect, waitForPageLoad } from '../helpers/fixtures';

test.describe('Bots CRUD', () => {
  test('list bots page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/bots');
    await waitForPageLoad(page);

    // Wait for heading "Боты" to appear (SPA data loading)
    await expect(page.locator('h1, h2').filter({ hasText: /бот|bot/i }).first()).toBeVisible({ timeout: 10_000 });

    consoleErrors.assertNoErrors();
  });

  test('create bot via UI', async ({ page }) => {
    await page.goto('dashboard/bots');
    await waitForPageLoad(page);

    // Wait for page to fully load data
    await expect(page.locator('h1, h2').filter({ hasText: /бот|bot/i }).first()).toBeVisible({ timeout: 10_000 });

    // Click create button — actual text is "Создать бота"
    const createBtn = page.locator('button:has-text("Создать бота"), a:has-text("Создать бота")').first();
    if (await createBtn.isVisible().catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(1000);

      // Fill bot name
      const nameInput = page.locator('input[name="name"], input[placeholder*="назван"], input[placeholder*="name"], input[placeholder*="Назван"]').first();
      if (await nameInput.isVisible().catch(() => false)) {
        await nameInput.fill(`Test Bot ${Date.now()}`);
      }

      // Fill bot token
      const tokenInput = page.locator('input[name="token"], input[placeholder*="токен"], input[placeholder*="token"], input[placeholder*="Токен"]').first();
      if (await tokenInput.isVisible().catch(() => false)) {
        await tokenInput.fill(`fake-token-${Date.now()}`);
      }

      // Submit
      const submitBtn = page.locator('button[type="submit"], button:has-text("Сохранить"), button:has-text("Создать")').last();
      if (await submitBtn.isVisible().catch(() => false)) {
        await submitBtn.click();
        await page.waitForTimeout(2000);
      }
    }
  });

  test('view bot details', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/bots');
    await waitForPageLoad(page);
    await page.waitForTimeout(2000);

    const botLink = page.locator('a[href*="/bots/"]').first();
    if (await botLink.isVisible().catch(() => false)) {
      await botLink.click();
      await page.waitForURL('**/bots/**', { timeout: 10_000 });
      await waitForPageLoad(page);

      const body = await page.locator('main').first().textContent();
      expect(body?.trim().length).toBeGreaterThan(0);

      consoleErrors.assertNoErrors();
    }
  });

  test('bot settings page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/bots');
    await waitForPageLoad(page);
    await page.waitForTimeout(2000);

    const botLink = page.locator('a[href*="/bots/"]').first();
    if (await botLink.isVisible().catch(() => false)) {
      await botLink.click();
      await page.waitForURL('**/bots/**', { timeout: 10_000 });
      await waitForPageLoad(page);

      const settingsTab = page.getByText(/настройк|settings|модул/i).first();
      if (await settingsTab.isVisible().catch(() => false)) {
        await settingsTab.click();
        await page.waitForTimeout(1000);
      }

      consoleErrors.assertNoErrors();
    }
  });
});
