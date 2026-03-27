import { test, expect, waitForPageLoad } from '../helpers/fixtures';

test.describe('Campaigns', () => {
  test('list campaigns page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/campaigns');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('create campaign page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/campaigns/create');
    await waitForPageLoad(page);

    // Should see campaign creation form
    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('create campaign flow', async ({ page }) => {
    await page.goto('dashboard/campaigns/create');
    await waitForPageLoad(page);

    // Fill campaign name
    const nameInput = page.locator('input[name="name"]').first();
    if (await nameInput.isVisible().catch(() => false)) {
      await nameInput.fill(`E2E Campaign ${Date.now()}`);
    }

    // Fill message text
    const messageInput = page.locator('textarea[name="message"], textarea').first();
    if (await messageInput.isVisible().catch(() => false)) {
      await messageInput.fill('Тестовая рассылка E2E');
    }

    // Select bot if dropdown exists
    const botSelect = page.locator('select[name="bot_id"], [data-testid="bot-select"]').first();
    if (await botSelect.isVisible().catch(() => false)) {
      // Select first option
      const options = botSelect.locator('option');
      const count = await options.count();
      if (count > 1) {
        await botSelect.selectOption({ index: 1 });
      }
    }
  });

  test('campaign templates page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/campaigns/templates');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('auto-scenarios page renders', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/campaigns/scenarios');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });
});
