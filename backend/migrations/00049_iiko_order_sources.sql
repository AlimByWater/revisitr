-- +goose Up

ALTER TABLE external_orders
    ADD COLUMN IF NOT EXISTS source VARCHAR(20) NOT NULL DEFAULT 'delivery',
    ADD COLUMN IF NOT EXISTS table_num VARCHAR(50),
    ADD COLUMN IF NOT EXISTS waiter_name VARCHAR(255);

ALTER TABLE external_orders
    DROP CONSTRAINT IF EXISTS external_orders_integration_id_external_id_key;

ALTER TABLE external_orders
    ADD CONSTRAINT external_orders_integration_id_external_id_key
    UNIQUE (integration_id, external_id);

UPDATE external_orders SET source = 'delivery' WHERE source = 'cloud';

CREATE INDEX IF NOT EXISTS idx_external_orders_source
    ON external_orders(integration_id, source, ordered_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_external_orders_source;

ALTER TABLE external_orders
    DROP COLUMN IF EXISTS waiter_name,
    DROP COLUMN IF EXISTS table_num,
    DROP COLUMN IF EXISTS source;
