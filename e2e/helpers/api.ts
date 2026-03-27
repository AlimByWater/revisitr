/**
 * HTTP client for direct API calls in E2E tests.
 * Used for seed data creation, state setup, and verification.
 */

const API_BASE = process.env.API_URL || 'http://localhost:8080/api/v1';

// ─── Types ───────────────────────────────────────────────────────────

export interface AuthResponse {
  user: {
    id: number;
    email: string;
    name: string;
    role: string;
    org_id: number;
  };
  tokens: {
    access_token: string;
    refresh_token: string;
    expires_in: number;
  };
}

export interface Bot {
  id: number;
  org_id: number;
  name: string;
  username: string;
  status: string;
  created_at: string;
}

export interface LoyaltyProgram {
  id: number;
  org_id: number;
  name: string;
  type: string;
  created_at: string;
}

export interface POSLocation {
  id: number;
  org_id: number;
  name: string;
  address: string;
}

export interface Subscription {
  id: number;
  org_id: number;
  tariff_id: number;
  status: string;
  tariff_slug: string;
}

export interface Promotion {
  id: number;
  org_id: number;
  name: string;
  type: string;
}

export interface Campaign {
  id: number;
  org_id: number;
  name: string;
  status: string;
}

// ─── HTTP helpers ────────────────────────────────────────────────────

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  token?: string,
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`API ${method} ${path} → ${res.status}: ${text}`);
  }

  // Handle 204 No Content
  if (res.status === 204) return {} as T;

  return res.json() as Promise<T>;
}

// ─── Auth ────────────────────────────────────────────────────────────

export async function register(
  email: string,
  password: string,
  name: string,
  organization: string,
): Promise<AuthResponse> {
  return request<AuthResponse>('POST', '/auth/register', {
    email,
    password,
    name,
    organization,
  });
}

export async function login(
  email: string,
  password: string,
): Promise<AuthResponse> {
  return request<AuthResponse>('POST', '/auth/login', { email, password });
}

// ─── Billing ─────────────────────────────────────────────────────────

export async function createSubscription(
  token: string,
  tariffSlug: string,
): Promise<Subscription> {
  return request<Subscription>(
    'POST',
    '/billing/subscription',
    { tariff_slug: tariffSlug },
    token,
  );
}

export async function upgradeSubscription(
  token: string,
  tariffSlug: string,
): Promise<Subscription> {
  return request<Subscription>(
    'PATCH',
    '/billing/subscription',
    { tariff_slug: tariffSlug },
    token,
  );
}

export async function getSubscription(
  token: string,
): Promise<Subscription | null> {
  try {
    return await request<Subscription>(
      'GET',
      '/billing/subscription',
      undefined,
      token,
    );
  } catch {
    return null;
  }
}

// ─── Bots ────────────────────────────────────────────────────────────

export async function createBot(
  token: string,
  name: string,
  botToken: string,
): Promise<Bot> {
  return request<Bot>('POST', '/bots', { name, token: botToken }, token);
}

export async function listBots(token: string): Promise<Bot[]> {
  return request<Bot[]>('GET', '/bots', undefined, token);
}

export async function deleteBot(token: string, id: number): Promise<void> {
  await request('DELETE', `/bots/${id}`, undefined, token);
}

// ─── Loyalty ─────────────────────────────────────────────────────────

export async function createLoyaltyProgram(
  token: string,
  name: string,
  type: string = 'bonus',
): Promise<LoyaltyProgram> {
  return request<LoyaltyProgram>(
    'POST',
    '/loyalty/programs',
    { name, type },
    token,
  );
}

export async function listLoyaltyPrograms(
  token: string,
): Promise<LoyaltyProgram[]> {
  return request<LoyaltyProgram[]>(
    'GET',
    '/loyalty/programs',
    undefined,
    token,
  );
}

// ─── POS ─────────────────────────────────────────────────────────────

export async function createPOS(
  token: string,
  name: string,
  address: string,
): Promise<POSLocation> {
  return request<POSLocation>('POST', '/pos', { name, address }, token);
}

export async function listPOS(token: string): Promise<POSLocation[]> {
  return request<POSLocation[]>('GET', '/pos', undefined, token);
}

export async function deletePOS(token: string, id: number): Promise<void> {
  await request('DELETE', `/pos/${id}`, undefined, token);
}

// ─── Promotions ──────────────────────────────────────────────────────

export async function createPromotion(
  token: string,
  data: {
    name: string;
    type: string;
    description?: string;
    discount?: number;
  },
): Promise<Promotion> {
  return request<Promotion>('POST', '/promotions', data, token);
}

export async function listPromotions(token: string): Promise<Promotion[]> {
  return request<Promotion[]>('GET', '/promotions', undefined, token);
}

export async function deletePromotion(
  token: string,
  id: number,
): Promise<void> {
  await request('DELETE', `/promotions/${id}`, undefined, token);
}

// ─── Campaigns ───────────────────────────────────────────────────────

export async function createCampaign(
  token: string,
  data: {
    name: string;
    message: string;
    audience_filter?: Record<string, unknown>;
  },
): Promise<Campaign> {
  return request<Campaign>('POST', '/campaigns', data, token);
}

export async function listCampaigns(token: string): Promise<Campaign[]> {
  const res = await request<{ items: Campaign[] }>(
    'GET',
    '/campaigns',
    undefined,
    token,
  );
  return res.items || [];
}

export async function deleteCampaign(
  token: string,
  id: number,
): Promise<void> {
  await request('DELETE', `/campaigns/${id}`, undefined, token);
}

// ─── Onboarding ──────────────────────────────────────────────────────

/**
 * Mark all 6 onboarding steps as completed via PATCH, then call complete.
 * Steps: info, loyalty, bot, pos, integrations, next_steps
 */
export async function completeOnboarding(token: string): Promise<void> {
  const steps = ['info', 'loyalty', 'bot', 'pos', 'integrations', 'next_steps'];
  for (const step of steps) {
    try {
      await request(
        'PATCH',
        '/onboarding',
        { step, completed: true },
        token,
      );
    } catch {
      // Step may already be completed or not supported
    }
  }
  await request('POST', '/onboarding/complete', undefined, token);
}

export async function getOnboarding(
  token: string,
): Promise<{ onboarding_completed: boolean }> {
  return request<{ onboarding_completed: boolean }>(
    'GET',
    '/onboarding',
    undefined,
    token,
  );
}

// ─── RFM ─────────────────────────────────────────────────────────────

export async function setRFMTemplate(
  token: string,
  templateKey: string,
): Promise<void> {
  await request('POST', '/rfm/set-template', { template_key: templateKey }, token);
}

// ─── Menus ───────────────────────────────────────────────────────────

export async function createMenu(
  token: string,
  name: string,
): Promise<{ id: number }> {
  return request<{ id: number }>('POST', '/menus', { name }, token);
}

// ─── Health ──────────────────────────────────────────────────────────

export async function healthCheck(): Promise<{ status: string }> {
  // Simple connectivity check — POST login with empty body returns 400, not connection error
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: '', password: '' }),
  });
  // Any HTTP response (even 400/401) means backend is up
  if (res.status > 0) {
    return { status: 'ok' };
  }
  throw new Error('Backend not reachable');
}
