-- +goose Up
CREATE TABLE emoji_packs (
    id          SERIAL PRIMARY KEY,
    org_id      INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_emoji_packs_org_id ON emoji_packs(org_id);

CREATE TABLE emoji_items (
    id              SERIAL PRIMARY KEY,
    pack_id         INT NOT NULL REFERENCES emoji_packs(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    image_url       TEXT NOT NULL,
    sort_order      INT NOT NULL DEFAULT 0,
    tg_sticker_set  VARCHAR(255),
    tg_custom_emoji_id VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_emoji_items_pack_id ON emoji_items(pack_id);

-- +goose Down
DROP TABLE IF EXISTS emoji_items;
DROP TABLE IF EXISTS emoji_packs;
