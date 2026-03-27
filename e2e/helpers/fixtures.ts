import { test as base, expect, Page } from '@playwright/test';
import { errorPatterns } from './selectors';

/**
 * Extended test fixtures for Revisitr E2E tests.
 */

// ─── Console error collector ─────────────────────────────────────────

export interface ConsoleErrors {
  errors: string[];
  /** Assert no unexpected console errors on the page */
  assertNoErrors: () => void;
}

// ─── Custom test fixture ─────────────────────────────────────────────

export const test = base.extend<{ consoleErrors: ConsoleErrors }>({
  consoleErrors: async ({ page }, use) => {
    const errors: string[] = [];

    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        const text = msg.text();
        if (text.includes('favicon.ico')) return;
        if (text.includes('net::ERR_')) return;
        errors.push(text);
      }
    });

    page.on('pageerror', (err) => {
      errors.push(`PageError: ${err.message}`);
    });

    await use({
      errors,
      assertNoErrors() {
        const critical = errors.filter(
          (e) =>
            !e.includes('ResizeObserver') &&
            !e.includes('Non-Error promise rejection') &&
            !e.includes('Failed to load resource'),
        );
        expect(critical, `Unexpected console errors: ${critical.join('\n')}`).toHaveLength(0);
      },
    });
  },
});

export { expect };

// ─── Page helpers ────────────────────────────────────────────────────

/**
 * Wait for page to fully load:
 * 1. DOM content loaded
 * 2. <main> element visible (SPA has rendered the layout)
 * 3. Loading spinners hidden
 */
export async function waitForPageLoad(page: Page, timeout = 15_000) {
  await page.waitForLoadState('domcontentloaded', { timeout });

  // Wait for <main> to appear — confirms SPA layout rendered
  try {
    await page.locator('main').first().waitFor({ state: 'visible', timeout });
  } catch {
    // Auth pages (login/register) may not have <main> — fallback to #root
    await page.locator('#root').first().waitFor({ state: 'visible', timeout });
  }

  // Wait for loading spinners to disappear
  const spinner = page.locator('.animate-spin, [data-testid="loading"]');
  try {
    await spinner.first().waitFor({ state: 'hidden', timeout: 5_000 });
  } catch {
    // No spinner found or already hidden
  }
}

/**
 * Assert the page has no error messages visible.
 */
export async function assertNoLoadingErrors(page: Page) {
  for (const pattern of errorPatterns) {
    const errorEl = page.getByText(pattern, { exact: false });
    const count = await errorEl.count();
    expect(count, `Found error text "${pattern}" on page ${page.url()}`).toBe(0);
  }
}

/**
 * Navigate to a page and verify it loads without errors.
 */
export async function navigateAndVerify(
  page: Page,
  path: string,
  options?: {
    waitForSelector?: string;
    allowApiErrors?: boolean;
    timeout?: number;
  },
) {
  await page.goto(path, { timeout: options?.timeout || 15_000 });
  await waitForPageLoad(page, options?.timeout);

  if (options?.waitForSelector) {
    await page.waitForSelector(options.waitForSelector, {
      timeout: options?.timeout || 10_000,
    });
  }

  if (!options?.allowApiErrors) {
    await assertNoLoadingErrors(page);
  }
}
