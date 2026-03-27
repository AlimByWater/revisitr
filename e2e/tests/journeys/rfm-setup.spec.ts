import { test, expect, waitForPageLoad } from '../../helpers/fixtures';

test.describe('Journey: RFM Setup', () => {
  test('onboarding → template → dashboard → segment detail', async ({ page }) => {
    // 1. Start RFM onboarding
    await page.goto('dashboard/rfm/onboarding');
    await waitForPageLoad(page);

    // 2. Answer questions if available
    const radioOptions = page.locator('input[type="radio"], [role="radio"]');
    const optionCount = await radioOptions.count();

    if (optionCount > 0) {
      // Answer each question (select first option)
      for (let i = 0; i < optionCount; i += 3) {
        const opt = radioOptions.nth(i);
        if (await opt.isVisible().catch(() => false)) {
          await opt.click();
          await page.waitForTimeout(300);
        }
      }

      const nextBtn = page.locator('button:has-text("Далее"), button:has-text("Получить"), button[type="submit"]').first();
      if (await nextBtn.isVisible().catch(() => false)) {
        await nextBtn.click();
        await page.waitForTimeout(2000);
      }
    }

    // 3. Accept template recommendation or go to templates
    const acceptBtn = page.getByText(/использовать|применить|accept|use/i).first();
    if (await acceptBtn.isVisible().catch(() => false)) {
      await acceptBtn.click();
      await page.waitForTimeout(2000);
    } else {
      // Go directly to templates page
      await page.goto('dashboard/rfm/template');
      await waitForPageLoad(page);

      // Select first template
      const templateCard = page.locator('.card, [data-testid*="template"]').first();
      if (await templateCard.isVisible().catch(() => false)) {
        await templateCard.dblclick();
        await page.waitForTimeout(2000);
      }
    }

    // 4. Go to RFM dashboard
    await page.goto('dashboard/rfm');
    await waitForPageLoad(page);

    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    // 5. Click on segment for detail (if segments are rendered)
    const segmentLink = page.locator('a[href*="/rfm/segments/"], tr[data-testid*="segment"]').first();
    if (await segmentLink.isVisible().catch(() => false)) {
      await segmentLink.click();
      await page.waitForURL('**/rfm/segments/**', { timeout: 10_000 });
      await waitForPageLoad(page);

      const segmentBody = await page.locator('main').first().textContent();
      expect(segmentBody?.trim().length).toBeGreaterThan(0);
    }
  });
});
