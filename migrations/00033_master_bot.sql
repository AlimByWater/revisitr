-- +goose Up

-- Managed bot fields on bots table
ALTER TABLE bots ADD COLUMN is_managed BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE bots ADD COLUMN managed_bot_id BIGINT;
ALTER TABLE bots ADD COLUMN created_by_telegram_id BIGINT;

-- Master bot links: ties org to Telegram account for bot management
CREATE TABLE master_bot_links (
    id SERIAL PRIMARY KEY,
    org_id INTEGER NOT NULL REFERENCES organizations(id),
    telegram_user_id BIGINT NOT NULL,
    telegram_username VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, telegram_user_id)
);

CREATE INDEX idx_master_bot_links_telegram_user_id ON master_bot_links(telegram_user_id);

-- +goose Down

DROP INDEX IF EXISTS idx_master_bot_links_telegram_user_id;
DROP TABLE IF EXISTS master_bot_links;

ALTER TABLE bots DROP COLUMN IF EXISTS created_by_telegram_id;
ALTER TABLE bots DROP COLUMN IF EXISTS managed_bot_id;
ALTER TABLE bots DROP COLUMN IF EXISTS is_managed;
