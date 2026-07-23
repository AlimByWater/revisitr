-- +goose Up
-- +goose StatementBegin

-- dev previously used version 48 for lunch while main used it for iiko menu
-- overrides. Reconcile both schemas under a unique timestamp version.
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

ALTER TABLE external_orders
    ADD COLUMN IF NOT EXISTS source VARCHAR(20) NOT NULL DEFAULT 'delivery',
    ADD COLUMN IF NOT EXISTS table_num VARCHAR(50),
    ADD COLUMN IF NOT EXISTS waiter_name VARCHAR(255);

UPDATE external_orders SET source = 'delivery' WHERE source = 'cloud';

CREATE INDEX IF NOT EXISTS idx_external_orders_source
    ON external_orders(integration_id, source, ordered_at DESC);

CREATE TABLE IF NOT EXISTS lunch_programs (
    id          SERIAL PRIMARY KEY,
    bot_id      INT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL DEFAULT 'Бизнес-ланч',
    description TEXT NOT NULL DEFAULT '',
    is_active   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (bot_id)
);

CREATE TABLE IF NOT EXISTS lunch_courses (
    id               SERIAL PRIMARY KEY,
    program_id       INT NOT NULL REFERENCES lunch_programs(id) ON DELETE CASCADE,
    code             VARCHAR(8) NOT NULL,
    title            VARCHAR(255) NOT NULL,
    menu_category_id INT NOT NULL REFERENCES menu_categories(id) ON DELETE CASCADE,
    sort_order       INT NOT NULL DEFAULT 0,
    UNIQUE (program_id, code)
);

CREATE INDEX IF NOT EXISTS idx_lunch_courses_program ON lunch_courses(program_id);

CREATE TABLE IF NOT EXISTS lunch_course_items (
    course_id    INT NOT NULL REFERENCES lunch_courses(id) ON DELETE CASCADE,
    menu_item_id INT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    surcharge    NUMERIC(10,2) NOT NULL DEFAULT 0,
    PRIMARY KEY (course_id, menu_item_id)
);

CREATE TABLE IF NOT EXISTS lunch_formats (
    id         SERIAL PRIMARY KEY,
    program_id INT NOT NULL REFERENCES lunch_programs(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    price_mode VARCHAR(32) NOT NULL DEFAULT 'fixed'
               CHECK (price_mode IN ('fixed', 'sum_of_items', 'base_plus_surcharge')),
    base_price NUMERIC(10,2) NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_lunch_formats_program ON lunch_formats(program_id);

CREATE TABLE IF NOT EXISTS lunch_format_courses (
    format_id INT NOT NULL REFERENCES lunch_formats(id) ON DELETE CASCADE,
    course_id INT NOT NULL REFERENCES lunch_courses(id) ON DELETE CASCADE,
    position  INT NOT NULL DEFAULT 0,
    PRIMARY KEY (format_id, course_id)
);

CREATE TABLE IF NOT EXISTS lunch_availability (
    id         SERIAL PRIMARY KEY,
    program_id INT NOT NULL REFERENCES lunch_programs(id) ON DELETE CASCADE,
    weekday    SMALLINT NOT NULL CHECK (weekday BETWEEN 1 AND 7),
    time_from  TIME NOT NULL,
    time_to    TIME NOT NULL,
    CHECK (time_from < time_to)
);

CREATE INDEX IF NOT EXISTS idx_lunch_availability_program ON lunch_availability(program_id);

CREATE TABLE IF NOT EXISTS lunch_orders (
    id            SERIAL PRIMARY KEY,
    bot_id        INT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    bot_client_id INT NOT NULL REFERENCES bot_clients(id) ON DELETE CASCADE,
    format_id     INT REFERENCES lunch_formats(id) ON DELETE SET NULL,
    format_name   VARCHAR(255) NOT NULL DEFAULT '',
    table_num     VARCHAR(16) NOT NULL,
    total_price   NUMERIC(10,2) NOT NULL,
    status        VARCHAR(16) NOT NULL DEFAULT 'new'
                  CHECK (status IN ('new', 'sent', 'cancelled')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_lunch_orders_bot_status ON lunch_orders(bot_id, status);

CREATE TABLE IF NOT EXISTS lunch_order_items (
    id             SERIAL PRIMARY KEY,
    lunch_order_id INT NOT NULL REFERENCES lunch_orders(id) ON DELETE CASCADE,
    course_id      INT REFERENCES lunch_courses(id) ON DELETE SET NULL,
    course_title   VARCHAR(255) NOT NULL DEFAULT '',
    menu_item_id   INT REFERENCES menu_items(id) ON DELETE SET NULL,
    item_name      VARCHAR(255) NOT NULL DEFAULT '',
    price          NUMERIC(10,2) NOT NULL DEFAULT 0,
    surcharge      NUMERIC(10,2) NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_lunch_order_items_order ON lunch_order_items(lunch_order_id);

-- +goose StatementEnd

-- +goose Down
-- Repair migration intentionally has no destructive rollback.
SELECT 1;
