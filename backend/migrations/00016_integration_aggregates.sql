-- +goose Up

CREATE TABLE integration_aggregates (
    id             SERIAL PRIMARY KEY,
    integration_id INT NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    date           DATE NOT NULL,
    revenue        NUMERIC(14,2) NOT NULL DEFAULT 0,
    avg_check      NUMERIC(10,2) NOT NULL DEFAULT 0,
    tx_count       INT NOT NULL DEFAULT 0,
    guest_count    INT NOT NULL DEFAULT 0,
    matched_count  INT NOT NULL DEFAULT 0,
    synced_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(integration_id, date)
);
CREATE INDEX idx_integration_aggregates_date ON integration_aggregates(integration_id, date);

CREATE TABLE integration_client_map (
    id              SERIAL PRIMARY KEY,
    integration_id  INT NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    external_phone  VARCHAR(20) NOT NULL,
    client_id       INT REFERENCES bot_clients(id) ON DELETE SET NULL,
    matched_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(integration_id, external_phone)
);
CREATE INDEX idx_integration_client_map_client ON integration_client_map(client_id);

-- +goose Down
DROP TABLE IF EXISTS integration_client_map;
DROP TABLE IF EXISTS integration_aggregates;
