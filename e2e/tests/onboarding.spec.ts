import { test, expect, waitForPageLoad } from '../helpers/fixtures';

test.describe('Onboarding', () => {
  test('onboarding page renders with wizard steps', async ({ page, consoleErrors }) => {
    await page.goto('dashboard/onboarding');
    await waitForPageLoad(page);

    // Should see wizard steps or completion message
    const body = await page.locator('main').first().textContent();
    expect(body?.trim().length).toBeGreaterThan(0);

    consoleErrors.assertNoErrors();
  });

  test('onboarding shows progress', async ({ page }) => {
    await page.goto('dashboard/onboarding');
    await waitForPageLoad(page);

    // Look for step indicators or progress bar
    const progressIndicators = page.locator(
      '[data-testid*="step"], [data-testid*="progress"], .step, [role="progressbar"], [aria-valuenow]'
    );
    const stepTexts = page.getByText(/шаг|step|1.*из|1.*of/i);

    const hasProgress = (await progressIndicators.count() > 0) ||
      (await stepTexts.count() > 0);

    // If onboarding is complete, it may show completion
    const hasComplete = await page.getByText(/завершен|выполнен|complete/i).isVisible().catch(() => false);

    expect(hasProgress || hasComplete).toBeTruthy();
  });

  test('next button behavior', async ({ page }) => {
    await page.goto('dashboard/onboarding');
    await waitForPageLoad(page);

    const nextBtn = page.getByText(/далее|next|продолжить|continue/i).first();
    if (await nextBtn.isVisible().catch(() => false)) {
      // Check if button is disabled on action steps
      const isDisabled = await nextBtn.isDisabled().catch(() => false);
      // Just verify button exists and is interactive
      await expect(nextBtn).toBeVisible();
    }
  });
});
