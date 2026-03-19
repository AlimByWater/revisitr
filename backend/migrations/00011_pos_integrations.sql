-- +goose Up
CREATE TABLE integrations (
    id           SERIAL PRIMARY KEY,
    org_id       INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    type         VARCHAR(30) NOT NULL CHECK (type IN ('iiko','rkeeper','1c')),
    config       JSONB NOT NULL DEFAULT '{}',
    status       VARCHAR(20) NOT NULL DEFAULT 'inactive' CHECK (status IN ('active','inactive','error')),
    last_sync_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_integrations_org_id ON integrations(org_id);

CREATE TABLE external_orders (
    id             SERIAL PRIMARY KEY,
    integration_id INT NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    external_id    VARCHAR(255) NOT NULL,
    client_id      INT REFERENCES bot_clients(id) ON DELETE SET NULL,
    items          JSONB NOT NULL DEFAULT '[]',
    total          NUMERIC(12,2) NOT NULL DEFAULT 0,
    ordered_at     TIMESTAMPTZ,
    synced_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(integration_id, external_id)
);
CREATE INDEX idx_external_orders_client_id      ON external_orders(client_id);
CREATE INDEX idx_external_orders_integration_id ON external_orders(integration_id);

-- +goose Down
DROP TABLE IF EXISTS external_orders;
DROP TABLE IF EXISTS integrations;
