-- +goose Up

-- Tariff plans
CREATE TABLE IF NOT EXISTS tariffs (
    id          SERIAL PRIMARY KEY,
    name        TEXT    NOT NULL,
    slug        TEXT    NOT NULL UNIQUE,
    price       INTEGER NOT NULL DEFAULT 0,       -- price in kopeks (RUB * 100)
    currency    TEXT    NOT NULL DEFAULT 'RUB',
    interval    TEXT    NOT NULL DEFAULT 'month',  -- 'month' | 'year'
    features    JSONB   NOT NULL DEFAULT '{}',
    limits      JSONB   NOT NULL DEFAULT '{}',
    active      BOOLEAN NOT NULL DEFAULT true,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Subscriptions (one active per org)
CREATE TABLE IF NOT EXISTS subscriptions (
    id                   SERIAL PRIMARY KEY,
    org_id               INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tariff_id            INTEGER NOT NULL REFERENCES tariffs(id),
    status               TEXT    NOT NULL DEFAULT 'active',  -- 'trialing' | 'active' | 'past_due' | 'canceled' | 'expired'
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT now(),
    current_period_end   TIMESTAMPTZ NOT NULL,
    canceled_at          TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_org_id ON subscriptions(org_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);

-- Invoices
CREATE TABLE IF NOT EXISTS invoices (
    id              SERIAL PRIMARY KEY,
    org_id          INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    subscription_id INTEGER REFERENCES subscriptions(id) ON DELETE SET NULL,
    amount          INTEGER NOT NULL,           -- amount in kopeks
    currency        TEXT    NOT NULL DEFAULT 'RUB',
    status          TEXT    NOT NULL DEFAULT 'pending',  -- 'pending' | 'paid' | 'failed' | 'refunded'
    due_date        TIMESTAMPTZ NOT NULL,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_invoices_org_id ON invoices(org_id);

-- Payments
CREATE TABLE IF NOT EXISTS payments (
    id                  SERIAL PRIMARY KEY,
    invoice_id          INTEGER NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    org_id              INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    amount              INTEGER NOT NULL,       -- amount in kopeks
    currency            TEXT    NOT NULL DEFAULT 'RUB',
    provider            TEXT    NOT NULL,       -- 'yukassa' | 'cloudpayments' | 'manual'
    provider_payment_id TEXT,
    status              TEXT    NOT NULL DEFAULT 'pending',  -- 'pending' | 'succeeded' | 'failed' | 'refunded'
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX IF NOT EXISTS idx_payments_org_id ON payments(org_id);

-- Seed default tariffs
INSERT INTO tariffs (name, slug, price, currency, interval, features, limits, sort_order) VALUES
(
    'Trial',
    'trial',
    0,
    'RUB',
    'month',
    '{"loyalty": true, "campaigns": true, "promotions": true, "integrations": true, "analytics": true, "rfm": false, "advanced_campaigns": false}'::jsonb,
    '{"max_clients": 100, "max_bots": 1, "max_campaigns_per_month": 10, "max_pos": 1}'::jsonb,
    0
),
(
    'Basic',
    'basic',
    290000,
    'RUB',
    'month',
    '{"loyalty": true, "campaigns": true, "promotions": true, "integrations": true, "analytics": true, "rfm": false, "advanced_campaigns": false}'::jsonb,
    '{"max_clients": 1000, "max_bots": 2, "max_campaigns_per_month": 50, "max_pos": 3}'::jsonb,
    1
),
(
    'Pro',
    'pro',
    790000,
    'RUB',
    'month',
    '{"loyalty": true, "campaigns": true, "promotions": true, "integrations": true, "analytics": true, "rfm": true, "advanced_campaigns": true}'::jsonb,
    '{"max_clients": 10000, "max_bots": 10, "max_campaigns_per_month": -1, "max_pos": 20}'::jsonb,
    2
),
(
    'Enterprise',
    'enterprise',
    0,
    'RUB',
    'month',
    '{"loyalty": true, "campaigns": true, "promotions": true, "integrations": true, "analytics": true, "rfm": true, "advanced_campaigns": true}'::jsonb,
    '{"max_clients": -1, "max_bots": -1, "max_campaigns_per_month": -1, "max_pos": -1}'::jsonb,
    3
);

-- Give all existing organizations a trial subscription (30 days)
INSERT INTO subscriptions (org_id, tariff_id, status, current_period_start, current_period_end)
SELECT o.id, t.id, 'trialing', now(), now() + INTERVAL '30 days'
FROM organizations o
CROSS JOIN tariffs t
WHERE t.slug = 'trial'
AND NOT EXISTS (
    SELECT 1 FROM subscriptions s WHERE s.org_id = o.id
);

-- +goose Down
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS tariffs;
