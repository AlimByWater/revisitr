import { test as setup } from '@playwright/test';
import fs from 'fs';
import path from 'path';
import {
  register,
  login,
  createSubscription,
  upgradeSubscription,
  getSubscription,
  createBot,
  createLoyaltyProgram,
  createPOS,
  completeOnboarding,
  healthCheck,
} from '../helpers/api';

// ─── Test credentials ────────────────────────────────────────────────

export const TEST_USER = {
  email: `e2e-${Date.now()}@test.revisitr.local`,
  password: 'E2eTestPass123!',
  name: 'E2E Tester',
  organization: 'E2E Test Org',
};

// ─── Global setup: create user, upgrade to Pro, seed data ────────────

setup('seed test data and authenticate', async () => {
  // 0. Ensure .auth directory exists
  const authDir = path.join(__dirname, '..', '.auth');
  if (!fs.existsSync(authDir)) {
    fs.mkdirSync(authDir, { recursive: true });
  }

  // 1. Wait for backend to be ready
  let retries = 10;
  while (retries > 0) {
    try {
      await healthCheck();
      break;
    } catch {
      retries--;
      if (retries === 0) throw new Error('Backend not available at API_URL');
      await new Promise((r) => setTimeout(r, 2000));
    }
  }

  // 2. Register test user
  console.log(`[setup] Registering user: ${TEST_USER.email}`);
  let auth;
  try {
    auth = await register(
      TEST_USER.email,
      TEST_USER.password,
      TEST_USER.name,
      TEST_USER.organization,
    );
  } catch (e: unknown) {
    // If user already exists (409), login instead
    const msg = e instanceof Error ? e.message : String(e);
    if (msg.includes('409')) {
      console.log('[setup] User exists, logging in...');
      auth = await login(TEST_USER.email, TEST_USER.password);
    } else {
      throw e;
    }
  }
  const token = auth.tokens.access_token;
  console.log(`[setup] Authenticated as user ${auth.user.id}, org ${auth.user.org_id}`);

  // 3. Upgrade to Pro plan
  console.log('[setup] Upgrading to Pro plan...');
  const existing = await getSubscription(token);
  if (!existing) {
    await createSubscription(token, 'pro');
  } else if (existing.tariff_slug !== 'pro') {
    await upgradeSubscription(token, 'pro');
  }
  console.log('[setup] Pro plan active');

  // 4. Seed bot
  console.log('[setup] Creating test bot...');
  try {
    await createBot(token, 'E2E Bot', `e2e-fake-${Date.now()}`);
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    console.log(`[setup] Bot creation: ${msg}`);
  }

  // 5. Seed loyalty program
  console.log('[setup] Creating loyalty program...');
  try {
    await createLoyaltyProgram(token, 'E2E Бонусная');
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    console.log(`[setup] Loyalty: ${msg}`);
  }

  // 6. Seed POS location
  console.log('[setup] Creating POS location...');
  try {
    await createPOS(token, 'E2E Кафе', 'ул. Тестовая, 1');
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    console.log(`[setup] POS: ${msg}`);
  }

  // 7. Complete onboarding
  console.log('[setup] Completing onboarding...');
  try {
    await completeOnboarding(token);
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    console.log(`[setup] Onboarding: ${msg}`);
  }

  // 8. Create storageState programmatically
  console.log('[setup] Creating storage state...');
  const baseURL = process.env.BASE_URL || 'http://localhost:5173/revisitr';
  const origin = new URL(baseURL).origin;

  const storageState = {
    cookies: [],
    origins: [
      {
        origin,
        localStorage: [
          { name: 'token', value: auth.tokens.access_token },
          { name: 'refresh_token', value: auth.tokens.refresh_token },
        ],
      },
    ],
  };

  const authPath = path.join(__dirname, '..', '.auth', 'user.json');
  fs.writeFileSync(authPath, JSON.stringify(storageState, null, 2));
  console.log(`[setup] Storage state saved to ${authPath}`);
});
