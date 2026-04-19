-- +goose Up

ALTER TABLE menus
    ADD COLUMN IF NOT EXISTS intro_content JSONB;

ALTER TABLE menu_categories
    ADD COLUMN IF NOT EXISTS icon_emoji VARCHAR(32),
    ADD COLUMN IF NOT EXISTS icon_image_url VARCHAR(500);

ALTER TABLE menu_items
    ADD COLUMN IF NOT EXISTS weight VARCHAR(64);

CREATE TABLE IF NOT EXISTS menu_pos_bindings (
    menu_id INT NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    pos_id INT NOT NULL REFERENCES pos_locations(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (menu_id, pos_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_menu_pos_bindings_active_pos
    ON menu_pos_bindings(pos_id)
    WHERE is_active = true;

-- +goose Down

DROP INDEX IF EXISTS idx_menu_pos_bindings_active_pos;
DROP TABLE IF EXISTS menu_pos_bindings;

ALTER TABLE menu_items
    DROP COLUMN IF EXISTS weight;

ALTER TABLE menu_categories
    DROP COLUMN IF EXISTS icon_emoji,
    DROP COLUMN IF EXISTS icon_image_url;

ALTER TABLE menus
    DROP COLUMN IF EXISTS intro_content;
