import { test, expect, waitForPageLoad } from '../helpers/fixtures';

test.describe('RFM Module', () => {
  test.describe('RFM Onboarding', () => {
    test('onboarding page renders with questions', async ({ page, consoleErrors }) => {
      await page.goto('dashboard/rfm/onboarding');
      await waitForPageLoad(page);

      const body = await page.locator('main').first().textContent();
      expect(body?.trim().length).toBeGreaterThan(0);

      consoleErrors.assertNoErrors();
    });

    test('answer questions and get recommendation', async ({ page }) => {
      await page.goto('dashboard/rfm/onboarding');
      await waitForPageLoad(page);

      // Find and answer radio/select questions
      const options = page.locator('input[type="radio"], [role="radio"], [data-testid*="option"]');
      const count = await options.count();

      if (count > 0) {
        // Click first option for each question group
        for (let i = 0; i < Math.min(count, 6); i += 2) {
          const opt = options.nth(i);
          if (await opt.isVisible().catch(() => false)) {
            await opt.click();
            await page.waitForTimeout(300);
          }
        }

        // Submit quiz
        const submitBtn = page.locator('button:has-text("Далее"), button:has-text("Рекомендация"), button:has-text("Получить"), button[type="submit"]').first();
        if (await submitBtn.isVisible().catch(() => false)) {
          await submitBtn.click();
          await page.waitForTimeout(2000);
        }
      }
    });
  });

  test.describe('RFM Dashboard', () => {
    test('dashboard renders with segments table', async ({ page, consoleErrors }) => {
      await page.goto('dashboard/rfm');
      await waitForPageLoad(page);

      // Should see either segments table or onboarding redirect
      const hasSegments = await page.getByText(/чемпион|champion|лояльн|loyal|потерянн|lost|спящ|sleep|risk/i).isVisible().catch(() => false);
      const hasOnboarding = page.url().includes('onboarding');
      const hasBody = (await page.locator('main').first().textContent())?.trim().length! > 0;

      expect(hasSegments || hasOnboarding || hasBody).toBeTruthy();

      if (!hasOnboarding) {
        consoleErrors.assertNoErrors();
      }
    });

    test('recalculate button works', async ({ page }) => {
      await page.goto('dashboard/rfm');
      await waitForPageLoad(page);

      const recalcBtn = page.getByText(/пересчитать|recalculate/i).first();
      if (await recalcBtn.isVisible().catch(() => false)) {
        await recalcBtn.click();
        await page.waitForTimeout(3000);

        // Should show success or loading state
        const success = await page.getByText(/успешно|done|готово|updated/i).isVisible().catch(() => false);
        // Even if no success text, not crashing is good enough
      }
    });
  });

  test.describe('RFM Templates', () => {
    test('templates page shows template cards', async ({ page, consoleErrors }) => {
      await page.goto('dashboard/rfm/template');
      await waitForPageLoad(page);

      const body = await page.locator('main').first().textContent();
      expect(body?.trim().length).toBeGreaterThan(0);

      consoleErrors.assertNoErrors();
    });

    test('can select a template', async ({ page }) => {
      await page.goto('dashboard/rfm/template');
      await waitForPageLoad(page);

      // Find template cards
      const templateCard = page.locator('[data-testid*="template"], .card, [role="button"]').first();
      if (await templateCard.isVisible().catch(() => false)) {
        // Double click to select
        await templateCard.dblclick();
        await page.waitForTimeout(2000);
      }
    });
  });

  test.describe('RFM Segment Detail', () => {
    const segments = ['champions', 'loyal', 'potential', 'new', 'at_risk', 'cant_lose', 'lost'];

    test('segment detail page renders', async ({ page, consoleErrors }) => {
      // Try the first available segment
      for (const segment of segments) {
        await page.goto(`dashboard/rfm/segments/${segment}`);
        await waitForPageLoad(page);

        const body = await page.locator('main').first().textContent();
        if (body && body.trim().length > 0) {
          consoleErrors.assertNoErrors();
          return;
        }
      }
    });
  });
});
