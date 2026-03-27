import { test, expect, waitForPageLoad, assertNoLoadingErrors } from '../helpers/fixtures';

/**
 * Navigation smoke test — verify every page renders without errors.
 * Uses authenticated storageState from globalSetup (Pro plan).
 */

const pages = [
  // Dashboard
  { path: 'dashboard', name: 'Dashboard' },

  // Bots
  { path: 'dashboard/bots', name: 'Bots list' },

  // Clients
  { path: 'dashboard/clients', name: 'Clients list' },
  { path: 'dashboard/clients/segments', name: 'Client segments' },
  { path: 'dashboard/clients/custom-segments', name: 'Custom segments', allowApiErrors: true },
  { path: 'dashboard/clients/predictions', name: 'Predictions', allowApiErrors: true },

  // Loyalty
  { path: 'dashboard/loyalty', name: 'Loyalty programs' },
  { path: 'dashboard/loyalty/wallet', name: 'Wallet passes' },

  // Campaigns
  { path: 'dashboard/campaigns', name: 'Campaigns list' },
  { path: 'dashboard/campaigns/create', name: 'Create campaign' },
  { path: 'dashboard/campaigns/templates', name: 'Campaign templates' },
  { path: 'dashboard/campaigns/scenarios', name: 'Auto-scenarios' },

  // Promotions
  { path: 'dashboard/promotions', name: 'Promotions list' },
  { path: 'dashboard/promotions/codes', name: 'Promo codes' },
  { path: 'dashboard/promotions/archive', name: 'Promotions archive' },

  // RFM
  { path: 'dashboard/rfm', name: 'RFM dashboard' },
  { path: 'dashboard/rfm/onboarding', name: 'RFM onboarding' },
  { path: 'dashboard/rfm/template', name: 'RFM templates' },

  // Analytics
  { path: 'dashboard/analytics/sales', name: 'Sales analytics' },
  { path: 'dashboard/analytics/loyalty', name: 'Loyalty analytics' },
  { path: 'dashboard/analytics/mailings', name: 'Mailings analytics' },

  // POS
  { path: 'dashboard/pos', name: 'POS locations' },

  // Menus
  { path: 'dashboard/menus', name: 'Menus' },

  // Integrations
  { path: 'dashboard/integrations', name: 'Integrations' },

  // Marketplace
  { path: 'dashboard/marketplace', name: 'Marketplace' },

  // Billing
  { path: 'dashboard/billing', name: 'Billing' },
  { path: 'dashboard/billing/invoices', name: 'Invoices' },

  // Account
  { path: 'dashboard/account', name: 'Account settings', allowApiErrors: true },

  // Onboarding
  { path: 'dashboard/onboarding', name: 'Onboarding' },
];

test.describe('Navigation Smoke — All Pages Render', () => {
  for (const p of pages) {
    test(`${p.name} (${p.path})`, async ({ page, consoleErrors }) => {
      await page.goto(p.path);
      await waitForPageLoad(page);

      // Page should not be blank
      const body = await page.locator('body').textContent();
      expect(body?.trim().length).toBeGreaterThan(0);

      // Check for loading errors (unless page is frontend-only)
      if (!p.allowApiErrors) {
        await assertNoLoadingErrors(page);
      }

      // No unhandled JS errors
      consoleErrors.assertNoErrors();
    });
  }
});
