-- +goose Up

CREATE TABLE post_codes (
    id SERIAL PRIMARY KEY,
    org_id INTEGER NOT NULL REFERENCES organizations(id),
    code VARCHAR(10) NOT NULL,
    content JSONB NOT NULL,
    telegram_message_ids JSONB,
    created_by_telegram_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, code)
);

CREATE INDEX idx_post_codes_org_id ON post_codes(org_id);

ALTER TABLE campaigns ADD COLUMN post_code_id INTEGER REFERENCES post_codes(id);

-- +goose Down

ALTER TABLE campaigns DROP COLUMN IF EXISTS post_code_id;
DROP INDEX IF EXISTS idx_post_codes_org_id;
DROP TABLE IF EXISTS post_codes;
