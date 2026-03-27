import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  // Auth tests don't use stored auth state
  test.use({ storageState: undefined });

  test('login with valid credentials redirects to dashboard', async ({ page }) => {
    await page.goto('auth/login');
    await page.waitForTimeout(1000);

    await page.locator('input[type="email"], input[name="email"]').first().fill(
      process.env.E2E_USER_EMAIL || 'admin@revisitr.com'
    );
    await page.locator('input[type="password"], input[name="password"]').first().fill(
      process.env.E2E_USER_PASSWORD || 'admin123'
    );
    await page.locator('button[type="submit"]').click();

    await page.waitForURL('**/dashboard**', { timeout: 15_000 });
    expect(page.url()).toContain('/dashboard');
  });

  test('login with invalid password shows error', async ({ page }) => {
    await page.goto('auth/login');
    await page.waitForTimeout(1000);

    await page.locator('input[type="email"], input[name="email"]').first().fill('admin@revisitr.com');
    await page.locator('input[type="password"], input[name="password"]').first().fill('wrongpassword');
    await page.locator('button[type="submit"]').click();

    // Should stay on login page or show error
    await page.waitForTimeout(3000);

    const stillOnLogin = page.url().includes('/auth/login');
    const errorVisible = await page.getByText(/неверн|ошибк|invalid|incorrect|неправильн/i).isVisible().catch(() => false);
    expect(stillOnLogin || errorVisible).toBeTruthy();
  });

  test('login with empty fields shows validation', async ({ page }) => {
    await page.goto('auth/login');
    await page.waitForTimeout(1000);

    await page.locator('button[type="submit"]').click();
    await page.waitForTimeout(1000);
    expect(page.url()).toContain('/auth/login');
  });

  test('register new user', async ({ page }) => {
    await page.goto('auth/register');
    await page.waitForTimeout(1000);

    const uniqueEmail = `e2e-reg-${Date.now()}@test.local`;

    // Fill all visible inputs in order
    const inputs = page.locator('form input:visible');
    const count = await inputs.count();

    // The register form has: fullName/name, email, password, organization, phone
    // We fill by type since field names may vary
    const emailInput = page.locator('input[type="email"], input[name="email"]').first();
    await emailInput.fill(uniqueEmail);

    const passwordInput = page.locator('input[type="password"], input[name="password"]').first();
    await passwordInput.fill('TestPassword123!');

    // Find and fill name field (not email, not password)
    const nameInputs = page.locator('input[type="text"]:visible');
    const nameCount = await nameInputs.count();
    if (nameCount >= 1) {
      await nameInputs.nth(0).fill('E2E Register Test');
    }
    if (nameCount >= 2) {
      await nameInputs.nth(1).fill('E2E Register Org');
    }

    // Phone field
    const phoneInput = page.locator('input[type="tel"], input[name="phone"]').first();
    if (await phoneInput.isVisible().catch(() => false)) {
      await phoneInput.fill('+79991234567');
    }

    await page.locator('button[type="submit"]').click();

    // Should redirect to dashboard or onboarding
    await page.waitForURL(/\/(dashboard|onboarding)/, { timeout: 15_000 });
  });

  test('register with existing email shows error', async ({ page }) => {
    await page.goto('auth/register');
    await page.waitForTimeout(1000);

    const emailInput = page.locator('input[type="email"], input[name="email"]').first();
    await emailInput.fill('admin@revisitr.com');

    const passwordInput = page.locator('input[type="password"], input[name="password"]').first();
    await passwordInput.fill('TestPassword123!');

    const nameInputs = page.locator('input[type="text"]:visible');
    const nameCount = await nameInputs.count();
    if (nameCount >= 1) await nameInputs.nth(0).fill('Dup Test');
    if (nameCount >= 2) await nameInputs.nth(1).fill('Dup Org');

    await page.locator('button[type="submit"]').click();
    await page.waitForTimeout(3000);

    const hasError = await page.getByText(/уже зарегист|already|существует|занят/i).isVisible().catch(() => false);
    const onRegisterPage = page.url().includes('/auth/register');
    expect(hasError || onRegisterPage).toBeTruthy();
  });

  test('protected route redirects to login when unauthenticated', async ({ page }) => {
    // Clear any stored tokens
    await page.goto('auth/login');
    await page.evaluate(() => {
      localStorage.clear();
    });
    await page.goto('dashboard');
    await page.waitForTimeout(5000);
    const url = page.url();
    const redirected = url.includes('/auth/login') || url.includes('/auth/');
    expect(redirected).toBeTruthy();
  });

  test('logout redirects to login page', async ({ page }) => {
    // First login
    await page.goto('auth/login');
    await page.waitForTimeout(1000);

    await page.locator('input[type="email"], input[name="email"]').first().fill(
      process.env.E2E_USER_EMAIL || 'admin@revisitr.com'
    );
    await page.locator('input[type="password"], input[name="password"]').first().fill(
      process.env.E2E_USER_PASSWORD || 'admin123'
    );
    await page.locator('button[type="submit"]').click();
    await page.waitForURL('**/dashboard**', { timeout: 15_000 });

    // Find and click logout via user menu
    const headerBtns = page.locator('header button, [data-testid="user-menu"]');
    const lastBtn = headerBtns.last();
    if (await lastBtn.isVisible().catch(() => false)) {
      await lastBtn.click();
      await page.waitForTimeout(500);
    }

    const logoutBtn = page.getByText(/выйти|logout/i);
    if (await logoutBtn.isVisible().catch(() => false)) {
      await logoutBtn.click();
      await page.waitForURL('**/auth/login**', { timeout: 10_000 });
      expect(page.url()).toContain('/auth/login');
    }
  });
});
