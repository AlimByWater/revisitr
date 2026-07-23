-- +goose Up

ALTER TABLE menu_categories
    ADD COLUMN IF NOT EXISTS pos_external_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS pos_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS display_name VARCHAR(255);

ALTER TABLE menu_items
    ADD COLUMN IF NOT EXISTS pos_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS pos_description TEXT,
    ADD COLUMN IF NOT EXISTS pos_image_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS pos_category_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS display_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS missing_in_pos BOOLEAN NOT NULL DEFAULT false;

UPDATE menu_categories
SET pos_name = name
WHERE pos_name IS NULL;

UPDATE menu_items
SET pos_name = name
WHERE external_id IS NOT NULL AND pos_name IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_menu_categories_pos_external_id
    ON menu_categories(menu_id, pos_external_id)
    WHERE pos_external_id IS NOT NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_menu_categories_pos_external_id;

ALTER TABLE menu_items
    DROP COLUMN IF EXISTS missing_in_pos,
    DROP COLUMN IF EXISTS display_name,
    DROP COLUMN IF EXISTS pos_category_name,
    DROP COLUMN IF EXISTS pos_image_url,
    DROP COLUMN IF EXISTS pos_description,
    DROP COLUMN IF EXISTS pos_name;

ALTER TABLE menu_categories
    DROP COLUMN IF EXISTS display_name,
    DROP COLUMN IF EXISTS pos_name,
    DROP COLUMN IF EXISTS pos_external_id;
