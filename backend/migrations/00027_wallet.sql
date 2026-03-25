-- +goose Up

-- Wallet configuration per organization (Apple/Google credentials)
CREATE TABLE wallet_configs (
    id          SERIAL PRIMARY KEY,
    org_id      INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    platform    VARCHAR(10) NOT NULL CHECK (platform IN ('apple', 'google')),
    is_enabled  BOOLEAN NOT NULL DEFAULT false,
    -- Apple: pass_type_id, team_id, certificate (base64)
    -- Google: issuer_id, service_account_key
    credentials JSONB NOT NULL DEFAULT '{}',
    -- Pass appearance
    design      JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, platform)
);

-- Individual wallet passes for clients
CREATE TABLE wallet_passes (
    id              SERIAL PRIMARY KEY,
    org_id          INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    client_id       INT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    platform        VARCHAR(10) NOT NULL CHECK (platform IN ('apple', 'google')),
    serial_number   VARCHAR(64) NOT NULL UNIQUE,
    auth_token      VARCHAR(128) NOT NULL,
    push_token      TEXT,
    -- Cached pass state
    last_balance    INT NOT NULL DEFAULT 0,
    last_level      VARCHAR(100) NOT NULL DEFAULT '',
    last_updated_at TIMESTAMPTZ,
    status          VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'revoked')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_wallet_passes_org_id ON wallet_passes(org_id);
CREATE INDEX idx_wallet_passes_client_id ON wallet_passes(client_id);
CREATE INDEX idx_wallet_passes_serial ON wallet_passes(serial_number);
CREATE UNIQUE INDEX idx_wallet_passes_client_platform ON wallet_passes(client_id, platform);

-- +goose Down
DROP TABLE IF EXISTS wallet_passes;
DROP TABLE IF EXISTS wallet_configs;
