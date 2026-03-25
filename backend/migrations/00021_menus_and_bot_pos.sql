-- +goose Up

-- Меню, импортированное из POS или созданное вручную
CREATE TABLE menus (
    id             SERIAL PRIMARY KEY,
    org_id         INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    integration_id INT REFERENCES integrations(id) ON DELETE SET NULL,
    name           VARCHAR(255) NOT NULL,
    source         VARCHAR(20) NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'pos_import')),
    last_synced_at TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_menus_org_id ON menus(org_id);

-- Категории меню
CREATE TABLE menu_categories (
    id         SERIAL PRIMARY KEY,
    menu_id    INT NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_menu_categories_menu_id ON menu_categories(menu_id);

-- Позиции меню
CREATE TABLE menu_items (
    id           SERIAL PRIMARY KEY,
    category_id  INT NOT NULL REFERENCES menu_categories(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    description  TEXT,
    price        NUMERIC(12,2) NOT NULL DEFAULT 0,
    image_url    VARCHAR(500),
    tags         JSONB NOT NULL DEFAULT '[]',
    external_id  VARCHAR(255),
    is_available BOOLEAN NOT NULL DEFAULT true,
    sort_order   INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_menu_items_category_id ON menu_items(category_id);
CREATE INDEX idx_menu_items_external_id ON menu_items(external_id);

-- Привязка бота к POS-локациям (many-to-many)
CREATE TABLE bot_pos_locations (
    bot_id INT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    pos_id INT NOT NULL REFERENCES pos_locations(id) ON DELETE CASCADE,
    PRIMARY KEY (bot_id, pos_id)
);

-- Добавить телефон и имя в external_orders для матчинга
ALTER TABLE external_orders
    ADD COLUMN IF NOT EXISTS customer_phone VARCHAR(50),
    ADD COLUMN IF NOT EXISTS customer_name  VARCHAR(255);

CREATE INDEX idx_external_orders_customer_phone ON external_orders(customer_phone);

-- +goose Down
DROP INDEX IF EXISTS idx_external_orders_customer_phone;
ALTER TABLE external_orders
    DROP COLUMN IF EXISTS customer_phone,
    DROP COLUMN IF EXISTS customer_name;
DROP TABLE IF EXISTS bot_pos_locations;
DROP TABLE IF EXISTS menu_items;
DROP TABLE IF EXISTS menu_categories;
DROP TABLE IF EXISTS menus;
