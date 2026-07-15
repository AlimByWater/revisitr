-- +goose Up
-- +goose StatementBegin

CREATE TABLE pos_plugin_keys (
    id             SERIAL PRIMARY KEY,
    org_id         INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    integration_id INT NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    key_hash       VARCHAR(64) NOT NULL UNIQUE,
    label          VARCHAR(255) NOT NULL DEFAULT '',
    last_used_at   TIMESTAMPTZ,
    revoked_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_pos_plugin_keys_integration_id ON pos_plugin_keys(integration_id);

CREATE TABLE pos_plugin_operations (
    id                SERIAL PRIMARY KEY,
    integration_id    INT NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    external_order_id VARCHAR(255) NOT NULL,
    op_type           VARCHAR(20) NOT NULL CHECK (op_type IN ('redeem','accrue')),
    client_id         INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    program_id        INT NOT NULL REFERENCES loyalty_programs(id) ON DELETE CASCADE,
    amount            DECIMAL(12,2) NOT NULL,
    balance_after     DECIMAL(12,2) NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (integration_id, external_order_id, op_type)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS pos_plugin_operations;
DROP TABLE IF EXISTS pos_plugin_keys;

-- +goose StatementEnd
