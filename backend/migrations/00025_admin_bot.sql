-- +goose Up

-- Admin bot account linking (Telegram → web panel user)
CREATE TABLE IF NOT EXISTS admin_bot_links (
    id                   SERIAL PRIMARY KEY,
    user_id              INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    telegram_id          BIGINT  UNIQUE,
    org_id               INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role                 TEXT    NOT NULL DEFAULT 'owner',  -- 'owner' | 'manager'
    linked_at            TIMESTAMPTZ,
    link_code            TEXT,
    link_code_expires_at TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_admin_bot_links_user_id ON admin_bot_links(user_id);
CREATE INDEX IF NOT EXISTS idx_admin_bot_links_telegram_id ON admin_bot_links(telegram_id);
CREATE INDEX IF NOT EXISTS idx_admin_bot_links_link_code ON admin_bot_links(link_code) WHERE link_code IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS admin_bot_links;
